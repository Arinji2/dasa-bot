package pb

import (
	"slices"
	"strings"
)

// Make sure we remove any other college with the same alias
func handleMultipleCollegeAlias(colleges []CollegeCollection, alias string) ([]CollegeCollection, error) {
	finalizedColleges := []CollegeCollection{}
	for _, college := range colleges {
		collegeAlias := strings.Split(college.Alias, ",")
		if slices.Contains(collegeAlias, alias) {
			finalizedColleges = append(finalizedColleges, college)
		}
	}
	return finalizedColleges, nil
}
