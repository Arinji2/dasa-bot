package rank

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	commands_utils "github.com/arinji2/dasa-bot/commands"
	"github.com/arinji2/dasa-bot/pb"
	"github.com/bwmarrin/discordgo"
)

func (r *RankCommand) branchesForCollege(collegeID string, ciwg bool, year, round int) []pb.BranchCollection {
	branches := []pb.BranchCollection{}
	for _, v := range r.RankData {
		if v.College == collegeID && v.Expand.Branch.Ciwg == ciwg && v.Year == year && v.Round == round {
			branches = append(branches, v.Expand.Branch)
		}
	}
	return branches
}

func (r *RankCommand) showBranchSelect(s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	collegeID := data.Options[0].StringValue()
	year := data.Options[1].StringValue()
	ciwg := data.Options[2].StringValue()
	round := data.Options[3].StringValue()

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		log.Printf("Error converting year to int: %v", err)
		commands_utils.RespondWithEphemeralError(s, i, "Invalid year format")
		return
	}

	roundInt, err := strconv.Atoi(round)
	if err != nil {
		log.Printf("Error converting round to int: %v", err)
		commands_utils.RespondWithEphemeralError(s, i, "Invalid round format")
		return
	}

	ciwgBool := (ciwg == "true")

	collegeData, err := r.PbAdmin.GetCollegeByID(collegeID)
	if err != nil {
		log.Printf("Error fetching college data: %v", err)
		commands_utils.RespondWithEphemeralError(s, i, "Could not retrieve college data")
		return
	}

	branches := r.branchesForCollege(collegeID, ciwgBool, yearInt, roundInt)
	if len(branches) == 0 {
		commands_utils.RespondWithEphemeralError(s, i, fmt.Sprintf("No branches found for %s with the selected criteria", collegeData.Name))
		return
	}

	var components []discordgo.MessageComponent
	pageNumber := 0
	for idx := 0; idx < len(branches); idx += maxSelectOptions {
		pageNumber++
		end := idx + maxSelectOptions
		end = min(end, len(branches))

		var options []discordgo.SelectMenuOption
		for _, branch := range branches[idx:end] {
			desc := branch.Code
			if len(desc) > 100 {
				desc = desc[:97] + "..."
			}

			options = append(options, discordgo.SelectMenuOption{
				Label:       branch.Name,
				Description: desc,
				Value:       fmt.Sprintf("%s,%s,%s,%s,%s", collegeID, year, ciwg, round, branch.Code),
			})
		}

		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("select_branch_%s_%d", collegeID, idx/maxSelectOptions),
					Placeholder: fmt.Sprintf("Select a branch to see cutoffs (List %d)", pageNumber),
					Options:     options,
				},
			},
		})
	}

	title := fmt.Sprintf("Cutoffs for %s", collegeData.Name)
	description := fmt.Sprintf("**Year:** %s\n**Round:** %s\n**%s Student**\n\nPlease select a branch to view cutoffs",
		year, round,
		map[bool]string{true: "CIWG", false: "Non-CIWG"}[ciwgBool])

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Total Available Branches",
			Value:  fmt.Sprintf("%d", len(branches)),
			Inline: true,
		},
	}

	err = commands_utils.RespondWithEphemeralEmbedAndComponents(s, i, r.BotEnv, title, description, fields, components)
	if err != nil {
		log.Printf("Error sending branch selection UI: %v", err)
	}
}

func (r *RankCommand) handleBranchSelection(s *discordgo.Session, i *discordgo.InteractionCreate) {
	values := i.MessageComponentData().Values
	if len(values) != 1 {
		log.Printf("Unexpected number of values in branch selection: %v", len(values))
		return
	}

	// Format: collegeID,year,ciwg,round,branchCode
	params := strings.Split(values[0], ",")
	if len(params) != 5 {
		log.Printf("Invalid branch selection value format: %v", values[0])
		return
	}

	collegeID := params[0]
	year := params[1]
	ciwg := params[2]
	round := params[3]
	branchCode := params[4]

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		log.Printf("Error converting year to int: %v", err)
		return
	}

	roundInt, err := strconv.Atoi(round)
	if err != nil {
		log.Printf("Error converting round to int: %v", err)
		return
	}

	ciwgBool := (ciwg == "true")

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("Error acknowledging branch selection: %v", err)
		return
	}

	collegeData, err := r.PbAdmin.GetCollegeByID(collegeID)
	if err != nil {
		log.Printf("Error fetching college data: %v", err)
		return
	}

	rankData, err := r.PbAdmin.GetSpecificRank(collegeID, branchCode, yearInt, roundInt, ciwgBool)
	if err != nil {
		log.Printf("Error fetching rank data: %v", err)
		return
	}

	description := ""
	if rankData.Expand.Branch.Ciwg {
		description += fmt.Sprintf("Course: %s (CIWG)\n", rankData.Expand.Branch.Name)
	} else {
		description += fmt.Sprintf("Course: %s\n", rankData.Expand.Branch.Name)
	}
	description += fmt.Sprintf("Branch Code: %s\nRound %d (%d)", rankData.Expand.Branch.Code, roundInt, yearInt)

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Send To DM",
					Style:    discordgo.PrimaryButton,
					CustomID: "college_send_dm",
				},
			},
		},
	}

	title := fmt.Sprintf("Cutoffs for %s", collegeData.Name)

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "JEE Opening Rank",
			Value:  fmt.Sprintf("%d", rankData.JEE_OPEN),
			Inline: true,
		},
		{
			Name:   "JEE Closing Rank",
			Value:  fmt.Sprintf("%d", rankData.JEE_CLOSE),
			Inline: true,
		},
		{
			Name:   "DASA Opening Rank",
			Value:  fmt.Sprintf("%d", rankData.DASA_OPEN),
			Inline: true,
		},
		{
			Name:   "DASA Closing Rank",
			Value:  fmt.Sprintf("%d", rankData.DASA_CLOSE),
			Inline: true,
		},
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{commands_utils.CreateBaseEmbed(title, description, r.BotEnv, fields)},
		Components: &components,
	})
	if err != nil {
		log.Printf("Error updating message with cutoff data: %v", err)
	}
}
