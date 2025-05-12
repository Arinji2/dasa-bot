package college

import (
	"fmt"
	"log"
	"strings"

	commands_utils "github.com/arinji2/dasa-bot/commands"
	"github.com/arinji2/dasa-bot/env"
	"github.com/arinji2/dasa-bot/pb"
	"github.com/bwmarrin/discordgo"
)

const (
	collegesPerPage = 10 // Define how many colleges to show per page
)

type CollegeCommand struct {
	CollegeData []pb.CollegeCollection
	PbAdmin     pb.PocketbaseAdmin
	BotEnv      env.Bot
}

func (c *CollegeCommand) HandleCollegeResponse(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionMessageComponent {
		c.handlePaginationButtons(s, i)
		return
	}

	data := i.ApplicationCommandData()

	switch len(data.Options) {
	case 0:
		c.handleAllColleges(s, i, 0)
	case 1:
		c.handleSpecificColleges(s, i, data)
	}
}

func (c *CollegeCommand) HandleCollegeAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	var choices []*discordgo.ApplicationCommandOptionChoice

	switch {
	case data.Options[0].Focused:
		searchTerm := strings.ToLower(data.Options[0].StringValue())
		count := 0

		for _, v := range c.CollegeData {
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
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		log.Printf("Error sending college response: %v", err)
	}
}

func (c *CollegeCommand) handleAllColleges(s *discordgo.Session, i *discordgo.InteractionCreate, page int) {
	if len(c.CollegeData) == 0 {
		commands_utils.RespondWithEphemeralError(s, i, "No colleges found")
		return
	}

	totalPages := (len(c.CollegeData) + collegesPerPage - 1) / collegesPerPage

	if page < 0 {
		page = 0
	} else if page >= totalPages {
		page = totalPages - 1
	}

	startIdx := page * collegesPerPage
	endIdx := startIdx + collegesPerPage
	endIdx = min(endIdx, len(c.CollegeData))

	description := fmt.Sprintf("Showing **All Colleges** (Page %d/%d)\n\n", page+1, totalPages)

	for i, v := range c.CollegeData[startIdx:endIdx] {
		description += fmt.Sprintf(
			"%d. **Name: ** %s \n\n **Alias: ** %s\n\n", startIdx+i+1, v.Name, v.Alias)
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Previous",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("college_prev_%d", page),
					Disabled: page <= 0,
				},
				discordgo.Button{
					Label:    fmt.Sprintf("Page %d/%d", page+1, totalPages),
					Style:    discordgo.SecondaryButton,
					CustomID: "college_page_info",
					Disabled: true,
				},
				discordgo.Button{
					Label:    "Next",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("college_next_%d", page),
					Disabled: page >= totalPages-1,
				},
			},
		},
	}

	err := commands_utils.RespondWithEmbedAndComponents(s, i, c.BotEnv, "Dasa College Details", description, components)
	if err != nil {
		log.Printf("Error sending college response: %v", err)
	}
}

func (c *CollegeCommand) handlePaginationButtons(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	var targetPage int

	if strings.HasPrefix(customID, "college_prev_") {
		fmt.Sscanf(customID, "college_prev_%d", &targetPage)
		targetPage--
	} else if strings.HasPrefix(customID, "college_next_") {
		fmt.Sscanf(customID, "college_next_%d", &targetPage)
		targetPage++

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("Error deferring response: %v", err)
		return
	}

	totalPages := (len(c.CollegeData) + collegesPerPage - 1) / collegesPerPage

	if targetPage < 0 {
		targetPage = 0
	} else if targetPage >= totalPages {
		targetPage = totalPages - 1
	}

	startIdx := targetPage * collegesPerPage
	endIdx := startIdx + collegesPerPage
	endIdx = min(endIdx, len(c.CollegeData))

	description := fmt.Sprintf("Showing **All Colleges** (Page %d/%d)\n\n", targetPage+1, totalPages)

	for i, v := range c.CollegeData[startIdx:endIdx] {
		description += fmt.Sprintf(
			"%d. **Name: ** %s \n\n **Alias: ** %s\n\n", startIdx+i+1, v.Name, v.Alias)
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Previous",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("college_prev_%d", targetPage),
					Disabled: targetPage <= 0,
				},
				discordgo.Button{
					Label:    fmt.Sprintf("Page %d/%d", targetPage+1, totalPages),
					Style:    discordgo.SecondaryButton,
					CustomID: "college_page_info",
					Disabled: true,
				},
				discordgo.Button{
					Label:    "Next",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("college_next_%d", targetPage),
					Disabled: targetPage >= totalPages-1,
				},
			},
		},
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{commands_utils.CreateBaseEmbed("Dasa College Details", description, c.BotEnv)},
		Components: &components,
	})
	if err != nil {
		log.Printf("Error editing college response: %v", err)
	}
}

func (c *CollegeCommand) handleSpecificColleges(s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	collegeID := data.Options[0].StringValue()

	collegeData, err := c.PbAdmin.GetCollegeByID(collegeID)
	if err != nil {
		log.Printf("Error fetching specific article: %v", err)
		commands_utils.RespondWithEphemeralError(s, i, "Could not retrieve article data")
		return
	}

	description := fmt.Sprintf(
		"Showing **College: %s**\n", collegeData.Name)

	description += fmt.Sprintf(
		"**Alias:** %s\n\n ", collegeData.Alias)
	err = commands_utils.RespondWithEmbed(s, i, c.BotEnv, "DASA College Details", description)
	if err != nil {
		log.Printf("Error sending college response: %v", err)
	}
}
