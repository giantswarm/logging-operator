package common

// ManagementClusterConfig holds configuration relevant to the management cluster environment
// where the operator is running. This configuration is typically derived from
// command-line flags at operator startup.
type ManagementClusterConfig struct {
	EnableLoggingFlag       bool
	DefaultLoggingAgent     string
	DefaultKubeEventsLogger string
	InstallationName        string
	InsecureCA              bool
}
