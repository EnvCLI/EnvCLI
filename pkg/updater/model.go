package updater

/**
 * The Application Update Configuration
 */
type ApplicationUpdater struct {
	BintrayOrg        string
	BintrayRepository string
	BintrayPackage    string
	GitHubOrg         string
	GitHubRepository  string
}
