// Package buttons contains utilities for buttons type interactions
package buttons

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func HandleSendToDMButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Defer the interaction to avoid timeout
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("Error deferring interaction: %v", err)
		return
	}

	// Open a DM channel with the user
	userID := i.Member.User.ID
	if userID == "" && i.User != nil {
		userID = i.User.ID // for direct message interactions
	}

	dm, err := s.UserChannelCreate(userID)
	if err != nil {
		log.Printf("Failed to open DM: %v", err)
		return
	}

	// Forward the embed from the original message
	if i.Message != nil && len(i.Message.Embeds) > 0 {
		_, err = s.ChannelMessageSendEmbed(dm.ID, i.Message.Embeds[0])
		if err != nil {
			log.Printf("Failed to send DM embed: %v", err)
		}
	} else {
		// Fallback: send a generic message
		_, err = s.ChannelMessageSend(dm.ID, "Sorry, couldn't find anything to send.")
		if err != nil {
			log.Printf("Failed to send fallback DM: %v", err)
		}
	}
}
