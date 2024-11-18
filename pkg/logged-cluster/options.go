package loggedcluster

// Options to be used for any loggedCluster
type Options struct {
	EnableLoggingFlag bool
	LoggingAgent      string
	KubeEventsLogger  string
	InstallationName  string
	InsecureCA        bool
}

// O blah
var O = Options{}
