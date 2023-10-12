{{- define "gvList" -}}
{{- $groupVersions := . -}}

// Generated documentation. Please do not edit.
:anchor_prefix: k8s-api

[id="{p}-api-reference"]
== API Reference

:minio-image: https://hub.docker.com/r/minio/minio/tags[minio/minio:RELEASE.2023-10-07T15-07-38Z]
:kes-image: https://hub.docker.com/r/minio/kes/tags[minio/kes:2023-10-03T00-48-37Z]

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
