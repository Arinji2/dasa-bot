package pb

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/arinji2/dasa-bot/network"
)

type BackupCeateRequest struct {
	Name string `json:"name"`
}

func (p *PocketbaseAdmin) ListBackups() ([]BackupCollection, error) {
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return nil, err
	}
	parsedURL.Path = "/api/backups"

	type request struct{}
	responseBody, err := network.MakeAuthenticatedRequest(parsedURL, "GET", request{}, p.Token)
	if err != nil {
		return nil, err
	}
	var response []BackupCollection
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (p *PocketbaseAdmin) DeleteBackup(key string) error {
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return err
	}
	parsedURL.Path = fmt.Sprintf("/api/backups/%s", key)

	type request struct{}
	_, err = network.MakeAuthenticatedRequest(parsedURL, "DELETE", request{}, p.Token)
	if err != nil {
		return err
	}

	return nil
}

func (p *PocketbaseAdmin) CreateBackup(userName string) (string, error) {
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return "", err
	}
	parsedURL.Path = "/api/backups"
	userName = sanitizeUsername(userName)

	backupName := fmt.Sprintf("%s_%s.zip", userName, time.Now().Format("02_01_2006_15_04_05"))

	type request struct {
		Name string `json:"name"`
	}
	_, err = network.MakeAuthenticatedRequest(parsedURL, "POST", BackupCeateRequest{
		Name: backupName,
	}, p.Token)
	if err != nil {
		return "", err
	}

	return backupName, nil
}

func sanitizeUsername(input string) string {
	input = strings.ToLower(input)

	re := regexp.MustCompile(`[^a-z0-9]+`)
	safe := re.ReplaceAllString(input, "")

	return safe
}
