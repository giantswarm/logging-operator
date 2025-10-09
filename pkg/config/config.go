package config

// Config holds the global configuration for the logging operator
// This replaces the loggedcluster.Options struct
type Config struct {
	EnableLoggingFlag       bool
	EnableTracingFlag       bool
	DefaultLoggingAgent     string
	DefaultKubeEventsLogger string
	InstallationName        string
	InsecureCA              bool
	// Static cluster configuration (from command line flags)
	Customer string
	Pipeline string
	Region   string
}
