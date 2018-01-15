package main

import (
	"fmt"
	"runtime"
	"strings"
	log "github.com/sirupsen/logrus"    // imports as package "log"
	"net/http"
  "github.com/inconshreveable/go-update"
	"github.com/blang/semver"
)

/**
 * The Application Update Configuration
 * Using equinox.io
 */
type ApplicationUpdater struct {
	BintrayOrg  string
	BintrayRepository string
	BintrayPackage string
}

/**
 * Update to version `version`
 */
func (appUpdater ApplicationUpdater) update(version string) {
	// Get Latest Version from Link
	if(version == "latest") {
		resp, err := http.Get(fmt.Sprintf("https://bintray.com/%s/%s/%s/_latestVersion", appUpdater.BintrayOrg, appUpdater.BintrayRepository, appUpdater.BintrayPackage))
	  if err != nil {
	      log.Fatalf("Unexpected Error: %s", err)
	  }
	  finalURL := resp.Request.URL.String()
		finalUrlParts := strings.Split(finalURL, "/")
		version = finalUrlParts[len(finalUrlParts)-1]
	}

	// parse semver
	applicationVersion, err := semver.Make(strings.TrimLeft(appVersion, "v"))
	if err != nil {
		 log.Fatalf("Unexpected Error: %s", err)
	}
	updateTargetVersion, err := semver.Make(strings.TrimLeft(version, "v"))
	if err != nil {
		 log.Fatalf("Unexpected Error: %s", err)
	}

	if(applicationVersion.Compare(updateTargetVersion) == 0) {
		log.Info("No update available, already at the latest version!")
		return
	}

	var downloadUrl string = fmt.Sprintf("https://dl.bintray.com/%s/%s/%s/%s/envcli_%s_%s", appUpdater.BintrayOrg, appUpdater.BintrayRepository, appUpdater.BintrayPackage, version, runtime.GOOS, runtime.GOARCH)
	log.Debugf("Downloading requested version from URL: %v", downloadUrl)

	// download new version
	resp, err := http.Get(downloadUrl)
  if err != nil {
		log.Fatalf("Unexpected Error: %s", err)
  }
  defer resp.Body.Close()
  err = update.Apply(resp.Body, update.Options{})
  if err != nil {
		log.Fatalf("Unexpected Error: %s", err)
  }

	// Log Result
	if applicationVersion.GT(updateTargetVersion) {
		log.Infof("Downgraded from [%s] to [%s]!", applicationVersion.String(), updateTargetVersion.String())
	} else if applicationVersion.LT(updateTargetVersion) {
		log.Infof("Upgraded from [%s] to [%s]!", applicationVersion.String(), updateTargetVersion.String())
	}
}
