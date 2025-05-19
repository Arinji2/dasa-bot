package rank

import (
	"errors"
	"fmt"
	"strings"

	"github.com/arinji2/dasa-bot/pb"
)

// Handles both id and name.
func (r *RankCommand) getCollegeData(collegeID string) (*pb.CollegeCollection, error) {
	fmt.Println(collegeID)
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
