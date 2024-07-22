{{- define "gvList" -}}
{{- $groupVersions := . -}}

// Generated documentation. Please do not edit.
:anchor_prefix: k8s-api

[id="{p}-api-reference"]
== API Reference

:minio-image: https://hub.docker.com/r/minio/minio/tags[minio/minio:RELEASE.2024-07-16T23-46-41Z]
:kes-image: https://hub.docker.com/r/minio/kes/tags[minio/kes:2024-06-17T15-47-05Z]
:mc-image: https://hub.docker.com/r/minio/mc/tags[minio/mc:RELEASE.2024-07-16T23-46-41Z]

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
