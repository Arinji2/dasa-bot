package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arinji2/dasa-bot/bot/college"
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
	PbAdmin        *pb.PocketbaseAdmin
	CollegeCommand college.CollegeCommand
	RankCommand    rank.RankCommand
	ModRole        []string
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
			Name:        "get-colleges",
			Description: "Get the colleges from Rounds",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "college-name",
					Description:  "College Name/Alias (Optional)",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     false,
					Autocomplete: true,
				},
			},
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
		"get-colleges": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			checkChannel(s, i)
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				CollegeCommand.HandleCollegeResponse(s, i)
			case discordgo.InteractionApplicationCommandAutocomplete:
				CollegeCommand.HandleCollegeAutocomplete(s, i)
			}
		},

		"cutoff": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			checkChannel(s, i)
			fmt.Println("HandleCutoffResponse", i.Type)
			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				RankCommand.HandleCutoffResponse(s, i)
			case discordgo.InteractionApplicationCommandAutocomplete:
				RankCommand.HandleRankAutocomplete(s, i)
			}
		},
	}
)
