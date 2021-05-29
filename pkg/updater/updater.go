package updater

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"runtime"
	"strings"

	"github.com/blang/semver"
	github "github.com/google/go-github/v26/github"
	update "github.com/inconshreveable/go-update"
)

// Find the latest version of the applicaton
func (appUpdater ApplicationUpdater) getLatestVersion() string {
	// Get Latest Version from Link
	var version = ""
	ctx := context.Background()
	client := github.NewClient(nil)
	opt := &github.ListOptions{Page: 0, PerPage: 500}
	tags, _, err := client.Repositories.ListTags(ctx, appUpdater.GitHubOrg, appUpdater.GitHubRepository, opt)
	if err != nil {
		log.Error().Err(err).Msg("Unexpected GitHub Error")
		return ""
	}

	// Find newest tag
	currentVersion, _ := semver.Make("0.0.0")
	for _, tag := range tags {
		log.Debug().Msg("Found Tag in Source Repository: "+*tag.Name+" ["+*tag.Commit.SHA+"]")

		tagVersion, err := semver.Make(strings.TrimLeft(*tag.Name, "v"))
		if err != nil {
			log.Debug().Err(err).Msg("Unexpected error parsing the github tag: " + err.Error())
			continue
		}
		// GTE: sourceVersion greater than or equal to targetVersion
		if tagVersion.GTE(currentVersion) {
			currentVersion = tagVersion
			version = fmt.Sprintf("v%s", tagVersion.String())
		}
	}

	log.Debug().Msg("Latest version is "+version+".")
	return version
}

// applyUpdate ...
func applyUpdate(resp *http.Response) {
	opts := update.Options{}
	err := opts.CheckPermissions()
	if err != nil {
		log.Error().Err(err).Msg("Missing permissions, update can't be executed: "+err.Error())
		return
	}
	err = update.Apply(resp.Body, opts)
	if err != nil {
		if rerr := update.RollbackError(err); rerr != nil {
			log.Error().Err(err).Msg("Broken update, failed to rollback. Please reinstall the application.")
		} else {
			log.Error().Err(err).Msg("Broken update detected, aborted.")
		}
		return
	}
}

// newVersionDownloader ...
func (appUpdater ApplicationUpdater) newVersionDownloader(version string) {
	var downloadURL = fmt.Sprintf("https://github.com/EnvCLI/EnvCLI/releases/download/%s/%s_%s", version, runtime.GOOS, runtime.GOARCH)
	log.Debug().Msg("Starting download from remote: " + downloadURL)

	// download new version
	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Error().Err(err).Msg("Unexpected Error: "+err.Error())
		return
	}
	if resp.StatusCode != 200 {
		log.Error().Msg("Update not found on remote server ... aborting.")
		return
	}

	defer resp.Body.Close()
	applyUpdate(resp)
}

// Update interface
func (appUpdater ApplicationUpdater) Update(version string, force bool, appVersion string) {
	// current application version
	applicationVersion, err := semver.Make(strings.TrimLeft(appVersion, "v"))
	if err != nil {
		log.Error().Err(err).Msg("Unexpected Error: "+err.Error())
		return
	}

	// set to latest version of no version is specified
	if version == "latest" {
		version = appUpdater.getLatestVersion()
	}

	// update target version
	updateTargetVersion, err := semver.Make(strings.TrimLeft(version, "v"))
	if err != nil {
		log.Error().Err(err).Msg("Unexpected Error: "+err.Error())
		return
	}

	if applicationVersion.Compare(updateTargetVersion) == 0 && force == false {
		log.Info().Msg("No update available, already at the latest version!")
		return
	}

	if force == true {
		log.Debug().Msg("Initiating forced update to version: " + updateTargetVersion.String())
	}
	appUpdater.newVersionDownloader(version)

	// Log Result
	if applicationVersion.GT(updateTargetVersion) {
		log.Info().Msg("Successfully downgraded from ["+applicationVersion.String()+"] to ["+updateTargetVersion.String()+"]!")
	} else if applicationVersion.LT(updateTargetVersion) {
		log.Info().Msg("Successfully upgraded from ["+applicationVersion.String()+"] to ["+updateTargetVersion.String()+"]!")
	} else {
		log.Info().Msg("Successfully downloaded ["+applicationVersion.String()+"]!")
	}
}

// Update interface
func (appUpdater ApplicationUpdater) IsUpdateAvailable(appVersion string) bool {
	// current application version
	applicationVersion, err := semver.Make(strings.TrimLeft(appVersion, "v"))
	if err != nil {
		log.Error().Err(err).Msg("Unexpected Error: "+err.Error())
		return false
	}

	var version = appUpdater.getLatestVersion()

	// update target version
	updateTargetVersion, err := semver.Make(strings.TrimLeft(version, "v"))
	if err != nil {
		log.Error().Err(err).Msg("Unexpected Error: "+err.Error())
		return false
	}

	// Evaluate
	if applicationVersion.LT(updateTargetVersion) {
		return true
	}

	return false
}
