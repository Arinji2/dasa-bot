package pb

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"sync"

	"github.com/arinji2/dasa-bot/network"
)

func (p *PocketbaseAdmin) GetAllRanks() ([]RankCollection, error) {
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return nil, err
	}
	parsedURL.Path = "/api/collections/ranks/records"

	perPage := 1000
	params := url.Values{}
	params.Add("perPage", strconv.Itoa(perPage))
	params.Add("expand", "college,branch")
	params.Add("page", "1")
	params.Add("sort", "-year,-round")

	parsedURL.RawQuery = params.Encode()

	type request struct{}
	responseBody, err := network.MakeAuthenticatedRequest(parsedURL, "GET", request{}, p.Token)
	if err != nil {
		return nil, err
	}

	var initialResponse PbResponse[RankCollection]
	err = json.Unmarshal(responseBody, &initialResponse)
	if err != nil {
		return nil, err
	}

	total := initialResponse.TotalItems
	pages := (total + perPage - 1) / perPage // ceil division
	allItems := initialResponse.Items

	if pages <= 1 {
		return allItems, nil
	}

	type result struct {
		Items []RankCollection
		Err   error
	}

	resultsChan := make(chan result, pages-1)
	var wg sync.WaitGroup

	for page := 2; page <= pages; page++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			pageURL := *parsedURL
			query := url.Values{}
			query.Add("perPage", strconv.Itoa(perPage))
			query.Add("page", strconv.Itoa(page))
			query.Add("expand", "college,branch")
			query.Add("sort", "-year,-round")
			pageURL.RawQuery = query.Encode()

			body, err := network.MakeAuthenticatedRequest(&pageURL, "GET", request{}, p.Token)
			if err != nil {
				resultsChan <- result{Err: err}
				return
			}

			var resp PbResponse[RankCollection]
			if err := json.Unmarshal(body, &resp); err != nil {
				resultsChan <- result{Err: err}
				return
			}

			resultsChan <- result{Items: resp.Items}
		}(page)
	}

	wg.Wait()
	close(resultsChan)

	for res := range resultsChan {
		if res.Err != nil {
			return nil, res.Err
		}
		allItems = append(allItems, res.Items...)
	}

	return allItems, nil
}

// GetSpecificRank by Year and Round for a College and Branch
func (p *PocketbaseAdmin) GetSpecificRank(college string, branch string, year int, round int, ciwg bool) (RankCollection, error) {
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return RankCollection{}, fmt.Errorf("failed to parse base domain: %w", err)
	}
	parsedURL.Path = "/api/collections/ranks/records"

	params := url.Values{}
	params.Add("filter", fmt.Sprintf("year='%d' && round='%d' && college.id='%s' && branch.code='%s' && branch.ciwg=%t", year, round, college, branch, ciwg))
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

	if len(response.Items) == 0 {
		return RankCollection{}, fmt.Errorf("no rank found for year: %d, round: %d, college: %s, branch: %s", year, round, college, branch)
	}

	return response.Items[0], nil
}

func (p *PocketbaseAdmin) GetSpecificRankWithBranchID(college string, branch string, year int, round int, ciwg bool) (RankCollection, error) {
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return RankCollection{}, fmt.Errorf("failed to parse base domain: %w", err)
	}
	parsedURL.Path = "/api/collections/ranks/records"

	params := url.Values{}
	params.Add("filter", fmt.Sprintf("year='%d' && round='%d' && college.id='%s' && branch.id='%s' && branch.ciwg=%t", year, round, college, branch, ciwg))

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

	if len(response.Items) == 0 {
		return RankCollection{}, fmt.Errorf("no rank found for year: %d, round: %d, college: %s, branch: %s", year, round, college, branch)
	}

	return response.Items[0], nil
}

// GetRanksByCollegeBranch for a College and Branch
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

// GetRanksByYearAndRound for a Year and Round
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

func (p *PocketbaseAdmin) CreateRank(rank RankCreateRequest, ciwg bool) (RankCollection, bool, error) {
	_, err := p.GetSpecificRankWithBranchID(rank.College, rank.Branch, rank.Year, rank.Round, ciwg)
	if err == nil {
		return RankCollection{}, true, fmt.Errorf("rank already exists")
	}
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return RankCollection{}, false, fmt.Errorf("failed to parse base domain: %w", err)
	}
	parsedURL.Path = "/api/collections/ranks/records"

	params := url.Values{}
	params.Add("expand", "college,branch")

	parsedURL.RawQuery = params.Encode()

	responseBody, err := network.MakeAuthenticatedRequest(parsedURL, "POST", rank, p.Token)
	if err != nil {
		return RankCollection{}, false, err
	}

	var response RankCollection
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return RankCollection{}, false, err
	}

	return response, false, nil
}
