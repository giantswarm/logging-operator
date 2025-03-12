package loggedcluster

type LoggingAgent struct {
	LoggingAgent     string
	KubeEventsLogger string
}

func (la *LoggingAgent) GetLoggingAgent() string {
	return la.LoggingAgent
}

func (la *LoggingAgent) SetLoggingAgent(loggingAgent string) {
	la.LoggingAgent = loggingAgent
}

func (la *LoggingAgent) GetKubeEventsLogger() string {
	return la.KubeEventsLogger
}

func (la *LoggingAgent) SetKubeEventsLogger(kubeEventsLogger string) {
	la.KubeEventsLogger = kubeEventsLogger
}
