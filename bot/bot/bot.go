// Package bot contains the main bot logic
package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arinji2/dasa-bot/bot/insert"
	rank "github.com/arinji2/dasa-bot/bot/ranks"
	"github.com/arinji2/dasa-bot/env"
	"github.com/arinji2/dasa-bot/pb"
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Session  *discordgo.Session
	GuildID  string
	Commands []*discordgo.ApplicationCommand
	BotEnv   env.Bot
}

var (
	PbAdmin       *pb.PocketbaseAdmin
	RankCommand   rank.RankCommand
	InsertCommand insert.InsertCommand
	ModRole       []string
	BotChannel    string
	AdminChannel  string
)

func NewBot(bot env.Bot) (*Bot, error) {
	var err error
	s, err := discordgo.New("Bot " + bot.Token)
	if err != nil {
		log.Fatalf("Invalid token: %v", err)
	}
	ModRole = bot.ModRole
	BotChannel = bot.BotChannel
	AdminChannel = bot.AdminChannel
	return &Bot{Session: s, GuildID: bot.GuildID, BotEnv: bot}, nil
}

func (b *Bot) Run(pbAdmin *pb.PocketbaseAdmin) {
	log.Println("Starting bot...")
	b.Session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	PbAdmin = pbAdmin
	refreshData(&b.BotEnv)
	createdCommands := b.registerCommands()
	b.Session.UpdateCustomStatus("Padhlo chahe kahi se, selection hoga dasa se")
	b.Commands = createdCommands

	b.Session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		for _, user := range m.Mentions {
			if user.ID == s.State.User.ID {
				msgRef := &discordgo.MessageReference{
					MessageID: m.ID,
					ChannelID: m.ChannelID,
					GuildID:   m.GuildID,
				}
				_, _ = s.ChannelMessageSendReply(m.ChannelID,
					fmt.Sprintf("Hey <@%s>! Use `/cutoff` or `/analyze` to get started.", m.Author.ID),
					msgRef,
				)
				break
			}
		}
	})
	log.Println("Bot is now running.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	log.Println("\nShutting down gracefully...")

	if err := b.Session.Close(); err != nil {
		log.Printf("Error closing Discord session: %v", err)
	} else {
		log.Println("Discord session closed successfully.")
	}

	b.unregisterCommands()
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "refresh-data",
			Description: "Refresh the data of the bot",
			Type:        discordgo.ChatApplicationCommand,
			Options:     []*discordgo.ApplicationCommandOption{},
		},
		{
			Name:        "cutoff",
			Description: "Displays the ranks of a specified college and branch based on the user provided-year and round",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "college",
					Description:  "College Name/Alias",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
				{
					Name:         "year",
					Description:  "Year",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
				{
					Name:         "ciwg",
					Description:  "Is a CIWG student",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
				{
					Name:         "round",
					Description:  "Round",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},

				{
					Name:         "all",
					Description:  "See All Branches for the specified College",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     false,
					Autocomplete: true,
				},
			},
		},

		{
			Name:        "analyze",
			Description: "Shows colleges and branches with closing ranks near your rank and given deviation",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "rank",
					Description:  "The JEE rank you want to analyze",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: false,
				},
				{
					Name:         "ciwg",
					Description:  "Is a CIWG student",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
				{
					Name:         "deviation",
					Description:  "Allowed deviation from the rank. Higher deviation means lower accuracy.",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     false,
					Autocomplete: true,
				},
			},
		},
		{
			Name:        "insert",
			Description: "Inserts rank data based on inserted PDF with year and round",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "file",
					Description:  "CSV file of ranks",
					Type:         discordgo.ApplicationCommandOptionAttachment,
					Required:     true,
					Autocomplete: false,
				},

				{
					Name:         "year",
					Description:  "Year of the ranks",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: false,
				},

				{
					Name:         "round",
					Description:  "Round of the ranks",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: false,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"refresh-data": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := checkPermissions(s, i)
			if err != nil {
				return
			}
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				timeStart := time.Now()
				refreshData(nil)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Data refreshed in %v.", time.Since(timeStart)),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
			}
		},

		"cutoff": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				RankCommand.HandleRankCutoffResponse(s, i)
			case discordgo.InteractionApplicationCommandAutocomplete:
				RankCommand.HandleRankCutoffAutocomplete(s, i)
			}
		},

		"analyze": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				RankCommand.HandleAnalyzeResponse(s, i)
			case discordgo.InteractionApplicationCommandAutocomplete:
				RankCommand.HandleAnalyzeAutocomplete(s, i)
			}
		},

		"insert": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				err := checkChannel(s, i, true)
				if err != nil {
					return
				}
				err = checkPermissions(s, i)
				if err != nil {
					return
				}
				refreshData(&InsertCommand.BotEnv)
				InsertCommand.HandleInsertResponse(s, i)
			}
		},
	}
)
