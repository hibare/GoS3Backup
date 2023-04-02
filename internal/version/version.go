package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hibare/GoS3Backup/internal/constants"
	log "github.com/sirupsen/logrus"
)

var (
	CurrentVersion            = "0.0.0"
	LatestVersion             = CurrentVersion
	UpdateNotificationMessage = "New update available: %s"
)

type Release struct {
	TagName string `json:"tag_name"`
}

func CheckLatestRelease() {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", constants.GithubOwner, constants.ProgramIdentifier)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Warnf("Error creating request: %v", err)
		return
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Warnf("Error getting latest release: %v", err)
		return
	}

	defer resp.Body.Close()

	var release Release
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		log.Warnf("Error unmarshalling response body: %v", err)
		return
	}

	LatestVersion = strings.TrimPrefix(release.TagName, "v")

	status, _ := IsNewVersionAvailable()
	log.Infof("Version update available: %v, Current version: %s, Latest Version: %s", status, CurrentVersion, LatestVersion)
}

func IsNewVersionAvailable() (bool, string) {
	if CurrentVersion != LatestVersion {
		return true, LatestVersion
	}
	return false, ""
}
