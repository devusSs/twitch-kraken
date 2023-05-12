package updater

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/devusSs/twitch-kraken/internal/logging"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

const (
	updateURL = "https://api.github.com/repos/devusSs/twitch-kraken/releases/latest"
)

var (
	buildVersion = ""
	buildDate    = ""
	buildOS      = runtime.GOOS
	buildArch    = runtime.GOARCH
	goVersion    = runtime.Version()
)

// Function to print build information without log.
func PrintBuildInformationRaw() {
	fmt.Printf("Build version: \t\t%s\n", buildVersion)
	fmt.Printf("Build date: \t\t%s\n", buildDate)
	fmt.Printf("Build OS: \t\t%s\n", buildOS)
	fmt.Printf("Build arch: \t\t%s\n", buildArch)
	fmt.Printf("Go version: \t\t%s\n", goVersion)
}

// Queries the latest release from Github repo.
func FindLatestReleaseURL() (string, string, string, error) {
	resp, err := http.Get(updateURL)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	var release githubRelease

	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return "", "", "", err
	}

	// Fix versions / architecture to match Github releases.
	if buildArch == "amd64" {
		buildArch = "x86_64"
	}

	if buildArch == "386" {
		buildArch = "i386"
	}

	// Find matching release for our OS & architecture.
	for _, asset := range release.Assets {
		releaseName := strings.ToLower(asset.Name)

		if strings.Contains(releaseName, buildArch) && strings.Contains(releaseName, buildOS) {
			// Format the changelog body accordingly.
			changeSplit := strings.Split(strings.ReplaceAll(strings.TrimSpace(release.Body), "## Changelog", ""), "\n")
			for i, line := range changeSplit {
				changeSplit[i] = strings.ReplaceAll(fmt.Sprintf("\t\t\t%s", line), "*", "-")
			}
			changelog := strings.Join(changeSplit, "\n")
			return asset.BrowserDownloadURL, release.TagName, changelog, nil
		}
	}

	return "", "", "", errors.New("no matching release found")
}

// Compare current version with latest version
func NewerVersionAvailable(newVersion string) (bool, error) {
	currentBuild := strings.ReplaceAll(buildVersion, "v", "")
	newBuild := strings.ReplaceAll(newVersion, "v", "")

	vOld, err := semver.NewVersion(currentBuild)
	if err != nil {
		return false, err
	}

	vNew, err := semver.NewVersion(newBuild)
	if err != nil {
		return false, err
	}

	return !vNew.Equal(vOld), nil
}

// Perform the actual patch.
func DoUpdate(url string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	if err := selfupdate.UpdateTo(url, exe); err != nil {
		return err
	}

	return nil
}

func PeriodicUpdateCheck() error {
	_, versionCheck, _, err := FindLatestReleaseURL()
	if err != nil {
		return err
	}

	newVersionAvailable, err := NewerVersionAvailable(versionCheck)
	if err != nil {
		return err
	}

	if newVersionAvailable {
		log.Printf("[%s] New version available (%s). Please restart your app soon\n", logging.WarnSign, versionCheck)
	}

	return nil
}
