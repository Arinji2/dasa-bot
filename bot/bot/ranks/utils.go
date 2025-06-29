package rank

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/arinji2/dasa-bot/pb"
	responses "github.com/arinji2/dasa-bot/responses"
	"github.com/bwmarrin/discordgo"
)

var AnalyzeBranch = []string{
	"Computer Science: cse, computer science, cs",
	"Electrical Engineering: electronics, electrical, elec, ece, eee",
	"Mechanical Engineering",
	"Civil Engineering",
	"Chemical Engineering",
	"Information Technology",
	"Architecture: architecture, arch",
	"Metallurgical Engineering",
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
		responses.RespondWithEphemeralError(s, i, "Invalid year format")
		return
	}

	roundInt, err := strconv.Atoi(round)
	if err != nil {
		log.Printf("Error converting round to int: %v", err)
		responses.RespondWithEphemeralError(s, i, "Invalid round format")
		return
	}

	ciwgBool := (ciwg == "true")

	collegeData, err := r.getCollegeData(collegeID)
	if err != nil {
		log.Printf("Error fetching college data: %v", err)
		responses.RespondWithEphemeralError(s, i, "Could not retrieve college data")
		return
	}

	branches := r.branchesForCollege(collegeData.ID, ciwgBool, yearInt, roundInt)
	if len(branches) == 0 {
		responses.RespondWithEphemeralError(s, i, fmt.Sprintf("No branches found for %s with the selected criteria", collegeData.Name))
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

	err = responses.RespondWithAutoEmbedAndComponents(s, i, r.BotEnv, title, description, fields, components, r.BotChannel)
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
		branchName := branch
		if strings.Contains(branch, ":") {
			branchName = strings.Split(branch, ":")[0]
		}

		options = append(options, discordgo.SelectMenuOption{
			Label:       branchName,
			Description: branchName,
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

	err := responses.RespondWithAutoEmbedAndComponents(s, i, r.BotEnv, title, description, nil, components, r.BotChannel)
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
		responses.RespondWithEphemeralError(s, i, "Invalid year format")
		return
	}

	roundInt, err := strconv.Atoi(round)
	if err != nil {
		log.Printf("Error converting round to int: %v", err)
		responses.RespondWithEphemeralError(s, i, "Invalid round format")
		return
	}

	ciwgBool := (ciwg == "true")

	collegeData, err := r.getCollegeData(collegeID)
	if err != nil {
		log.Printf("Error fetching college data: %v", err)
		responses.RespondWithEphemeralError(s, i, "Could not retrieve college data")
		return
	}

	branches := r.branchesForCollege(collegeData.ID, ciwgBool, yearInt, roundInt)

	if len(branches) == 0 {
		responses.RespondWithEphemeralError(s, i, fmt.Sprintf("No branches found for %s with the selected criteria", collegeData.Name))
		return
	}

	title := fmt.Sprintf("Cutoffs for %s", collegeData.Name)
	description := fmt.Sprintf("**Year:** %s\n**Round:** %s\n**%s Student**\n\n",
		year, round,
		map[bool]string{true: "CIWG", false: "Non-CIWG"}[ciwgBool])

	fields := []*discordgo.MessageEmbedField{}

	rankData, err := r.ranksForCollege(collegeData.ID, ciwgBool, yearInt, roundInt)
	if err != nil {
		responses.RespondWithEphemeralError(s, i, fmt.Sprintf("No ranks found for %s with the selected criteria", collegeData.Name))
		return
	}

	for _, rank := range rankData {
		ciwgString := ""
		if ciwgBool {
			ciwgString = " (CIWG)"
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s%s", rank.Expand.Branch.Code, ciwgString),
			Value:  fmt.Sprintf("JEE OPENING: %d\n JEE CLOSING: %d", rank.JEE_OPEN, rank.JEE_CLOSE),
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

	err = responses.RespondWithAutoEmbedAndComponents(s, i, r.BotEnv, title, description, fields, components, r.BotChannel)
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
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{responses.CreateBaseEmbed(title, description, r.BotEnv, fields)},
		Components: &components,
	})
	if err != nil {
		log.Printf("Error updating message with cutoff data: %v", err)
	}
}

func (r *RankCommand) handleAnalyzeSelection(s *discordgo.Session, i *discordgo.InteractionCreate) {
	values := i.MessageComponentData().Values
	if len(values) != 1 {
		log.Printf("Unexpected number of values in branch selection: %v", len(values))
		return
	}

	// Format: rank,ciwg,deviation,branchID
	params := strings.Split(values[0], ",")
	if len(params) != 4 {
		log.Printf("Invalid branch selection value format for analyze selection")
		return
	}

	rank := params[0]
	ciwg := params[1]
	deviation := params[2]
	branchCode := params[3]

	branchID, err := strconv.Atoi(branchCode)
	if err != nil {
		log.Printf("Error converting branch code to int: %v", err)
		return
	}

	branchData := AnalyzeBranch[branchID]
	ciwgBool := (ciwg == "true")

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("Error acknowledging branch selection: %v", err)
		return
	}

	matchingRankChunks, err := r.findMatchingRanks(rank, deviation, branchData, ciwgBool)
	if err != nil {
		log.Printf("Error fetching matching data: %v", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{responses.CreateBaseEmbed("Could not find ranks", "Could not find any ranks matching your selections.", r.BotEnv, nil)},
		})

		return
	}

	// Start with page 0
	currentPage := 0
	r.displayAnalyzePage(s, i, matchingRankChunks, currentPage, branchData, ciwgBool, rank, ciwg, deviation, branchCode)
}

