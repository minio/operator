{{- define "gvList" -}}
{{- $groupVersions := . -}}

// Generated documentation. Please do not edit.
:anchor_prefix: k8s-api

[id="{p}-api-reference"]
== API Reference

:minio-image: https://hub.docker.com/r/minio/minio/tags[minio/minio:RELEASE.2021-07-30T00-02-00Z]
:console-image: https://hub.docker.com/r/minio/console/tags[minio/console:v0.8.2]
:kes-image: https://hub.docker.com/r/minio/kes/tags[minio/kes:v0.15.1]
:prometheus-image: https://quay.io/prometheus/prometheus:latest[prometheus/prometheus:latest]
:logsearch-image: https://hub.docker.com/r/minio/logsearchapi/tags[minio/logsearchapi:v4.1.3]
:postgres-image: https://github.com/docker-library/postgres[library/postgres]

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
