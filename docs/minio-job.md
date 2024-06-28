# MinioJob is a Kubernetes Job that runs mc commands

Requirements:
- Operator Enabled STS

Tips:
MinioJob will use `myminio` as reference tenant `ALIAS`

here is an example of a MinioJob:
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mc-job-sa
---
apiVersion: sts.min.io/v1alpha1
kind: PolicyBinding
metadata:
  name: mc-job-binding
spec:
  application:
    serviceaccount: mc-job-sa
  policies:
    - consoleAdmin
---
apiVersion: v1
kind: Secret
metadata:
  name: mytestsecret
data:
  PASSWORD: cGVkcm8xMjM= # echo pedro123 | base64
---
apiVersion: v1
kind: Secret
metadata:
  name: mytestsecretenvs
data:
  USER: ZGFuaWVs # echo daniel | base64
  PASSWORD: ZGFuaWVsMTIz # echo daniel123 | base64
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: mytestconfig
data:
  policy.json:  |
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Action": [
                    "s3:*"
                ],
                "Resource": [
                    "arn:aws:s3:::memes",
                    "arn:aws:s3:::memes/*"
                ]
            }
        ]
    }
---
apiVersion: job.min.io/v1alpha1
kind: MinIOJob
metadata:
  name: minio-test-job
spec:
#  mcImage: quay.io/minio/mc:latest
  serviceAccountName: mc-job-sa
  securityContext: {}
  containerSecurityContext: {}
  tenant:
    name: mytest-minio
  commands:
    - op: make-bucket
      args:
        name: memes
    - name: add-my-user-1
      op: admin/user/add
      args:
        user: ${USER}
        password: ${PASSWORD}
      envFrom:
        - secretRef:
            name: mytestsecretenvs
    - name: add-my-user-2
      op: admin/user/add
      args:
        user: pedro
        password: $PASSWORD
      env:
        - name: PASSWORD
          valueFrom:
            secretKeyRef:
              name: mytestsecret
              key: PASSWORD
    - name: add-my-policy
      op: admin/policy/create
      args:
        name: memes-access
        policy: /temp/policy.json
      volumeMounts:
        - name: policy
          mountPath: /temp
      volumes:
        - name: policy
          configMap:
            name: mytestconfig
            items:
              - key: policy.json
                path: policy.json
    - op: admin/policy/attach
      dependsOn:
        - add-my-user-1
        - add-my-user-2
        - add-my-policy
      args:
        policy: memes-access
        user: daniel
    - op: admin/policy/attach
      dependsOn:
        - add-my-user-1
        - add-my-user-2
        - add-my-policy
      args:
        policy: memes-access
        user: pedro
    - op: stat
      command:
        - "mc"
        - "stat"
        - "myminio/memes"
```
The MinioJob is a Kubernetes Job that runs mc commands. It uses the MinIO client (mc) to interact with the MinIO server.
## mcImage
Optional, defaults to `quay.io/minio/mc:latest`
The `mcImage` field specifies the Docker image that will be used to run the mc commands.
## serviceAccountName
The `serviceAccountName` field specifies the name of the Kubernetes ServiceAccount that will be used to run the mc commands. In this case, the ServiceAccount is `mc-job-sa`.
## securityContext
example:
```yaml
runAsUser: 1000
runAsGroup: 1000
fsGroup: 1000
fsGroupChangePolicy: "OnRootMismatch"
runAsNonRoot: true
allowPrivilegeEscalation: false
capabilities:
  drop:
    - ALL
```
The `securityContext` field specifies the security context that will be used to run the mc commands. 
## containerSecurityContext
The `containerSecurityContext` field specifies the security context that will be used to run the `mc` commands in the container.
## tenant
```yaml
name: tenantName
namespace: tenantNamespace
```
The target tenant that the job will run against.
## commands
### args
if you set this field, the `mc` command will be executed with the arguments.
`op` must be one of these:
`mb`,`make-bucket`, `admin/user/add`,`admin/policy/create`,`admin/policy/attach`, `admin/config/set`, `support/callhome`,`license/register`
```yaml
op: make-bucket
args:
  name: memes
  --with-locks: ""
```
Will do a job like `mc mb --with-locks myminio/memes`
```yaml
name: add-my-policy
op: admin/policy/create
args:
  name: memes-access
policy: /temp/policy.json
volumeMounts:
- name: policy
  mountPath: /temp
volumes:
- name: policy
  configMap:
    name: mytestconfig
    items:
      - key: policy.json
        path: policy.json
```
Will do a job like `mc admin policy create myminio memes-access /temp/policy.json`
### command
The `command` field specifies the command that will be executed by the `mc` command.
`args` must be empty. 
`op` optional, can be set to the main command name.
```
op: stat
command:
  - "mc"
  - "stat"
  - "myminio/memes"
```
or
```
command:
  - "mc"
  - "stat"
  - "myminio/memes"
```
Will do a job like `mc stat myminio/memes`
### env/envFrom/volumeMounts/volumes
The `env/envFrom/volumeMounts/volumes` fields specify the environment variables/volumes that will be used by the `mc` command
### resources
```yaml
resources:
  requests:
    cpu: "100m"
    memory: "128Mi"
  limits:
    cpu: "500m"
    memory: "256Mi"
```
The `resources` field specifies the resource requirements that will be used by the container.
### dependsOn
The `dependsOn` field specifies the commands that must be executed before the current command.