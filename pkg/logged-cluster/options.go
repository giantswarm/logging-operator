package loggedcluster

// Options to be used for any loggedCluster
type Options struct {
	EnableLoggingFlag      bool
	InstallationName       string
	InstallationProvider   string
	InstallationRegion     string
	InstallationBaseDomain string
}

// O blah
var O = Options{}
