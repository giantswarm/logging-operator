package loggedcluster

// Options to be used for any loggedCluster
type Options struct {
	EnableLoggingFlag bool
	InstallationName  string
	InsecureCA        bool
}

// O blah
var O = Options{}
