package updater

/**
 * The Application Update Configuration
 * Using equinox.io
 */
type ApplicationUpdater struct {
	BintrayOrg        string
	BintrayRepository string
	BintrayPackage    string
	GitHubOrg         string
	GitHubRepository  string
}
