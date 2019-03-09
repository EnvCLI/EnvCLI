package updater

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	analytic "github.com/EnvCLI/EnvCLI/pkg/analytic"
	sentry "github.com/EnvCLI/EnvCLI/pkg/sentry"
	"github.com/blang/semver"
	github "github.com/google/go-github/github"
	update "github.com/inconshreveable/go-update"
	log "github.com/sirupsen/logrus"
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
		sentry.HandleError(err)
		log.Errorf("Unexpected GitHub Error: %s", err)
		return ""
	}

	// Find newest tag
	currentVersion, _ := semver.Make("0.0.0")
	for _, tag := range tags {
		log.Debugf("Found Tag in Source Repository: %s [%s]", *tag.Name, *tag.Commit.SHA)

		tagVersion, err := semver.Make(strings.TrimLeft(*tag.Name, "v"))
		if err != nil {
			log.Debugf("Unexpected error parsing the github tag: %s", err)
			continue
		}
		// GTE: sourceVersion greater than or equal to targetVersion
		if tagVersion.GTE(currentVersion) {
			currentVersion = tagVersion
			version = fmt.Sprintf("v%s", tagVersion.String())
		}
	}

	log.Debugf("Latest version is [%s].", version)
	return version
}

// applyUpdate ...
func applyUpdate(resp *http.Response) {
	opts := update.Options{}
	err := opts.CheckPermissions()
	if err != nil {
		log.Errorf("Missing permissions, update can't be executed: %s", err)
		return
	}
	err = update.Apply(resp.Body, opts)
	if err != nil {
		if rerr := update.RollbackError(err); rerr != nil {
			log.Errorf("Broken update, failed to rollback. Please reinstall the application. [%s]", err)
		} else {
			log.Errorf("Broken update detected, aborted. [%s]", err)
		}
		return
	}
}

// newVersionDownloader ...
func (appUpdater ApplicationUpdater) newVersionDownloader(version string) {
	var downloadURL = fmt.Sprintf("https://dl.bintray.com/%s/%s/%s/%s/envcli_%s_%s",
		appUpdater.BintrayOrg, appUpdater.BintrayRepository, appUpdater.BintrayPackage, version, runtime.GOOS, runtime.GOARCH)
	log.Debugf("Starting download from remote: %v", downloadURL)

	// download new version
	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Errorf("Unexpected Error: %s", err)
		return
	}
	if resp.StatusCode != 200 {
		log.Error("Update not found on remote server ... aborting.")
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
		sentry.HandleError(err)
		log.Errorf("Unexpected Error: %s", err)
		return
	}

	// set to latest version of no version is specified
	if version == "latest" {
		version = appUpdater.getLatestVersion()
	}

	// update target version
	updateTargetVersion, err := semver.Make(strings.TrimLeft(version, "v"))
	if err != nil {
		sentry.HandleError(err)
		log.Errorf("Unexpected Error: %s", err)
		return
	}

	if applicationVersion.Compare(updateTargetVersion) == 0 && force == false {
		log.Info("No update available, already at the latest version!")
		return
	}

	if force == true {
		log.Debugf("Initiating forced update to version: %s", updateTargetVersion.String())
	}
	appUpdater.newVersionDownloader(version)
	// Log Result
	if applicationVersion.GT(updateTargetVersion) {
		log.Infof("Successfully downgraded from [%s] to [%s]!", applicationVersion.String(), updateTargetVersion.String())
		analytic.TriggerEvent("Downgrade", updateTargetVersion.String())
	} else if applicationVersion.LT(updateTargetVersion) {
		log.Infof("Successfully upgraded from [%s] to [%s]!", applicationVersion.String(), updateTargetVersion.String())
		analytic.TriggerEvent("Update", updateTargetVersion.String())
	} else {
		log.Infof("Successfully downloaded [%s]!", applicationVersion.String())
		analytic.TriggerEvent("Update", updateTargetVersion.String())
	}
}

// Update interface
func (appUpdater ApplicationUpdater) IsUpdateAvailable(appVersion string) bool {
	// current application version
	applicationVersion, err := semver.Make(strings.TrimLeft(appVersion, "v"))
	if err != nil {
		sentry.HandleError(err)
		log.Errorf("Unexpected Error: %s", err)
		return false
	}

	var version = appUpdater.getLatestVersion()

	// update target version
	updateTargetVersion, err := semver.Make(strings.TrimLeft(version, "v"))
	if err != nil {
		sentry.HandleError(err)
		log.Errorf("Unexpected Error: %s", err)
		return false
	}

	// Evaluate
	if applicationVersion.LT(updateTargetVersion) {
		analytic.TriggerEvent("UpdateAvailable", "true")
		return true
	}

	return false
}
