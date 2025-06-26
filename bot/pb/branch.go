package pb

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/arinji2/dasa-bot/network"
)

func (p *PocketbaseAdmin) GetAllBranches() ([]BranchCollection, error) {
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return nil, err
	}
	parsedURL.Path = "/api/collections/branches/records"

	params := url.Values{}
	params.Add("perPage", "10000")

	parsedURL.RawQuery = params.Encode()

	type request struct{}
	responseBody, err := network.MakeAuthenticatedRequest(parsedURL, "GET", request{}, p.Token)
	if err != nil {
		return nil, err
	}

	var response PbResponse[BranchCollection]
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, err
	}

	return response.Items, nil
}

func (p *PocketbaseAdmin) GetBranchByCode(code string) (BranchCollection, error) {
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return BranchCollection{}, err
	}
	parsedURL.Path = "/api/collections/branches/records"

	params := url.Values{}
	params.Add("filter", fmt.Sprintf("code='%s'", code))

	parsedURL.RawQuery = params.Encode()

	type request struct{}
	responseBody, err := network.MakeAuthenticatedRequest(parsedURL, "GET", request{}, p.Token)
	if err != nil {
		return BranchCollection{}, err
	}

	var response PbResponse[BranchCollection]
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return BranchCollection{}, err
	}

	if len(response.Items) == 0 {
		return BranchCollection{}, fmt.Errorf("no branch found for code: %s", code)
	}

	return response.Items[0], nil
}

func (p *PocketbaseAdmin) CreateBranch(branch BranchCreateRequest) (BranchCollection, error) {
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return BranchCollection{}, err
	}
	parsedURL.Path = "/api/collections/branches/records"

	type request struct{}
	responseBody, err := network.MakeAuthenticatedRequest(parsedURL, "POST", BranchCreateRequest{
		Name: branch.Name,
		Code: branch.Code,
		Ciwg: branch.Ciwg,
	}, p.Token)
	if err != nil {
		return BranchCollection{}, err
	}

	var response BranchCollection
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return BranchCollection{}, err
	}

	return response, nil
}
