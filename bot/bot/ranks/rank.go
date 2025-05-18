package rank

import (
	"log"
	"sort"
	"strconv"
	"strings"

	bot_utils "github.com/arinji2/dasa-bot/bot/utils"
	"github.com/arinji2/dasa-bot/env"
	"github.com/arinji2/dasa-bot/pb"
	"github.com/bwmarrin/discordgo"
)

// Discord has a limit of 25 per list select, so we handle that with this constant
const maxSelectOptions = 25

type RankCommand struct {
	RankData    []pb.RankCollection
	CollegeData []pb.CollegeCollection
	PbAdmin     pb.PocketbaseAdmin
	BotEnv      env.Bot
}

func (r *RankCommand) HandleCutoffResponse(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionMessageComponent {
		customID := i.MessageComponentData().CustomID

		if customID == "college_send_dm" {
			bot_utils.HandleSendToDMButton(s, i)
			return
		}

		if strings.HasPrefix(customID, "select_branch_") {
			r.handleBranchSelection(s, i)
			return
		}
	}

	data := i.ApplicationCommandData()
	if len(data.Options) == 4 {
		r.showBranchSelect(s, i, data)
	}

	if len(data.Options) == 5 {
		r.handleCollegeBranches(s, i, data)
	}
}

func (r *RankCommand) HandleRankAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	var choices []*discordgo.ApplicationCommandOptionChoice

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
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  "Yes",
			Value: "true",
		})
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
