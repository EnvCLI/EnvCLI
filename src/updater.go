package main

import (
	log "github.com/sirupsen/logrus" // imports as package "log"
	"github.com/equinox-io/equinox"
)

/**
 * The Application Update Configuration
 * Using equinox.io
 */
type ApplicationUpdater struct {
	AppId  string
	PublicKey string
}

/**
 * Load the .devcli.yml Configuration
 */
func (appUpdater ApplicationUpdater) update() {
	var opts equinox.Options
  if err := opts.SetPublicKeyPEM([]byte(appUpdater.PublicKey)); err != nil {
		log.Errorf("Update failed: %s", err)
    return
  }

  // check for the update
  resp, err := equinox.Check(appUpdater.AppId, opts)
  switch {
	  case err == equinox.NotAvailableErr:
			log.Info("No update available, already at the latest version!")
	    return
	  case err != nil:
			log.Errorf("Update failed: %s", err)
	    return
  }

  // fetch the update and apply it
  err = resp.Apply()
  if err != nil {
		log.Errorf("Update failed: %s", err)
    return
  }

	log.Infof("Updated to new version: %s!", resp.ReleaseVersion)
}
