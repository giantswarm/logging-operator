package config

// Config holds the global configuration for the logging operator
// This replaces the loggedcluster.Options struct
type Config struct {
	EnableLoggingFlag bool
	EnableTracingFlag bool
	InstallationName  string
	InsecureCA        bool
}
