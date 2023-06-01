package project

var (
	description = "The logging-operator does something."
	gitSHA      = "n/a"
	name        = "logging-operator"
	source      = "https://github.com/giantswarm/logging-operator"
	version     = "0.1.0-dev"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
