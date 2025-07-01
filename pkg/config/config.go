package config

// Config holds the global configuration for the logging operator
// This replaces the loggedcluster.Options struct
type Config struct {
	EnableLoggingFlag       bool
	DefaultLoggingAgent     string
	DefaultKubeEventsLogger string
	InstallationName        string
	InsecureCA              bool
}

// NewConfig creates a new Config instance
func NewConfig(enableLogging bool, loggingAgent, eventsLogger, installationName string, insecureCA bool) *Config {
	return &Config{
		EnableLoggingFlag:       enableLogging,
		DefaultLoggingAgent:     loggingAgent,
		DefaultKubeEventsLogger: eventsLogger,
		InstallationName:        installationName,
		InsecureCA:              insecureCA,
	}
}
