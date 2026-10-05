{{/*
Expand the name of the chart.
*/}}
{{- define "event-horizon.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "event-horizon.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "event-horizon.labels" -}}
app.kubernetes.io/name: {{ include "event-horizon.name" . }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app: event-horizon
{{- end }}

{{- define "event-horizon.selectorLabels" -}}
app: event-horizon
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "event-horizon.secretName" -}}
{{- .Values.secret.name }}
{{- end }}

{{- define "event-horizon.image" -}}
{{- printf "%s/%s:%s" .registry .name .tag -}}
{{- end }}