func (r *RankCommand) displayAnalyzePage(s *discordgo.Session, i *discordgo.InteractionCreate, matchingRankChunks [][]pb.RankCollection, currentPage int, branchData string, ciwgBool bool, rank, ciwg, deviation, branchCode string) {
	if currentPage < 0 || currentPage >= len(matchingRankChunks) {
		log.Printf("Invalid page number: %d", currentPage)
		return
	}

	matchingRanks := matchingRankChunks[currentPage]
	description := ""
	if ciwgBool {
		description += fmt.Sprintf("Course: %s (CIWG)\n", strings.Split(branchData, ":")[0])
	} else {
		description += fmt.Sprintf("Course: %s\n", strings.Split(branchData, ":")[0])
	}
	description += fmt.Sprintf("Page %d of %d", currentPage+1, len(matchingRankChunks))

	// Create pagination buttons
	var components []discordgo.MessageComponent
	var buttons []discordgo.MessageComponent

	// Only show pagination if there are multiple pages
	if len(matchingRankChunks) > 1 {
		// Previous page button
		prevPage := currentPage - 1
		if prevPage < 0 {
			prevPage = len(matchingRankChunks) - 1 // Loop to last page
		}
		buttons = append(buttons, discordgo.Button{
			Label:    "Previous",
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("aprev_%s_%s_%s_%s_%d", rank, ciwg, deviation, branchCode, prevPage),
		})

		// Current page button (disabled)
		buttons = append(buttons, discordgo.Button{
			Label:    fmt.Sprintf("%d/%d", currentPage+1, len(matchingRankChunks)),
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("acur_%d", currentPage),
			Disabled: true,
		})

		// Next page button
		nextPage := currentPage + 1
		if nextPage >= len(matchingRankChunks) {
			nextPage = 0 // Loop to first page
		}
		buttons = append(buttons, discordgo.Button{
			Label:    "Next",
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("anext_%s_%s_%s_%s_%d", rank, ciwg, deviation, branchCode, nextPage),
		})
	}

	// Add Send to DM button
	buttons = append(buttons, discordgo.Button{
		Label:    "Send To DM",
		Style:    discordgo.PrimaryButton,
		CustomID: "college_send_dm",
	})

	components = append(components, discordgo.ActionsRow{
		Components: buttons,
	})

	title := "Chances based off of your JEE(Main) CRL-Rank"

	fields := []*discordgo.MessageEmbedField{}
	for idx, rankData := range matchingRanks {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d. %s", (currentPage*10)+idx+1, rankData.Expand.College.Name),
			Value:  fmt.Sprintf("JEE CLOSING: %d\n BRANCH CODE: %s", rankData.JEE_CLOSE, rankData.Expand.Branch.Code),
			Inline: true,
		})
	}

	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{responses.CreateBaseEmbed(title, description, r.BotEnv, fields)},
		Components: &components,
	})
	if err != nil {
		log.Printf("Error updating message with analyze data: %v", err)
	}
}

func (r *RankCommand) HandleAnalyzePagination(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID

	var parts []string

	if strings.HasPrefix(customID, "anext_") {
		parts = strings.Split(customID, "_")
	} else if strings.HasPrefix(customID, "aprev_") {
		parts = strings.Split(customID, "_")
	} else {
		log.Printf("Invalid analyze pagination customID: %v", customID)
		return
	}

	// Format: anext_{rank}_{ciwg}_{deviation}_{branchCode}_{page} or aprev_{rank}_{ciwg}_{deviation}_{branchCode}_{page}
	if len(parts) != 6 {
		log.Printf("Invalid analyze pagination customID format: %v", customID)
		return
	}

	rank := parts[1]
	ciwg := parts[2]
	deviation := parts[3]
	branchCode := parts[4]
	pageStr := parts[5]

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		log.Printf("Error converting page to int: %v", err)
		return
	}

	branchID, err := strconv.Atoi(branchCode)
	if err != nil {
		log.Printf("Error converting branch code to int: %v", err)
		return
	}

	branchData := AnalyzeBranch[branchID]
	ciwgBool := (ciwg == "true")

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("Error acknowledging pagination: %v", err)
		return
	}

	matchingRankChunks, err := r.findMatchingRanks(rank, deviation, branchData, ciwgBool)
	if err != nil {
		log.Printf("Error fetching matching data for pagination: %v", err)
		return
	}

	r.displayAnalyzePage(s, i, matchingRankChunks, page, branchData, ciwgBool, rank, ciwg, deviation, branchCode)
}
