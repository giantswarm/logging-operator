package config

// Config holds the global configuration for the logging operator
// This replaces the loggedcluster.Options struct
type Config struct {
	EnableLoggingFlag           bool
	EnableNodeFilteringFlag     bool
	EnableTracingFlag           bool
	EnableNetworkMonitoringFlag bool
	InstallationName            string
	Customer                    string
	Pipeline                    string
	Region                      string
	InsecureCA                  bool
}
