package rank

import "github.com/arinji2/dasa-bot/pb"

func (r *RankCommand) branchesForCollege(collegeID string, ciwg bool, year, round int) []pb.BranchCollection {
	branches := []pb.BranchCollection{}
	for _, v := range r.RankData {
		if v.College == collegeID && v.Expand.Branch.Ciwg == ciwg && v.Year == year && v.Round == round {
			branches = append(branches, v.Expand.Branch)
		}
	}
	return branches
}
