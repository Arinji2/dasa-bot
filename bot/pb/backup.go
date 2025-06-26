package pb

import (
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

func (p *PocketbaseAdmin) CreateBackup(userName string) error {
	parsedURL, err := url.Parse(p.BaseDomain)
	if err != nil {
		return err
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
		return err
	}

	return nil
}

func sanitizeUsername(input string) string {
	input = strings.ToLower(input)

	// Replace all non-alphanumeric characters with "_"
	re := regexp.MustCompile(`[^a-z0-9]+`)
	safe := re.ReplaceAllString(input, "_")

	return safe
}
