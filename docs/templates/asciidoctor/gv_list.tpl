{{- define "gvList" -}}
{{- $groupVersions := . -}}

// Generated documentation. Please do not edit.
:anchor_prefix: k8s-api

[id="{p}-api-reference"]
== API Reference

:minio-image: https://hub.docker.com/r/minio/minio/tags[minio/minio:RELEASE.2024-08-17T01-24-54Z]
:kes-image: https://hub.docker.com/r/minio/kes/tags[minio/kes:2024-08-16T14-39-28Z]
:mc-image: https://hub.docker.com/r/minio/mc/tags[minio/mc:RELEASE.2024-08-17T01-24-54Z]

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
