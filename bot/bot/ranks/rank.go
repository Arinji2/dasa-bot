package rank

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	bot_utils "github.com/arinji2/dasa-bot/bot/utils"
	commands_utils "github.com/arinji2/dasa-bot/commands"
	"github.com/arinji2/dasa-bot/env"
	"github.com/arinji2/dasa-bot/pb"
	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
)

type RankCommand struct {
	RankData    []pb.RankCollection
	CollegeData []pb.CollegeCollection
	PbAdmin     pb.PocketbaseAdmin
	BotEnv      env.Bot
}

func (r *RankCommand) branchesForCollege(collegeID string, ciwg bool) []pb.BranchCollection {
	branches := []pb.BranchCollection{}
	for _, v := range r.RankData {
		if v.College == collegeID && v.Expand.Branch.Ciwg == ciwg {
			branches = append(branches, v.Expand.Branch)
		}
	}
	return branches
}

func (r *RankCommand) HandleCutoffResponse(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionMessageComponent {
		bot_utils.HandleSendToDMButton(s, i)
		return
	}

	data := i.ApplicationCommandData()
	if len(data.Options) == 5 {
		r.handleCutoffWithBranch(s, i, data)
	}
}

func (r *RankCommand) HandleRankAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	var choices []*discordgo.ApplicationCommandOptionChoice
	spew.Dump(data.Options)

	switch {
	case data.Options[0].Focused:
		searchTerm := strings.ToLower(data.Options[0].StringValue())
		count := 0

		for _, v := range r.CollegeData {
			if count >= 25 {
				break
			}
			if searchTerm == "" || strings.Contains(strings.ToLower(v.Alias), searchTerm) ||
				strings.Contains(strings.ToLower(v.Name), searchTerm) {
				name := v.Name
				if len(name) > 100 {
					name = name[:97] + "..."
				}
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  name,
					Value: v.ID,
				})
				count++
			}
		}

	case data.Options[1].Focused:
		searchTerm := strings.ToLower(data.Options[1].StringValue())
		count := 0

		yearSet := make(map[int]struct{})
		for _, v := range r.RankData {
			yearSet[v.Year] = struct{}{}
		}

		var years []int
		for y := range yearSet {
			years = append(years, y)
		}
		sort.Ints(years)

		for _, v := range years {
			stringYear := strconv.Itoa(v)
			if count >= 25 {
				break
			}
			if searchTerm == "" || strings.Contains(stringYear, searchTerm) {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  stringYear,
					Value: stringYear,
				})
				count++
			}
		}

	case data.Options[2].Focused:
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  "CIWG",
			Value: "true",
		})
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  "Non-CIWG",
			Value: "false",
		})

	case data.Options[3].Focused:
		searchTerm := strings.ToLower(data.Options[3].StringValue())
		count := 0

		rounds := []int{1, 2, 3}
		for _, v := range rounds {
			stringRound := strconv.Itoa(v)
			if count >= 25 {
				break
			}
			if searchTerm == "" || strings.Contains(stringRound, searchTerm) {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  stringRound,
					Value: stringRound,
				})
				count++
			}
		}

	case data.Options[4].Focused:
		collegeID := strings.ToLower(data.Options[0].StringValue())
		branch := strings.ToLower(data.Options[2].StringValue())

		branchBool := false
		if branch == "true" {
			branchBool = true
		}
		searchTerm := strings.ToLower(data.Options[4].StringValue())
		count := 0

		branches := r.branchesForCollege(collegeID, branchBool)

		for _, b := range branches {
			if count >= 25 {
				break
			}
			if searchTerm == "" || strings.Contains(strings.ToLower(b.Code), searchTerm) || strings.Contains(strings.ToLower(b.Name), searchTerm) {
				name := b.Name
				if len(name) > 80 {
					name = name[:77] + "..."
				}
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  fmt.Sprintf("%s (%s)", b.Code, name),
					Value: b.Code,
				})
				count++
			}
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		log.Printf("Error sending autocomplete response: %v", err)
	}
}

func (r *RankCommand) handleCutoffWithBranch(s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	collegeID := data.Options[0].StringValue()
	year := data.Options[1].StringValue()
	ciwg := data.Options[2].StringValue()
	round := data.Options[3].StringValue()
	branch := data.Options[4].StringValue()

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		log.Printf("Error converting year to int: %v", err)
		return
	}
	roundInt, err := strconv.Atoi(round)
	if err != nil {
		log.Printf("Error converting round to int: %v", err)
		return
	}
	ciwgBool := (ciwg == "true")

	collegeData, err := r.PbAdmin.GetCollegeByID(collegeID)
	if err != nil {
		log.Printf("Error fetching college data: %v", err)
		return
	}
	rankData, err := r.PbAdmin.GetSpecificRank(collegeID, branch, yearInt, roundInt, ciwgBool)
	if err != nil {
		log.Printf("Error fetching rank data: %v", err)
		return
	}

	description := ""
	if rankData.Expand.Branch.Ciwg {
		description += fmt.Sprintf("Course %s (CIWG)\n", rankData.Expand.Branch.Name)
	} else {
		description += fmt.Sprintf("Course %s\n", rankData.Expand.Branch.Name)
	}
	description += fmt.Sprintf("Branch Code: %s\nRound %d (%d)", rankData.Expand.Branch.Code, roundInt, yearInt)

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Send To DM",
					Style:    discordgo.PrimaryButton,
					CustomID: "college_send_dm",
				},
			},
		},
	}

	title := fmt.Sprintf("Cutoffs for %s", collegeData.Name)

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "JEE Opening Rank",
			Value:  "22121",
			Inline: true,
		},
		{
			Name:   "JEE Closing Rank",
			Value:  "119645",
			Inline: true,
		},
		{
			Name:   "CIWG Opening Rank",
			Value:  "87",
			Inline: true,
		},
		{
			Name:   "CIWG Closing Rank",
			Value:  "375",
			Inline: true,
		},
	}

	err = commands_utils.RespondWithEmbedAndComponents(s, i, r.BotEnv, title, description, fields, components)
	if err != nil {
		log.Printf("Error sending college response: %v", err)
	}
}
