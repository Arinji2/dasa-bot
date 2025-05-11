package college

import (
	"fmt"
	"log"
	"strings"

	commands_utils "github.com/arinji2/dasa-bot/commands"
	"github.com/arinji2/dasa-bot/pb"
	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
)

type CollegeCommand struct {
	CollegeData []pb.CollegeCollection
	PbAdmin     pb.PocketbaseAdmin
}

func (c *CollegeCommand) HandleCollegeResponse(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	fmt.Println(data.Options)
	switch len(data.Options) {
	case 0:
		c.handleAllColleges(s, i)
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
	spew.Dump(choices)

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

func (c *CollegeCommand) handleAllColleges(s *discordgo.Session, i *discordgo.InteractionCreate) {
	description := "Showing **All Colleges** \n"
	if len(c.CollegeData) == 0 {
		commands_utils.RespondWithEphemeralError(s, i, "No articles found ")
		return
	}

	for i, v := range c.CollegeData {
		if i > 20 {
			break
		}
		description += fmt.Sprintf(
			"%d. **Name: ** %s \n\n **Alias: ** %s\n\n", i+1, v.Name, v.Alias)
	}

	fmt.Println(description)

	commands_utils.RespondWithEmbed(s, i, "Dasa College Details", description)
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
	err = commands_utils.RespondWithEmbed(s, i, "DASA CollegeDetails", description)
	if err != nil {
		log.Printf("Error sending college response: %v", err)
	}
}
