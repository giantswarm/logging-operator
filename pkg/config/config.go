package config

// Config holds the global configuration for the logging operator
// This replaces the loggedcluster.Options struct
type Config struct {
	EnableLoggingFlag       bool
	EnableNodeFilteringFlag bool
	EnableTracingFlag       bool
	InstallationName        string
	InsecureCA              bool
}

// NewConfig creates a new Config instance
func NewConfig(enableLogging, enableNodeFiltering, enableTracing bool, installationName string, insecureCA bool) *Config {
	return &Config{
		EnableLoggingFlag:       enableLogging,
		EnableNodeFilteringFlag: enableNodeFiltering,
		EnableTracingFlag:       enableTracing,
		InstallationName:        installationName,
		InsecureCA:              insecureCA,
	}
}
