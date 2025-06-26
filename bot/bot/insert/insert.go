// Package insert contains the logic for the Insert command
package insert

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	commands_utils "github.com/arinji2/dasa-bot/commands"
	"github.com/arinji2/dasa-bot/env"
	"github.com/arinji2/dasa-bot/pb"
	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
)

type InsertCommand struct {
	RankData    []pb.RankCollection
	BranchData  []pb.BranchCollection
	CollegeData []pb.CollegeCollection
	PbAdmin     pb.PocketbaseAdmin
	BotEnv      env.Bot
}

func (c *InsertCommand) HandleInsertResponse(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	c.HandleInsertData(s, i, &data)
}

func (c *InsertCommand) HandleInsertData(s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) {
	year := data.Options[1].StringValue()

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		log.Printf("Error converting year to int: %v", err)
		commands_utils.RespondWithEphemeralError(s, i, "Invalid year format")
		return
	}

	round := data.Options[2].StringValue()

	roundInt, err := strconv.Atoi(round)
	if err != nil {
		log.Printf("Error converting round to int: %v", err)
		commands_utils.RespondWithEphemeralError(s, i, "Invalid round format")
		return
	}
	attachmentID := i.ApplicationCommandData().Options[0].Value.(string)
	attachmentURL := i.ApplicationCommandData().Resolved.Attachments[attachmentID].URL

	// get the file contents
	res, err := http.DefaultClient.Get(attachmentURL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	reader := bytes.NewReader(body)
	csvReader := csv.NewReader(reader)
	_, err = csvReader.Read()
	if err != nil {
		log.Fatalf("Error: Could not read header from file: %v", err)
	}

	parsedRanks, parsedErrs, err := c.parseRankingData(csvReader, yearInt, roundInt)
	if err != nil {
		log.Fatal(err)
	}

	if len(parsedErrs) > 0 {
		var description string
		if len(parsedErrs) > 10 {
			description += fmt.Sprintf("First 10 errors out of %d: \n", len(parsedErrs))
		}
		for i, err := range parsedErrs {
			if i > 10 {
				break
			}
			// add 1 since we remove the header
			description += fmt.Sprintf("Line Number: %d \n %s \n\n", (err.Line + 1), err.Message)
		}
		fmt.Println(description)

		commands_utils.RespondWithEphermalEmbed(s, i, c.BotEnv, "Error with parsing data", description, nil)
	}

	spew.Dump(parsedRanks)
}

type RankParseError struct {
	Line    int
	Record  []string
	Message string
}

