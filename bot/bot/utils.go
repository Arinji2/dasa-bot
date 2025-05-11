package bot

import (
	"log"
	"slices"

	"github.com/arinji2/dasa-bot/pb"
	"github.com/bwmarrin/discordgo"
)

func isAdmin(i *discordgo.InteractionCreate) bool {
	for _, role := range i.Member.Roles {
		if slices.Contains(ModRole, role) {
			return true
		}
	}
	return false
}

func checkPermissions(s *discordgo.Session, i *discordgo.InteractionCreate) {
	hasPermission := isAdmin(i)
	if !hasPermission {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You do not have permission to use this command.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}

func checkChannel(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	isAdmin := isAdmin(i)
	if isAdmin {
		return true
	}

	// TODO: Incorporate Actual Allowed Channels
	// hasPermission := slices.Contains(AllowedChannels, i.ChannelID)
	// if !hasPermission {
	if false {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You cannot use this command in this channel.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return false
	}
	return true
}

func refreshData() {
	log.Println("Refreshing data...")

	locCollegeData, err := PbAdmin.GetAllColleges()
	if err != nil {
		log.Panicf("Cannot get articles: %v", err)
		locCollegeData = make([]pb.CollegeCollection, 0)
	}
	log.Printf("Found %d colleges", len(locCollegeData))

	CollegeCommand.CollegeData = locCollegeData
	CollegeCommand.PbAdmin = *PbAdmin
}

func (b *Bot) registerCommands() []*discordgo.ApplicationCommand {
	b.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
	err := b.Session.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")

	createdCommands, err := b.Session.ApplicationCommandBulkOverwrite(b.Session.State.User.ID, b.GuildID, commands)
	if err != nil {
		log.Panicf("Cannot create commands: %v", err)
	}
	return createdCommands
}

func (b *Bot) unregisterCommands() {
	log.Println("Removing commands...")

	for _, v := range b.Commands {
		err := b.Session.ApplicationCommandDelete(b.Session.State.User.ID, b.GuildID, v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}
}
