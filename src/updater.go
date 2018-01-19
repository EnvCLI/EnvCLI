package main

import (
	"fmt"
	"runtime"
	"strings"
	"context"
	log "github.com/sirupsen/logrus"    // imports as package "log"
	"net/http"
  "github.com/inconshreveable/go-update"
	"github.com/blang/semver"
	"github.com/google/go-github/github"
)

/**
 * The Application Update Configuration
 * Using equinox.io
 */
type ApplicationUpdater struct {
	BintrayOrg  string
	BintrayRepository string
	BintrayPackage string
	GitHubOrg string
	GitHubRepository string
}

/**
 * Update to version `version`
 */
func (appUpdater ApplicationUpdater) update(version string, force bool) {
	// current application version
	applicationVersion, err := semver.Make(strings.TrimLeft(appVersion, "v"))
	if err != nil {
		 log.Errorf("Unexpected Error: %s", err)
		 return
	}

	// Get Latest Version from Link
	if(version == "latest") {
		ctx := context.Background()
		client := github.NewClient(nil)
		opt := &github.ListOptions{Page: 0, PerPage: 500}
		tags, _, err := client.Repositories.ListTags(ctx, appUpdater.GitHubOrg, appUpdater.GitHubRepository, opt)
		if err != nil {
			 log.Errorf("Unexpected GitHub Error: %s", err)
			 return
		}

		// Find newest tag
		currentVersion, _ := semver.Make("0.0.0")
		for _, tag := range tags {
			log.Debugf("Found Tag in Source Repository: %s [%s] %d", *tag.Name, *tag.Commit.SHA)

			tagVersion, err := semver.Make(strings.TrimLeft(*tag.Name, "v"))
			if err != nil {
				 log.Debugf("Unexpected error parsing the github tag: %s", err)
				 continue
			}

			if tagVersion.GTE(applicationVersion) && tagVersion.GTE(currentVersion) {
				currentVersion = tagVersion
				version = fmt.Sprintf("v%s", tagVersion.String())
			}
		}

		if version == "latest" {
			log.Errorf("Couldn't determinate the latest version.")
			return
		} else {
			log.Debugf("Latest version is [%s].", version)
		}
	}

	// update target version
	updateTargetVersion, err := semver.Make(strings.TrimLeft(version, "v"))
	if err != nil {
		 log.Errorf("Unexpected Error: %s", err)
		 return
	}

	if(applicationVersion.Compare(updateTargetVersion) == 0 && force == false) {
		log.Info("No update available, already at the latest version!")
		return
	}

	if force == true {
		log.Debugf("Initiating forced update to version: %s", updateTargetVersion.String())
	}

	var downloadUrl string = fmt.Sprintf("https://dl.bintray.com/%s/%s/%s/%s/envcli_%s_%s", appUpdater.BintrayOrg, appUpdater.BintrayRepository, appUpdater.BintrayPackage, version, runtime.GOOS, runtime.GOARCH)
	log.Debugf("Starting download from remote: %v", downloadUrl)

	// download new version
	resp, err := http.Get(downloadUrl)
  if err != nil {
		log.Errorf("Unexpected Error: %s", err)
		return
  }
	if resp.StatusCode != 200 {
		log.Error("Update not found on remote server ... aborting.")
		return
	}

  defer resp.Body.Close()
	opts := update.Options{}
	err = opts.CheckPermissions()
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

	// Log Result
	if applicationVersion.GT(updateTargetVersion) {
		log.Infof("Successfully downgraded from [%s] to [%s]!", applicationVersion.String(), updateTargetVersion.String())
	} else if applicationVersion.LT(updateTargetVersion) {
		log.Infof("Successfully upgraded from [%s] to [%s]!", applicationVersion.String(), updateTargetVersion.String())
	} else {
		log.Infof("Successfully downloaded [%s]!", applicationVersion.String())
	}
}
