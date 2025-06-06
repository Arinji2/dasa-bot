package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	PbAdmin     *pb.PocketbaseAdmin
	RankCommand rank.RankCommand
	ModRole     []string
)

func NewBot(bot env.Bot) (*Bot, error) {
	var err error
	s, err := discordgo.New("Bot " + bot.Token)
	if err != nil {
		log.Fatalf("Invalid token: %v", err)
	}
	ModRole = bot.ModRole
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
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"refresh-data": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			checkChannel(s, i)
			checkPermissions(s, i)
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
			checkChannel(s, i)
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				log.Println("Handling Cutoff Response")
				RankCommand.HandleRankCutoffResponse(s, i)
			case discordgo.InteractionApplicationCommandAutocomplete:
				log.Println("Handling Cutoff Autocomplete")
				RankCommand.HandleRankCutoffAutocomplete(s, i)
			}
		},

		"analyze": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			checkChannel(s, i)
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				log.Println("Handling Analyze Response")
				RankCommand.HandleAnalyzeResponse(s, i)
			case discordgo.InteractionApplicationCommandAutocomplete:
				log.Println("Handling Analyze Autocomplete")
				RankCommand.HandleAnalyzeAutocomplete(s, i)
			}
		},
	}
)
