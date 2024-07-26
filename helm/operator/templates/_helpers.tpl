{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "minio-operator.name" -}}
  {{- default .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "minio-operator.chart" -}}
  {{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels for operator
*/}}
{{- define "minio-operator.labels" -}}
helm.sh/chart: {{ include "minio-operator.chart" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- range $key, $val :=  .Values.operator.additionalLabels }}
{{ $key }}: {{ $val | quote }}
{{- end }}
{{- end -}}

{{/*
Selector labels Operator
*/}}
{{- define "minio-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "minio-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
