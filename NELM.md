# Using nelm to deploy minio operators
This document provides instruction how (nelm)[https://github.com/werf/nelm] could be used for installing and upgrading helm charts

Main advantages of nelm are Server-Side Apply instead of 3-Way Merge, build in pod logs and status tracking for installs, diff display of install changes.

# Installing nelm
Check latest release page for installation instruction: (releases)[https://github.com/werf/nelm/releases].
Nelm can be installed directly as binary or with (tldr)[https://github.com/werf/trdl].

# Installing operators using nelm

## Installing straight from sources
1. In the root of the repo, run the following commands:
```bash
$ nelm release install -n minio -r operator ./helm/operator
```

Nelm will output progress status of created resources and containers logs:

```bash
Starting release "operator" (namespace: "minio")
┌ Progress status
│ RESOURCE (→READY)                          STATE    INFO
│ Deployment/minio-operator                  WAITING  Ready:0/2
│  • Pod/minio-operator-84867f7cd-67fhk      UNKNOWN  Status:ContainerCreating
│  • Pod/minio-operator-84867f7cd-jqhxc      UNKNOWN  Status:ContainerCreating
│ ClusterRole/minio-operator-role            READY
│ ClusterRoleBinding/minio-operator-binding  READY
│ Service/operator                           READY
│ Service/sts                                READY
│ ServiceAccount/minio-operator              READY
└ Progress status

┌ Logs for Pod/minio-operator-84867f7cd-67fhk, container/operator
│ I0602 13:57:56.451191       1 controller.go:81] Starting MinIO Operator
│ I0602 13:57:56.452231       1 main-controller.go:293] Setting up event handlers
│ I0602 13:57:56.459735       1 main-controller.go:514] Using Kubernetes CSR Version: v1
│ I0602 13:57:56.459758       1 main-controller.go:534] Waiting for STS API to start
│ I0602 13:57:56.459776       1 leaderelection.go:257] attempting to acquire leader lease minio/minio-operator-lock...
│ I0602 13:57:56.459829       1 main-controller.go:399] Starting STS API server
│ I0602 13:57:56.461263       1 tls.go:54] Waiting for the sts certificates secret to be issued
│ I0602 13:57:56.468473       1 leaderelection.go:271] successfully acquired lease minio/minio-operator-lock
│ I0602 13:57:56.468564       1 main-controller.go:577] minio-operator-84867f7cd-67fhk: I am the leader, applying leader labels on myself
│ I0602 13:57:56.468645       1 main-controller.go:428] Waiting for Upgrade Server to start
│ I0602 13:57:56.468674       1 main-controller.go:432] Starting Tenant controller
│ I0602 13:57:56.468679       1 main-controller.go:435] Waiting for informer caches to sync
│ I0602 13:57:56.468703       1 main-controller.go:388] Starting HTTP Upgrade Tenant Image server
│ I0602 13:57:56.569827       1 main-controller.go:456] STS Autocert is enabled, starting API certificate setup.
│ I0602 13:57:56.573769       1 tls.go:127] sts-tls TLS secret not found: secrets "sts-tls" not found
│ I0602 13:57:56.591452       1 csr.go:183] Start polling for certificate of csr/sts-minio-csr, every 5s, timeout after 20m0s
└ Logs for Pod/minio-operator-84867f7cd-67fhk, container/operator

┌ Logs for Pod/minio-operator-84867f7cd-jqhxc, container/operator
│ I0602 13:57:57.107739       1 controller.go:81] Starting MinIO Operator
│ I0602 13:57:57.108397       1 main-controller.go:293] Setting up event handlers
│ I0602 13:57:57.115257       1 main-controller.go:514] Using Kubernetes CSR Version: v1
│ I0602 13:57:57.115285       1 main-controller.go:534] Waiting for STS API to start
│ I0602 13:57:57.115306       1 leaderelection.go:257] attempting to acquire leader lease minio/minio-operator-lock...
│ I0602 13:57:57.115368       1 main-controller.go:399] Starting STS API server
│ I0602 13:57:57.116599       1 tls.go:54] Waiting for the sts certificates secret to be issued
│ I0602 13:57:57.116653       1 main-controller.go:595] minio-operator-84867f7cd-67fhk: is the leader, removing any leader labels that I 'minio-operator-84867f7cd-jqhxc' might have
└ Logs for Pod/minio-operator-84867f7cd-jqhxc, container/operator

┌ Progress status
│ RESOURCE (→READY)                      STATE  INFO
│ Deployment/minio-operator              READY  Ready:2/2
│  • Pod/minio-operator-84867f7cd-67fhk  READY  Status:Running
│  • Pod/minio-operator-84867f7cd-jqhxc  READY  Status:Running
└ Progress status

┌ Completed operations
│ Create resource: ClusterRole/minio-operator-role
│ Create resource: ClusterRoleBinding/minio-operator-binding
│ Create resource: CustomResourceDefinition/policybindings.sts.min.io
│ Create resource: CustomResourceDefinition/tenants.minio.min.io
│ Create resource: Deployment/minio-operator
│ Create resource: Service/operator
│ Create resource: Service/sts
│ Create resource: ServiceAccount/minio-operator
└ Completed operations

Succeeded release "operator" (namespace: "minio")
```

## Installing operator from remote index


1. Add index repo:

```bash
$ nelm repo add minio https://operator.min.io/
```

```bash
"minio" has been added to your repositories
```

2. Install release:
To install from remote chart index `REMOTE_CHART` feature needs to be enabled.
Optionally add `chart-version` parameter to install specific version.

```bash
$ NELM_FEAT_REMOTE_CHARTS=true nelm release install \
--chart-version 7.1.0 -n minio \
-r operator \
 minio/operator
```

```bash
Starting release "operator" (namespace: "minio")
┌ Progress status
│ RESOURCE (→READY)                          STATE    INFO
│ Deployment/minio-operator                  WAITING  Ready:0/2
│  • Pod/minio-operator-5b7b99967-8bv2s      UNKNOWN  Status:ContainerCreating
│  • Pod/minio-operator-5b7b99967-ksjzm      UNKNOWN  Status:ContainerCreating
│ ClusterRole/minio-operator-role            READY
│ ClusterRoleBinding/minio-operator-binding  READY
│ Service/operator                           READY
│ Service/sts                                READY
│ ServiceAccount/minio-operator              READY
└ Progress status

┌ Logs for Pod/minio-operator-5b7b99967-8bv2s, container/operator
│ I0602 14:18:28.924869       1 controller.go:81] Starting MinIO Operator
│ I0602 14:18:28.926301       1 main-controller.go:293] Setting up event handlers
│ I0602 14:18:28.935441       1 main-controller.go:514] Using Kubernetes CSR Version: v1
│ I0602 14:18:28.935472       1 main-controller.go:534] Waiting for STS API to start
│ I0602 14:18:28.935488       1 leaderelection.go:257] attempting to acquire leader lease minio/minio-operator-lock...
│ I0602 14:18:28.935627       1 main-controller.go:399] Starting STS API server
│ I0602 14:18:28.937841       1 tls.go:54] Waiting for the sts certificates secret to be issued
│ I0602 14:18:28.937905       1 main-controller.go:595] minio-operator-84867f7cd-rlcnj: is the leader, removing any leader labels that I 'minio-operator-5b7b99967-8bv2s' might have
└ Logs for Pod/minio-operator-5b7b99967-8bv2s, container/operator

┌ Logs for Pod/minio-operator-5b7b99967-ksjzm, container/operator
│ I0602 14:18:28.029927       1 controller.go:81] Starting MinIO Operator
│ I0602 14:18:28.030491       1 main-controller.go:293] Setting up event handlers
│ I0602 14:18:28.037500       1 main-controller.go:514] Using Kubernetes CSR Version: v1
│ I0602 14:18:28.037550       1 main-controller.go:534] Waiting for STS API to start
│ I0602 14:18:28.037573       1 leaderelection.go:257] attempting to acquire leader lease minio/minio-operator-lock...
│ I0602 14:18:28.037630       1 main-controller.go:399] Starting STS API server
│ I0602 14:18:28.039072       1 tls.go:54] Waiting for the sts certificates secret to be issued
│ I0602 14:18:28.039127       1 main-controller.go:595] minio-operator-84867f7cd-rlcnj: is the leader, removing any leader labels that I 'minio-operator-5b7b99967-ksjzm' might have
└ Logs for Pod/minio-operator-5b7b99967-ksjzm, container/operator

┌ Progress status
│ RESOURCE (→READY)                      STATE  INFO
│ Deployment/minio-operator              READY  Ready:2/2
│  • Pod/minio-operator-5b7b99967-ksjzm  READY  Status:Running
│  • Pod/minio-operator-5b7b99967-8bv2s  READY  Status:Running
└ Progress status

┌ Completed operations
│ Create resource: ClusterRole/minio-operator-role
│ Create resource: ClusterRoleBinding/minio-operator-binding
│ Create resource: CustomResourceDefinition/policybindings.sts.min.io
│ Create resource: CustomResourceDefinition/tenants.minio.min.io
│ Create resource: Deployment/minio-operator
│ Create resource: Service/operator
│ Create resource: Service/sts
│ Create resource: ServiceAccount/minio-operator
└ Completed operations

Succeeded release "operator" (namespace: "minio")
```

3. Display installed release:

```bash
$ nelm release list -n minio
```

```bash
NAME            NAMESPACE       REVISION        UPDATED                                         STATUS          CHART           APP VERSION
operator        minio           1               2025-06-02 16:18:21.332122514 +0200 CEST        deployed        operator-7.1.0  v7.1.0
```
## Upgrading

### Patch version example

1. Before upgrading release you can display changes that will be applied this showcases patch version upgrade.

```bash
$ NELM_FEAT_REMOTE_CHARTS=true nelm release plan install  -n minio -r operator --chart-version 7.1.1 minio/operator
```

```bash
┌ Update ClusterRole/minio-operator-role
│     annotations: {}
│     labels:
│       app.kubernetes.io/managed-by: Helm
│ -     app.kubernetes.io/version: v7.1.0
│ -     helm.sh/chart: operator-7.1.0
│ +     app.kubernetes.io/version: v7.1.1
│ +     helm.sh/chart: operator-7.1.1
│     name: minio-operator-role
│   rules:
│   - apiGroups:
└ Update ClusterRole/minio-operator-role

┌ Update ClusterRoleBinding/minio-operator-binding
│     annotations: {}
│     labels:
│       app.kubernetes.io/managed-by: Helm
│ -     app.kubernetes.io/version: v7.1.0
│ -     helm.sh/chart: operator-7.1.0
│ +     app.kubernetes.io/version: v7.1.1
│ +     helm.sh/chart: operator-7.1.1
│     name: minio-operator-binding
│   roleRef:
│     apiGroup: rbac.authorization.k8s.io
└ Update ClusterRoleBinding/minio-operator-binding

┌ Update CustomResourceDefinition/policybindings.sts.min.io
│   metadata:
│     annotations:
│       controller-gen.kubebuilder.io/version: v0.17.2
│ -     operator.min.io/version: v5.0.15
│ +     operator.min.io/version: v7.1.1
│     labels:
│       app.kubernetes.io/managed-by: Helm
│     name: policybindings.sts.min.io
└ Update CustomResourceDefinition/policybindings.sts.min.io

┌ Update CustomResourceDefinition/tenants.minio.min.io
│   metadata:
│     annotations:
│       controller-gen.kubebuilder.io/version: v0.17.2
│ -     operator.min.io/version: v7.1.0
│ +     operator.min.io/version: v7.1.1
│     labels:
│       app.kubernetes.io/managed-by: Helm
│     name: tenants.minio.min.io
└ Update CustomResourceDefinition/tenants.minio.min.io

┌ Update Deployment/minio-operator
│       deployment.kubernetes.io/revision: "1"
│     labels:
│       app.kubernetes.io/managed-by: Helm
│ -     app.kubernetes.io/version: v7.1.0
│ -     helm.sh/chart: operator-7.1.0
│ +     app.kubernetes.io/version: v7.1.1
│ +     helm.sh/chart: operator-7.1.1
│     name: minio-operator
│     namespace: minio
│   spec:
│    ...
│           app.kubernetes.io/instance: operator
│           app.kubernetes.io/managed-by: Helm
│           app.kubernetes.io/name: operator
│ -         app.kubernetes.io/version: v7.1.0
│ -         helm.sh/chart: operator-7.1.0
│ +         app.kubernetes.io/version: v7.1.1
│ +         helm.sh/chart: operator-7.1.1
│       spec:
│         affinity:
│           podAntiAffinity:
│    ...
│           env:
│           - name: OPERATOR_STS_ENABLED
│             value: "on"
│ -         image: quay.io/minio/operator:v7.1.0
│ +         image: quay.io/minio/operator:v7.1.1
│           imagePullPolicy: IfNotPresent
│           name: operator
│           resources:
└ Update Deployment/minio-operator

┌ Update Service/operator
│     annotations: {}
│     labels:
│       app.kubernetes.io/managed-by: Helm
│ -     app.kubernetes.io/version: v7.1.0
│ -     helm.sh/chart: operator-7.1.0
│ +     app.kubernetes.io/version: v7.1.1
│ +     helm.sh/chart: operator-7.1.1
│     name: operator
│     namespace: minio
│   spec:
└ Update Service/operator

┌ Update Service/sts
│     annotations: {}
│     labels:
│       app.kubernetes.io/managed-by: Helm
│ -     app.kubernetes.io/version: v7.1.0
│ -     helm.sh/chart: operator-7.1.0
│ +     app.kubernetes.io/version: v7.1.1
│ +     helm.sh/chart: operator-7.1.1
│     name: sts
│     namespace: minio
│   spec:
└ Update Service/sts

┌ Update ServiceAccount/minio-operator
│     annotations: {}
│     labels:
│       app.kubernetes.io/managed-by: Helm
│ -     app.kubernetes.io/version: v7.1.0
│ -     helm.sh/chart: operator-7.1.0
│ +     app.kubernetes.io/version: v7.1.1
│ +     helm.sh/chart: operator-7.1.1
│     name: minio-operator
│     namespace: minio
└ Update ServiceAccount/minio-operator

Planned changes summary for release "operator" (namespace: "minio"):
- update: 8 resource(s)
```

2. Upgrade to newer version:

```bash
$ NELM_FEAT_REMOTE_CHARTS=true nelm release install -n minio -r operator --chart-version 7.1.1 minio/operator
```

```bash
Starting release "operator" (namespace: "minio")
┌ Logs for Pod/minio-operator-84867f7cd-fc87z, container/operator
│ I0602 15:05:20.063616       1 controller.go:81] Starting MinIO Operator
│ I0602 15:05:20.064459       1 main-controller.go:293] Setting up event handlers
│ I0602 15:05:20.071184       1 main-controller.go:514] Using Kubernetes CSR Version: v1
│ I0602 15:05:20.071239       1 main-controller.go:534] Waiting for STS API to start
│ I0602 15:05:20.071283       1 leaderelection.go:257] attempting to acquire leader lease minio/minio-operator-lock...
│ I0602 15:05:20.071338       1 main-controller.go:399] Starting STS API server
│ I0602 15:05:20.072908       1 main-controller.go:595] minio-operator-5b7b99967-ksjzm: is the leader, removing any leader labels that I 'minio-operator-84867f7cd-fc87z' might have
└ Logs for Pod/minio-operator-84867f7cd-fc87z, container/operator

┌ Logs for Pod/minio-operator-84867f7cd-rx2bl, container/operator
│ I0602 15:05:21.689566       1 controller.go:81] Starting MinIO Operator
│ I0602 15:05:21.690547       1 main-controller.go:293] Setting up event handlers
│ I0602 15:05:21.697783       1 main-controller.go:514] Using Kubernetes CSR Version: v1
│ I0602 15:05:21.697805       1 main-controller.go:534] Waiting for STS API to start
│ I0602 15:05:21.697824       1 leaderelection.go:257] attempting to acquire leader lease minio/minio-operator-lock...
│ I0602 15:05:21.697880       1 main-controller.go:399] Starting STS API server
│ I0602 15:05:21.707643       1 leaderelection.go:271] successfully acquired lease minio/minio-operator-lock
│ I0602 15:05:21.707713       1 main-controller.go:577] minio-operator-84867f7cd-rx2bl: I am the leader, applying leader labels on myself
│ I0602 15:05:21.707776       1 main-controller.go:428] Waiting for Upgrade Server to start
│ I0602 15:05:21.707787       1 main-controller.go:432] Starting Tenant controller
│ I0602 15:05:21.707791       1 main-controller.go:435] Waiting for informer caches to sync
│ I0602 15:05:21.707814       1 main-controller.go:456] STS Autocert is enabled, starting API certificate setup.
│ I0602 15:05:21.707834       1 main-controller.go:388] Starting HTTP Upgrade Tenant Image server
└ Logs for Pod/minio-operator-84867f7cd-rx2bl, container/operator

┌ Progress status
│ RESOURCE (→READY)                          STATE  INFO
│ Deployment/minio-operator                  READY  Ready:4/2
│  • Pod/minio-operator-84867f7cd-rx2bl      READY  Status:Running
│  • Pod/minio-operator-5b7b99967-8bv2s      READY
│  • Pod/minio-operator-5b7b99967-ksjzm      READY
│  • Pod/minio-operator-84867f7cd-fc87z      READY  Status:Running
│ ClusterRole/minio-operator-role            READY
│ ClusterRoleBinding/minio-operator-binding  READY
│ Service/operator                           READY
│ Service/sts                                READY
│ ServiceAccount/minio-operator              READY
└ Progress status

┌ Completed operations
│ Update resource: ClusterRole/minio-operator-role
│ Update resource: ClusterRoleBinding/minio-operator-binding
│ Update resource: CustomResourceDefinition/policybindings.sts.min.io
│ Update resource: CustomResourceDefinition/tenants.minio.min.io
│ Update resource: Deployment/minio-operator
│ Update resource: Service/operator
│ Update resource: Service/sts
│ Update resource: ServiceAccount/minio-operator
└ Completed operations

Succeeded release "operator" (namespace: "minio")
```

# Uninstall
```bash
$ nelm release uninstall -n minio -r operator
```

```bash
Deleting release "operator" (namespace: "minio")
┌ Waiting for resources elimination: services/sts, services/operator, deployments/minio-operator, clusterrolebindings/minio-operator-binding, clusterroles/minio-operator-role, customresourcedefinitions/policybindings.sts.min.io, customresourcedefinitions/tenants.minio.min.io, serviceaccounts/minio-operator
└ Waiting for resources elimination: services/sts, services/operator, deployments/minio-operator, clusterrolebindings/minio-operator-binding, clusterroles/minio-operator-role, customresourcedefinitions/policybindings.sts.min.io, customresourcedefinitions/tenants.minio.min.io, serviceaccounts/minio-op ... (0.21 seconds)

release "operator" uninstalled
Uninstalled release "operator" (namespace: "minio")
```
