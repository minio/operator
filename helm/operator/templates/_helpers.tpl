{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "minio-operator.name" -}}
  {{- default .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "minio-operator.fullname" -}}
  {{- $name := default .Chart.Name -}}
  {{- if contains $name .Release.Name -}}
    {{- .Release.Name | trunc 63 | trimSuffix "-" -}}
  {{- else -}}
    {{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
  {{- end -}}
{{- end -}}

{{/*
Create a default fully qualified console name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "minio-operator.console-fullname" -}}
  {{- printf "%s-%s" .Release.Name "console" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "minio-operator.chart" -}}
  {{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels for operator. Does not include selectorLabels.
*/}}
{{- define "minio-operator.labels-common" -}}
helm.sh/chart: {{ include "minio-operator.chart" . }}
{{- range $key, $val :=  .Values.operator.additionalLabels }}
{{ $key }}: {{ $val | quote }}
{{- end }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
All labels for operator. Includes selectorLabels.
*/}}
{{- define "minio-operator.labels-all" -}}
{{ include "minio-operator.labels-common" . }}
{{ include "minio-operator.selectorLabels" . }}
{{- end -}}

{{/*
Selector labels Operator
*/}}
{{- define "minio-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "minio-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Common labels for console. Does not include selectorLabels.
*/}}
{{- define "minio-operator.console-labels-common" -}}
helm.sh/chart: {{ include "minio-operator.chart" . }}
{{- range $key, $val := .Values.console.additionalLabels }}
{{ $key }}: {{ $val | quote }}
{{- end }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
All labels for console. Combines minio-operator.console-labels-common and minio-operator.console-selectorLabels.
*/}}
{{- define "minio-operator.console-labels-all" -}}
{{ include "minio-operator.console-labels-common" . }}
{{ include "minio-operator.console-selectorLabels" . }}
{{- end -}}

{{/*
Selector labels Console
*/}}
{{- define "minio-operator.console-selectorLabels" -}}
app.kubernetes.io/name: {{ include "minio-operator.name" . }}
app.kubernetes.io/instance: {{ printf "%s-%s" .Release.Name "console" }}
{{- end -}}
