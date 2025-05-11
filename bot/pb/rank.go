package pb

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/arinji2/dasa-bot/network"
)

// Get Rank by Year and Round for a College and Branch
func (p *PocketbaseAdmin) GetSpecificRank(college string, branch string, year int, round int) (RankCollection, error) {
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return RankCollection{}, fmt.Errorf("failed to parse base domain: %w", err)
	}
	parsedURL.Path = "/api/collections/ranks/records"

	params := url.Values{}
	params.Add("filter", fmt.Sprintf("year='%d' && round='%d' && college.alias~'%s' && branch.code='%s'", year, round, college, branch))
	params.Add("expand", "college,branch")

	parsedURL.RawQuery = params.Encode()

	type request struct{}
	responseBody, err := network.MakeAuthenticatedRequest(parsedURL, "GET", request{}, p.Token)
	if err != nil {
		return RankCollection{}, fmt.Errorf("failed to make authenticated request: %w", err)
	}

	var response PbResponse[RankCollection]
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return RankCollection{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.Items[0], nil
}

// Get All Ranks for a College and Branch
func (p *PocketbaseAdmin) GetRanksByCollegeBranch(college string, branch string) ([]RankCollection, error) {
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base domain: %w", err)
	}
	parsedURL.Path = "/api/collections/ranks/records"

	params := url.Values{}
	params.Add("filter", fmt.Sprintf("college.alias~'%s' && branch.code='%s'", college, branch))
	params.Add("expand", "college,branch")

	parsedURL.RawQuery = params.Encode()

	type request struct{}
	responseBody, err := network.MakeAuthenticatedRequest(parsedURL, "GET", request{}, p.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to make authenticated request: %w", err)
	}

	var response PbResponse[RankCollection]
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.Items, nil
}

// Get All Ranks for a Year and Round
func (p *PocketbaseAdmin) GetRanksByYearAndRound(year int, round int) ([]RankCollection, error) {
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base domain: %w", err)
	}
	parsedURL.Path = "/api/collections/ranks/records"

	params := url.Values{}
	params.Add("filter", fmt.Sprintf("year='%d' && round='%d'", year, round))
	params.Add("expand", "college,branch")

	parsedURL.RawQuery = params.Encode()

	type request struct{}
	responseBody, err := network.MakeAuthenticatedRequest(parsedURL, "GET", request{}, p.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to make authenticated request: %w", err)
	}

	var response PbResponse[RankCollection]
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.Items, nil
}
