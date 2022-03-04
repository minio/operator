#!/usr/bin/env python
import subprocess

version = "v4.4.10"

template = f"""apiVersion: krew.googlecontainertools.github.com/v1alpha2
kind: Plugin
metadata:
  name: minio
spec:
  version: {version}
  homepage: https://github.com/minio/operator/tree/master/kubectl-minio
  shortDescription: Deploy and manage MinIO Operator and Tenant(s)
  description: |
    The kubectl-minio plugin wraps the MinIO Operator and provides a simplified 
    interface to create and manage MinIO tenant clusters.
  caveats: |
    * For resources that are not in default namespace, currently you must
      specify -n/--namespace explicitly (the current namespace setting is not
      yet used).
  platforms:
"""

main_url = "https://github.com/minio/operator/releases/download/{version}/kubectl-minio_{os}_{arch}.zip"

builds = {
    "darwin": [
        "amd64",
        "arm64"
    ],
    "linux": [
        "amd64",
        "arm64"
    ],
    "windows": [
        "amd64",
    ],
}

buffer = template

cmd = "curl -L {url} | sha256sum"
for os_key in builds:
    for arch in builds[os_key]:
        url = main_url.format(version=version, os=os_key, arch=arch)
        ps = subprocess.Popen(('curl', '-L', url), stdout=subprocess.PIPE)
        output = subprocess.check_output(('/usr/bin/sha256sum'), stdin=ps.stdout)
        ps.wait()
        hash = output.strip().decode("utf-8", "ignore").replace("  -", "")
        # print(hash)
        binaryext = ""
        if os_key == "windows":
            binaryext = ".exe"
        buffer += f"""  - selector:
      matchLabels:
        os: {os_key}
        arch: {arch}
    uri: https://github.com/minio/operator/releases/download/{version}/kubectl-minio_{os_key}_{arch}.zip
    sha256: {hash}
    bin: kubectl-minio{binaryext}
"""

with open("minio.yaml", "w") as f:
    f.write(buffer)