func (c *InsertCommand) parseRankingData(reader *csv.Reader, year, round int) ([]pb.RankCollection, []RankParseError, error) {
	var ranks []pb.RankCollection
	var errors []RankParseError
	lineNumber := 0

	for {
		record, err := reader.Read()
		lineNumber++
		if err == io.EOF {
			break
		}
		if err != nil {
			errors = append(errors, RankParseError{
				Line:    lineNumber,
				Message: fmt.Sprintf("Error reading record: %v", err),
			})
			continue
		}

		if len(record) < 7 {
			errors = append(errors, RankParseError{
				Line:    lineNumber,
				Record:  record,
				Message: fmt.Sprintf("Malformed row with %d columns", len(record)),
			})
			continue
		}

		collegeName := record[0]
		if collegeName == "" {
			errors = append(errors, RankParseError{
				Line:    lineNumber,
				Record:  record,
				Message: "Empty 'college_name' value",
			})
			continue
		}

		branchCode := record[1]
		if branchCode == "" {
			errors = append(errors, RankParseError{
				Line:    lineNumber,
				Record:  record,
				Message: "Empty 'branch_code' value",
			})
			continue
		}

		branchName := record[2]
		if branchName == "" {
			errors = append(errors, RankParseError{
				Line:    lineNumber,
				Record:  record,
				Message: "Empty 'branch_name' value",
			})
			continue
		}

		isCiWg, err := strconv.ParseBool(record[3])
		if err != nil {
			errors = append(errors, RankParseError{
				Line:    lineNumber,
				Record:  record,
				Message: fmt.Sprintf("Invalid 'is_ciwg' value: %v", err),
			})
			continue
		}

		firstRank, err := strconv.Atoi(record[4])
		if err != nil {
			errors = append(errors, RankParseError{
				Line:    lineNumber,
				Record:  record,
				Message: fmt.Sprintf("Invalid 'first_rank' value: %v", err),
			})
			continue
		}

		lastRank, err := strconv.Atoi(record[5])
		if err != nil {
			errors = append(errors, RankParseError{
				Line:    lineNumber,
				Record:  record,
				Message: fmt.Sprintf("Invalid 'last_rank' value: %v", err),
			})
			continue
		}

		extraIDS := record[6]
		var branchID string
		var collegeID string

		if extraIDS != "" {
			if strings.Contains(extraIDS, ":") {
				parts := strings.Split(extraIDS, ":")
				if len(parts) == 2 {
					if after, ok := strings.CutPrefix(parts[0], "b-"); ok {
						branchID = after
					}
					if after, ok := strings.CutPrefix(parts[1], "c-"); ok {
						collegeID = after
					}
				}
			} else if after, ok := strings.CutPrefix(extraIDS, "c-"); ok {
				collegeID = after
			} else if after, ok := strings.CutPrefix(extraIDS, "b-"); ok {
				branchID = after
			} else {
				errors = append(errors, RankParseError{
					Line:    lineNumber,
					Record:  record,
					Message: fmt.Sprintf("Invalid 'extra_id' value: %v", extraIDS),
				})
			}
		}

		if branchID != "" {
			fmt.Println("[bot/bot/insert/insert.go:189] branchID = ", branchID)
		}
		if collegeID != "" {
			fmt.Println("[bot/bot/insert/insert.go:191] collegeID = ", collegeID)
		}

		var collegeData pb.CollegeCollection

		if collegeID != "" {
			index := slices.IndexFunc(c.CollegeData, func(v pb.CollegeCollection) bool {
				return v.ID == collegeID
			})
			if index == -1 {
				errors = append(errors, RankParseError{
					Line:    lineNumber,
					Record:  record,
					Message: fmt.Sprintf("Invalid 'college_id' value: %v", collegeID),
				})
				continue
			} else {
				collegeData = c.CollegeData[index]
			}
		} else {
			index := slices.IndexFunc(c.CollegeData, func(v pb.CollegeCollection) bool {
				cleanedCollegeName := strings.ReplaceAll(collegeName, ",", "")
				cleanedv2 := strings.ReplaceAll(v.Name, ",", "")

				return strings.EqualFold(cleanedCollegeName, cleanedv2)
			})
			if index == -1 {
				errors = append(errors, RankParseError{
					Line:    lineNumber,
					Record:  record,
					Message: fmt.Sprintf("College of name **%v** dosent exist. Try adding the college id with c-(collegeID) as a 6th argument", collegeName),
				})
				continue
			} else {
				collegeData = c.CollegeData[index]
			}
		}

		var branchData pb.BranchCollection
		if branchID != "" {
			index := slices.IndexFunc(c.BranchData, func(v pb.BranchCollection) bool {
				return branchID == v.ID
			})
			if index == -1 {
				errors = append(errors, RankParseError{
					Line:    lineNumber,
					Record:  record,
					Message: fmt.Sprintf("Invalid 'branch_id' value: %v for 'college id' %v", branchID, collegeData.ID),
				})
				continue
			} else {
				branchData = c.BranchData[index]
			}
		} else {
			index := slices.IndexFunc(c.BranchData, func(v pb.BranchCollection) bool {
				return v.Name == branchName && v.Code == branchCode && v.Ciwg == isCiWg
			})
			if index == -1 {
				fmt.Printf("Creating branch %v\n", branchName)
				branchData, err = c.PbAdmin.CreateBranch(pb.BranchCreateRequest{
					Name: branchName,
					Code: branchCode,
					Ciwg: isCiWg,
				})

				c.BranchData = append(c.BranchData, branchData)
				if err != nil {
					errors = append(errors, RankParseError{
						Line:    lineNumber,
						Record:  record,
						Message: fmt.Sprintf("Error creating branch: %v", err),
					})
					continue
				}
			} else {
				branchData = c.BranchData[index]
			}
		}

		rankCollection := pb.RankCollection{
			JEE_OPEN:  firstRank,
			JEE_CLOSE: lastRank,
			Year:      year,
			Round:     round,
			Expand: struct {
				College pb.CollegeCollection "json:\"college\""
				Branch  pb.BranchCollection  "json:\"branch\""
			}{
				College: collegeData,
				Branch:  branchData,
			},
		}

		ranks = append(ranks, rankCollection)
	}

	return ranks, errors, nil
}
