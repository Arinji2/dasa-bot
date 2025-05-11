package pb

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/arinji2/dasa-bot/network"
)

func (p *PocketbaseAdmin) GetAllColleges() ([]CollegeCollection, error) {
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return nil, err
	}
	parsedURL.Path = "/api/collections/colleges/records"

	params := url.Values{}
	params.Add("perPage", "10000")

	parsedURL.RawQuery = params.Encode()

	type request struct{}
	responseBody, err := network.MakeAuthenticatedRequest(parsedURL, "GET", request{}, p.Token)
	if err != nil {
		return nil, err
	}

	var response PbResponse[CollegeCollection]
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, err
	}
	return response.Items, nil
}

func (p *PocketbaseAdmin) GetCollegeByID(id string) (CollegeCollection, error) {
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return CollegeCollection{}, err
	}
	parsedURL.Path = "/api/collections/colleges/records"

	params := url.Values{}
	params.Add("filter", fmt.Sprintf("id='%s'", id))

	parsedURL.RawQuery = params.Encode()

	type request struct{}
	responseBody, err := network.MakeAuthenticatedRequest(parsedURL, "GET", request{}, p.Token)
	if err != nil {
		return CollegeCollection{}, err
	}

	var response PbResponse[CollegeCollection]
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return CollegeCollection{}, err
	}

	if len(response.Items) == 0 {
		return CollegeCollection{}, fmt.Errorf("no college found for id: %s", id)
	}

	// if response.TotalItems > 1 {
	// 	response.Items, err = handleMultipleCollegeAlias(response.Items, id)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return response.Items[0], nil
}
