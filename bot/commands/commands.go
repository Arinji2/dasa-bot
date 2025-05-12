package commands_utils

import (
	"github.com/arinji2/dasa-bot/env"
	"github.com/bwmarrin/discordgo"
)

const (
	botFooter  = "Dasa Bot"
	embedColor = 0xA000FF // Garconia Red
)

func CreateBaseEmbed(title, description string, botEnv env.Bot) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       embedColor,
		Footer: &discordgo.MessageEmbedFooter{
			Text: botFooter,
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: botEnv.Thumbnail,
		},
	}
}

func RespondWithEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, botEnv env.Bot, title, description string) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{CreateBaseEmbed(title, description, botEnv)},
		},
	})
	return err
}

func RespondWithEmbedAndComponents(s *discordgo.Session, i *discordgo.InteractionCreate, botEnv env.Bot, title, description string, components []discordgo.MessageComponent) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{CreateBaseEmbed(title, description, botEnv)},
			Components: components,
		},
	})
	return err
}

func RespondWithEphemeralError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
