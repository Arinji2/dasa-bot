// Package responses contains utility functions for responding to commands and interactions
package responses

import (
	"github.com/arinji2/dasa-bot/env"
	"github.com/bwmarrin/discordgo"
)

const (
	botFooter  = "Dasa Bot \nMade by Arinji :D"
	embedColor = 0xA000FF
)

func CreateBaseEmbed(title string, description string, botEnv env.Bot, fields []*discordgo.MessageEmbedField) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       embedColor,
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: botFooter,
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: botEnv.Thumbnail,
		},
	}
}

func RespondWithEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, botEnv env.Bot, title, description string, fields []*discordgo.MessageEmbedField) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				CreateBaseEmbed(title, description, botEnv, fields),
			},
		},
	})
}

func RespondWithEphermalEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, botEnv env.Bot, title, description string, fields []*discordgo.MessageEmbedField) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				CreateBaseEmbed(title, description, botEnv, fields),
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func RespondWithEmbedAndComponents(s *discordgo.Session, i *discordgo.InteractionCreate, botEnv env.Bot, title, description string, fields []*discordgo.MessageEmbedField, components []discordgo.MessageComponent) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				CreateBaseEmbed(title, description, botEnv, fields),
			},
			Components: components,
		},
	})
}

func RespondWithEphemeralEmbedAndComponents(s *discordgo.Session, i *discordgo.InteractionCreate, botEnv env.Bot, title, description string, fields []*discordgo.MessageEmbedField, components []discordgo.MessageComponent) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				CreateBaseEmbed(title, description, botEnv, fields),
			},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

func RespondWithAutoEmbedAndComponents(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	botEnv env.Bot,
	title, description string,
	fields []*discordgo.MessageEmbedField,
	components []discordgo.MessageComponent,
	BotChannel string,
) error {
	isEphemeral := i.ChannelID != BotChannel

	data := &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			CreateBaseEmbed(title, description, botEnv, fields),
		},
		Components: components,
	}

	if isEphemeral {
		data.Flags = discordgo.MessageFlagsEphemeral
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})
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
