package bot

import (
	"fmt"
	"log"
	"slices"
	"strings"

	buttons "github.com/arinji2/dasa-bot/bot/buttons"
	"github.com/arinji2/dasa-bot/env"
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

func checkPermissions(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	hasPermission := isAdmin(i)
	if !hasPermission {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You do not have permission to use this command.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return fmt.Errorf("you do not have permission to use this command")
	}
	return nil
}

func checkChannel(s *discordgo.Session, i *discordgo.InteractionCreate, isAdminCheck bool) error {
	var hasPermission bool
	if isAdminCheck {
		hasPermission = AdminChannel == i.ChannelID
	} else {
		hasPermission = BotChannel == i.ChannelID
	}

	if !hasPermission {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You cannot use this command in this channel.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return fmt.Errorf("you cannot use this command in this channel")
	}
	return nil
}

func refreshData(botEnv *env.Bot) {
	log.Println("Refreshing data...")

	locCollegeData, err := PbAdmin.GetAllColleges()
	if err != nil {
		log.Panicf("Cannot get colleges: %v", err)
		locCollegeData = make([]pb.CollegeCollection, 0)
	}
	log.Printf("Found %d colleges", len(locCollegeData))

	locRankData, err := PbAdmin.GetAllRanks()
	if err != nil {
		log.Panicf("Cannot get ranks: %v", err)
		locRankData = make([]pb.RankCollection, 0)
	}
	log.Printf("Found %d ranks", len(locRankData))

	locBranchData, err := PbAdmin.GetAllBranches()
	if err != nil {
		log.Panicf("Cannot get branches: %v", err)
		locBranchData = make([]pb.BranchCollection, 0)
	}
	log.Printf("Found %d branches", len(locBranchData))

	InsertCommand.BranchData = locBranchData

	RankCommand.CollegeData = locCollegeData
	InsertCommand.CollegeData = locCollegeData

	RankCommand.RankData = locRankData
	InsertCommand.RankData = locRankData

	RankCommand.PbAdmin = *PbAdmin
	InsertCommand.PbAdmin = *PbAdmin

	RankCommand.BotChannel = BotChannel
	if botEnv != nil {
		RankCommand.BotEnv = *botEnv
		InsertCommand.BotEnv = *botEnv
	}
}

func (b *Bot) registerCommands() []*discordgo.ApplicationCommand {
	b.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:
			if i.MessageComponentData().CustomID == "college_send_dm" {
				buttons.HandleSendToDMButton(s, i)
			} else if strings.HasPrefix(i.MessageComponentData().CustomID, "select_branch_") {
				RankCommand.HandleRankCutoffResponse(s, i)
			} else if strings.HasPrefix(i.MessageComponentData().CustomID, "select_analyze_branch") {
				RankCommand.HandleAnalyzeResponse(s, i)
			} else if strings.HasPrefix(i.MessageComponentData().CustomID, "anext_") || strings.HasPrefix(i.MessageComponentData().CustomID, "aprev_") {
				RankCommand.HandleAnalyzePagination(s, i)
			}
		case discordgo.InteractionApplicationCommandAutocomplete:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
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
