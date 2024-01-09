{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "labels.common" -}}
app: {{ include "name" . | quote }}
{{ include "labels.selector" . }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
application.giantswarm.io/team: {{ index .Chart.Annotations "application.giantswarm.io/team" | default "atlas" | quote }}
helm.sh/chart: {{ include "chart" . | quote }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "labels.selector" -}}
app.kubernetes.io/name: {{ include "name" . | quote }}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
{{- end -}}


{{/*
Profiling annotations
*/}}
{{- define "annotations.profiling" -}}
profiles.grafana.com/memory.scrape: "true"
profiles.grafana.com/memory.port: {{ .Values.profiling.port | quote }}
profiles.grafana.com/cpu.scrape: "true"
profiles.grafana.com/cpu.port: {{ .Values.profiling.port | quote }}
profiles.grafana.com/goroutine.scrape: "true"
profiles.grafana.com/goroutine.port: {{ .Values.profiling.port | quote }}
{{- end -}}
