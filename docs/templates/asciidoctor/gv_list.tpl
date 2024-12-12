{{- define "gvList" -}}
{{- $groupVersions := . -}}

// Generated documentation. Please do not edit.
:anchor_prefix: k8s-api

[id="{p}-api-reference"]
== API Reference

:minio-image: https://hub.docker.com/r/minio/minio/tags[minio/minio:RELEASE.2024-11-07T00-52-20Z]
:kes-image: https://hub.docker.com/r/minio/kes/tags[minio/kes:2024-11-25T13-44-31Z]
:mc-image: https://hub.docker.com/r/minio/mc/tags[minio/mc:RELEASE.2024-10-02T08-27-28Z]

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
