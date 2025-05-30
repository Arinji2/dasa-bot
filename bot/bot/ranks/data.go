package rank

import (
	"errors"
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

	collegeSet := make(map[string]struct{})
	collegeToRank := make(map[string]pb.RankCollection)

	for _, v := range r.RankData {
		if v.Expand.Branch.Ciwg != ciwg {
			continue
		}

		if !strings.Contains(strings.ToLower(v.Expand.Branch.Name), strings.ToLower(branchData)) {
			continue
		}

		openRank := v.JEE_OPEN
		closeRank := v.JEE_CLOSE

		if openRank == 0 && closeRank == 0 {
			continue
		}

		if closeRank >= inputRank || openRank >= inputRank || (openRank <= inputRank && closeRank >= lowerBound) {
			if _, exists := collegeSet[v.College]; !exists {
				collegeSet[v.College] = struct{}{}
				collegeToRank[v.College] = v
			}
		}
	}

	if len(collegeToRank) == 0 {
		return nil, errors.New("no ranks matched the given criteria")
	}

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
