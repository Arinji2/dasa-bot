// Package insert contains the logic for the Insert command
package insert

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/arinji2/dasa-bot/env"
	"github.com/arinji2/dasa-bot/pb"
	responses "github.com/arinji2/dasa-bot/responses"
	"github.com/bwmarrin/discordgo"
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
	var logs []string
	year := data.Options[1].StringValue()

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		log.Printf("Error converting year to int: %v", err)
		responses.RespondWithEphemeralError(s, i, "Invalid year format")
		return
	}

	round := data.Options[2].StringValue()

	roundInt, err := strconv.Atoi(round)
	if err != nil {
		log.Printf("Error converting round to int: %v", err)
		responses.RespondWithEphemeralError(s, i, "Invalid round format")
		return
	}
	attachmentID := i.ApplicationCommandData().Options[0].Value.(string)
	attachmentURL := i.ApplicationCommandData().Resolved.Attachments[attachmentID].URL

	// get the file contents
	res, err := http.DefaultClient.Get(attachmentURL)
	if err != nil {
		responses.RespondWithEmbed(s, i, c.BotEnv, "Error getting file", err.Error(), nil)
		return
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		responses.RespondWithEmbed(s, i, c.BotEnv, "Error reading file", err.Error(), nil)
		return
	}

	reader := bytes.NewReader(body)
	csvReader := csv.NewReader(reader)
	_, err = csvReader.Read()
	if err != nil {
		responses.RespondWithEmbed(s, i, c.BotEnv, "Error: Could not read header from file", err.Error(), nil)
		return
	}

	userName := i.Member.User.Username
	backupList, err := c.PbAdmin.ListBackups()
	if err != nil {
		responses.RespondWithEmbed(s, i, c.BotEnv, "Error listing backup", err.Error(), nil)
		return
	}

	logs = append(logs, fmt.Sprintf("Found **%d** backups", len(backupList)))
	if len(backupList) > 3 {
		logs = append(logs, fmt.Sprintf("Reached limit of 3, deleting backup of key **%s**", backupList[0].Key))
		err = c.PbAdmin.DeleteBackup(backupList[0].Key)
		if err != nil {
			responses.RespondWithEmbed(s, i, c.BotEnv, "Error deleting backup", err.Error(), nil)
			return
		}
	}
	backupName, err := c.PbAdmin.CreateBackup(userName)
	if err != nil {
		responses.RespondWithEmbed(s, i, c.BotEnv, "Error creating backup", err.Error(), nil)
		return
	}

	logs = append(logs, fmt.Sprintf("Created backup with name **%s**", backupName))

	parsedRanks, parsedErrs, err := c.parseRankingData(csvReader, yearInt, roundInt)
	if err != nil {
		responses.RespondWithEmbed(s, i, c.BotEnv, "Error parsing data", err.Error(), nil)
		return
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

		responses.RespondWithEmbed(s, i, c.BotEnv, "Error with parsing data", description, nil)
		return
	}

	// Replace the existing rank creation section with this code

	logs = append(logs, fmt.Sprintf("Parsed **%d** ranks", len(parsedRanks)))

	var wg sync.WaitGroup
	type RankCreateError struct {
		Rank pb.RankCollection
		Err  error
	}

	errorChan := make(chan RankCreateError, len(parsedRanks))

	// Create a buffered channel to limit concurrent goroutines to 10
	semaphore := make(chan struct{}, 10)

	skipped := 0
	var skippedMutex sync.Mutex

	for _, rank := range parsedRanks {
		wg.Add(1)
		go func(rank pb.RankCollection) {
			defer wg.Done()

			// Acquire semaphore (blocks if 10 goroutines are already running)
			semaphore <- struct{}{}
			defer func() { <-semaphore }() // Release semaphore when done

			_, exists, err := c.PbAdmin.CreateRank(pb.RankCreateRequest{
				Year:       rank.Year,
				Round:      rank.Round,
				JEE_OPEN:   rank.JEE_OPEN,
				JEE_CLOSE:  rank.JEE_CLOSE,
				DASA_OPEN:  rank.DASA_OPEN,
				DASA_CLOSE: rank.DASA_CLOSE,
				College:    rank.College,
				Branch:     rank.Branch,
			}, rank.Expand.Branch.Ciwg)

			if exists {
				skippedMutex.Lock()
				skipped++
				skippedMutex.Unlock()
			}

			if err != nil && !exists {
				errorChan <- RankCreateError{
					Rank: rank,
					Err:  err,
				}
			}
		}(rank)
	}

	wg.Wait()
	close(errorChan)

	var createErrors []RankCreateError
	for err := range errorChan {
		createErrors = append(createErrors, err)
	}

	if len(createErrors) > 0 {
		var description string
		if len(createErrors) > 10 {
			description += fmt.Sprintf("First 10 errors out of %d: \n", len(createErrors))
		}
		for i, err := range createErrors {
			if i > 10 {
				break
			}
			description += fmt.Sprintf("Failed to create rank with Jee Open: %d and College Name %s \n %s \n\n", err.Rank.JEE_OPEN, err.Rank.Expand.College.Name, err.Err.Error())
		}
		responses.RespondWithEmbed(s, i, c.BotEnv, "Error with creating data", description, nil)
		return
	}

	logs = append(logs, fmt.Sprintf("Skipped **%d** ranks", skipped))
	logs = append(logs, fmt.Sprintf("Successfully created **%d** ranks", len(parsedRanks)-skipped))

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Logs",
			Value:  strings.Join(logs, "\n"),
			Inline: true,
		},
	}
	responses.RespondWithEmbed(s, i, c.BotEnv, "Successfully created ranks", fmt.Sprintf("Successfully inserted ranks for Year: %d and Round %d", yearInt, roundInt), fields)
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

	collegeIDMap := make(map[string]pb.CollegeCollection, len(c.CollegeData))
	collegeNameMap := make(map[string]pb.CollegeCollection, len(c.CollegeData))
	for _, college := range c.CollegeData {
		collegeIDMap[college.ID] = college
		normalizedName := strings.ToLower(strings.ReplaceAll(college.Name, ",", ""))
		collegeNameMap[normalizedName] = college
	}

	branchIDMap := make(map[string]pb.BranchCollection, len(c.BranchData))
	branchKeyMap := make(map[string]pb.BranchCollection, len(c.BranchData))
	for _, branch := range c.BranchData {
		branchIDMap[branch.ID] = branch
		key := fmt.Sprintf("%s-%s-%t", strings.ToLower(branch.Name), strings.ToLower(branch.Code), branch.Ciwg)
		branchKeyMap[key] = branch
	}

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

		var collegeData pb.CollegeCollection
		var found bool

		if collegeID != "" {
			collegeData, found = collegeIDMap[collegeID]
			if !found {
				errors = append(errors, RankParseError{
					Line:    lineNumber,
					Record:  record,
					Message: fmt.Sprintf("Invalid 'college_id' value: %v", collegeID),
				})
				continue
			}
		} else {
			normalizedName := strings.ToLower(strings.ReplaceAll(collegeName, ",", ""))
			collegeData, found = collegeNameMap[normalizedName]
			if !found {
				errors = append(errors, RankParseError{
					Line:    lineNumber,
					Record:  record,
					Message: fmt.Sprintf("College of name **%v** dosent exist. Try adding the college id with c-(collegeID) as a 6th argument", collegeName),
				})
				continue
			}
		}

		var branchData pb.BranchCollection
		if branchID != "" {
			branchData, found = branchIDMap[branchID]
			if !found {
				errors = append(errors, RankParseError{
					Line:    lineNumber,
					Record:  record,
					Message: fmt.Sprintf("Invalid 'branch_id' value: %v for 'college id' %v", branchID, collegeData.ID),
				})
				continue
			}
		} else {
			key := fmt.Sprintf("%s-%s-%t", strings.ToLower(branchName), strings.ToLower(branchCode), isCiWg)
			branchData, found = branchKeyMap[key]
			if !found {
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
				// add new branch to map
				branchIDMap[branchData.ID] = branchData
				branchKeyMap[key] = branchData
			}
		}

		rankCollection := pb.RankCollection{
			JEE_OPEN:  firstRank,
			JEE_CLOSE: lastRank,
			Year:      year,
			Round:     round,
			College:   collegeData.ID,
			Branch:    branchData.ID,
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
