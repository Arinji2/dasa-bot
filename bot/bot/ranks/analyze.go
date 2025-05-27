package rank

import (
	"fmt"
	"log"

	bot_utils "github.com/arinji2/dasa-bot/bot/utils"
	"github.com/bwmarrin/discordgo"
)

func (r *RankCommand) HandleAnalyzeResponse(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionMessageComponent {
		customID := i.MessageComponentData().CustomID

		if customID == "college_send_dm" {
			bot_utils.HandleSendToDMButton(s, i)
			return
		}

		// if strings.HasPrefix(customID, "select_analyze_branch") {
		// 	r.handleAnalyzeBranchSelection(s, i)
		// 	return
		// }
	}

	data := i.ApplicationCommandData()
	if len(data.Options) == 3 {
		r.showAnalyzeBranchSelect(s, i, data)
	}
}

func (r *RankCommand) HandleAnalyzeAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	var choices []*discordgo.ApplicationCommandOptionChoice

	switch {
	case data.Options[1].Focused:
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  "CIWG",
			Value: "true",
		})
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  "Non-CIWG",
			Value: "false",
		})

	case data.Options[2].Focused:
		for _, v := range Devations {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  fmt.Sprintf("%s%%", v),
				Value: v,
			})
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
