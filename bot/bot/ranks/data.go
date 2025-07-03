package rank

import (
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/arinji2/dasa-bot/pb"
)

// Handles both id and name.
func (r *RankCommand) getCollegeData(collegeID string) (*pb.CollegeCollection, error) {
	if collegeID == "" {
		return nil, errors.New("no college ID provided")
	}
	isIDSearch := !strings.Contains(collegeID, " ")
	if isIDSearch {
		for _, v := range r.CollegeData {
			if v.ID == collegeID {
				return &v, nil
			}
		}
		return nil, errors.New("no college found with that ID")
	}
	for _, v := range r.CollegeData {
		if strings.EqualFold(v.Name, collegeID) {
			return &v, nil
		}
	}

	return nil, errors.New("no college found with that name")
}

func (r *RankCommand) branchesForCollege(collegeID string, ciwg bool, year, round int) []pb.BranchCollection {
	branches := []pb.BranchCollection{}
	for _, v := range r.RankData {
		if v.College == collegeID && v.Expand.Branch.Ciwg == ciwg && v.Year == year && v.Round == round {
			branches = append(branches, v.Expand.Branch)
		}
	}
	return branches
}

func (r *RankCommand) ranksForCollege(collegeID string, ciwg bool, year, round int) ([]pb.RankCollection, error) {
	rank := []pb.RankCollection{}
	for _, v := range r.RankData {
		if v.College == collegeID && v.Expand.Branch.Ciwg == ciwg && v.Year == year && v.Round == round {
			rank = append(rank, v)
		}
	}
	if len(rank) == 0 {
		return rank, errors.New("no ranks found for the selected criteria")
	}
	return rank, nil
}

func (r *RankCommand) specificRank(collegeID, branchCode string, ciwg bool, year, round int) (pb.RankCollection, error) {
	rank := pb.RankCollection{}
	for _, v := range r.RankData {
		if v.College == collegeID && v.Expand.Branch.Code == branchCode && v.Expand.Branch.Ciwg == ciwg && v.Year == year && v.Round == round {
			rank = v
		}
	}
	if rank.ID == "" {
		return rank, errors.New("no rank found for the selected criteria")
	}
	return rank, nil
}

func (r *RankCommand) findMatchingRanks(rankStr, deviationStr, branchData string, ciwg bool) ([][]pb.RankCollection, error) {
	inputRank, err := strconv.Atoi(rankStr)
	if err != nil {
		return nil, errors.New("invalid rank value")
	}

	deviationPercent, err := strconv.Atoi(deviationStr)
	if err != nil {
		return nil, errors.New("invalid deviation percentage")
	}

	lowerBound := inputRank - (inputRank * deviationPercent / 100)

	var collegeToRank []pb.RankCollection

	branchKeywordsList := []string{}

	if strings.Contains(branchData, ":") {
		branchKeywords := strings.Split(branchData, ":")[1]
		keywords := strings.SplitSeq(branchKeywords, ",")
		for k := range keywords {
			branchKeywordsList = append(branchKeywordsList, strings.TrimSpace(k))
		}
	} else {
		branchKeywordsList = append(branchKeywordsList, strings.TrimSpace(branchData))
	}

	type keywordRegex struct {
		raw    string
		regexp *regexp.Regexp
	}

	var compiledKeywords []keywordRegex
	for _, k := range branchKeywordsList {
		trimmed := strings.TrimSpace(strings.ToLower(k))
		pattern := `\b` + regexp.QuoteMeta(trimmed) + `\b`
		compiledKeywords = append(compiledKeywords, keywordRegex{
			raw:    trimmed,
			regexp: regexp.MustCompile(pattern),
		})
	}

	latestYear := r.RankData[0].Year
	latestRound := r.RankData[0].Round

	for _, v := range r.RankData {
		if v.Expand.Branch.Ciwg != ciwg {
			continue
		}

		if v.Year != latestYear {
			continue
		}

		if v.Round != latestRound {
			continue
		}

		match := false
		branchName := strings.ToLower(v.Expand.Branch.Name)
		branchCode := strings.ToLower(v.Expand.Branch.Code)

		for _, keyword := range compiledKeywords {
			if keyword.regexp.MatchString(branchName) || keyword.regexp.MatchString(branchCode) {
				match = true
				break
			}
		}
		if !match {
			continue
		}

		openRank := v.JeeOpen
		closeRank := v.JeeClose

		if openRank == 0 && closeRank == 0 {
			continue
		}

		if closeRank >= lowerBound {
			collegeToRank = append(collegeToRank, v)
		}
	}

	if len(collegeToRank) == 0 {
		return nil, errors.New("no ranks matched the given criteria")
	}

	sort.Slice(collegeToRank, func(i, j int) bool {
		return collegeToRank[i].JeeClose < collegeToRank[j].JeeClose
	})

	var chunks [][]pb.RankCollection
	currentChunk := []pb.RankCollection{}

	for _, rank := range collegeToRank {
		currentChunk = append(currentChunk, rank)
		if len(currentChunk) == 10 {
			chunks = append(chunks, currentChunk)
			currentChunk = []pb.RankCollection{}
		}
	}
	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks, nil
}
