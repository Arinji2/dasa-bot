package rank

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	commands_utils "github.com/arinji2/dasa-bot/commands"
	"github.com/bwmarrin/discordgo"
)

var AnalyzeBranch = []string{
	"Computer Science",
	"Electrical Engineering",
	"Mechanical Engineering",
	"Civil Engineering",
	"Chemical Engineering",
	"Electronics Engineering",
	"Information Technology",
}

var Devations = []string{
	"10", "20", "30", "40",
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

	collegeData, err := r.getCollegeData(collegeID)
	if err != nil {
		log.Printf("Error fetching college data: %v", err)
		commands_utils.RespondWithEphemeralError(s, i, "Could not retrieve college data")
		return
	}

	branches := r.branchesForCollege(collegeData.ID, ciwgBool, yearInt, roundInt)
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

func (r *RankCommand) showAnalyzeBranchSelect(s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	rank := data.Options[0].StringValue()
	ciwg := data.Options[1].StringValue()
	deviation := "10"
	if len(data.Options) == 3 {
		deviation = data.Options[2].StringValue()
		if !slices.Contains(Devations, deviation) {
			deviation = "10"
		}
	}

	ciwgBool := (ciwg == "true")

	var components []discordgo.MessageComponent
	var options []discordgo.SelectMenuOption
	for i, branch := range AnalyzeBranch {
		desc := branch

		options = append(options, discordgo.SelectMenuOption{
			Label:       branch,
			Description: desc,
			Value:       fmt.Sprintf("%s,%s,%s,%d", rank, ciwg, deviation, i),
		})
	}

	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    "select_analyze_branch",
				Placeholder: "Select a branch to see its analysis with your rank",
				Options:     options,
			},
		},
	})

	title := "Analyze for Rank"
	description := fmt.Sprintf("**Rank:** %s\n**%s Student**\n\nPlease select a branch to view cutoffs",
		rank,
		map[bool]string{true: "CIWG", false: "Non-CIWG"}[ciwgBool])

	err := commands_utils.RespondWithEphemeralEmbedAndComponents(s, i, r.BotEnv, title, description, nil, components)
	if err != nil {
		log.Printf("Error sending branch selection UI: %v", err)
	}
}

func (r *RankCommand) handleCollegeBranches(s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
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

	collegeData, err := r.getCollegeData(collegeID)
	if err != nil {
		log.Printf("Error fetching college data: %v", err)
		commands_utils.RespondWithEphemeralError(s, i, "Could not retrieve college data")
		return
	}

	branches := r.branchesForCollege(collegeData.ID, ciwgBool, yearInt, roundInt)

	if len(branches) == 0 {
		commands_utils.RespondWithEphemeralError(s, i, fmt.Sprintf("No branches found for %s with the selected criteria", collegeData.Name))
		return
	}

	title := fmt.Sprintf("Cutoffs for %s", collegeData.Name)
	description := fmt.Sprintf("**Year:** %s\n**Round:** %s\n**%s Student**\n\n",
		year, round,
		map[bool]string{true: "CIWG", false: "Non-CIWG"}[ciwgBool])

	fields := []*discordgo.MessageEmbedField{}

	rankData, err := r.ranksForCollege(collegeData.ID, ciwgBool, yearInt, roundInt)
	if err != nil {
		commands_utils.RespondWithEphemeralError(s, i, fmt.Sprintf("No ranks found for %s with the selected criteria", collegeData.Name))
		return
	}

	for _, rank := range rankData {
		ciwgString := ""
		if ciwgBool {
			ciwgString = " (CIWG)"
		}

		fieldCiwgString := "DASA"
		if ciwgBool {
			fieldCiwgString = "CIWG"
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s%s", rank.Expand.Branch.Code, ciwgString),
			Value:  fmt.Sprintf("JEE OPENING: %d\n JEE CLOSING: %d\n %s OPENING: %d\n %s CLOSING: %d", rank.JEE_OPEN, rank.JEE_CLOSE, fieldCiwgString, rank.DASA_OPEN, fieldCiwgString, rank.DASA_CLOSE),
			Inline: true,
		})
	}

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

	collegeData, err := r.getCollegeData(collegeID)
	if err != nil {
		log.Printf("Error fetching college data: %v", err)
		return
	}

	rankData, err := r.specificRank(collegeData.ID, branchCode, ciwgBool, yearInt, roundInt)
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

	ciwgString := "DASA"
	if ciwgBool {
		ciwgString = " CIWG"
	}

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
			Name:   fmt.Sprintf("%s Opening Rank", ciwgString),
			Value:  fmt.Sprintf("%d", rankData.DASA_OPEN),
			Inline: true,
		},
		{
			Name:   fmt.Sprintf("%s Closing Rank", ciwgString),
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

