{{- define "gvList" -}}
{{- $groupVersions := . -}}

// Generated documentation. Please do not edit.
:anchor_prefix: k8s-api

[id="{p}-api-reference"]
== API Reference

:minio-image: https://hub.docker.com/r/minio/minio/tags[minio/minio:RELEASE.2025-04-22T22-12-26Z]
:kes-image: https://hub.docker.com/r/minio/kes/tags[minio/kes:2025-03-12T09-35-18Z]
:mc-image: https://hub.docker.com/r/minio/mc/tags[minio/mc:RELEASE.2025-04-22T22-12-26Z]

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
