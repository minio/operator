// Copyright (C) 2020, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package v2

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Tenant is a https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/[Kubernetes object] describing a MinIO Tenant. +
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:object:generate=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=tenant,singular=tenant
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.currentState"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:storageversion
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Scheduler TenantScheduler `json:"scheduler,omitempty"`
	// *Required* +
	//
	// The root field for the MinIO Tenant object.
	Spec TenantSpec `json:"spec"`
	// Status provides details of the state of the Tenant
	// +optional
	Status TenantStatus `json:"status"`
}

// TenantScheduler (`scheduler`) - Object describing Kubernetes Scheduler to use for deploying the MinIO Tenant.
type TenantScheduler struct {
	// *Optional* +
	//
	// Specify the name of the https://kubernetes.io/docs/concepts/scheduling-eviction/kube-scheduler/[Kubernetes scheduler] to be used to schedule Tenant pods
	Name string `json:"name"`
}

// Bucket describes the default created buckets
type Bucket struct {
	Name          string `json:"name,omitempty"`
	Region        string `json:"region,omitempty"`
	ObjectLocking bool   `json:"objectLock,omitempty"`
}

// TenantDomains (`domains`) - List of domains used to access the tenant from outside the kubernetes clusters.
// this will only configure MinIO for the domains listed, but external DNS configuration is still needed.
// The listed domains should include schema and port if any is used, i.e. https://minio.domain.com:8123
type TenantDomains struct {
	// List of Domains used by MinIO. This will enable DNS style access to the object store where the bucket name is
	// inferred from a subdomain in the domain.
	Minio []string `json:"minio,omitempty"`
	// Domain used to expose the MinIO Console, this will configure the redirect on MinIO when visiting from the browser
	// If Console is exposed via a subpath, the domain should include it, i.e. https://console.domain.com:8123/subpath/
	Console string `json:"console,omitempty"`
}

// Features (`features`) - Object describing which MinIO features to enable/disable in the MinIO Tenant. +
type Features struct {
	// *Optional* +
	//
	// Specify `true` to allow clients to access buckets using the DNS path `<bucket>.minio.default.svc.cluster.local`. Defaults to `false`.
	//
	BucketDNS bool `json:"bucketDNS,omitempty"`
	// *Optional* +
	//
	// Specify a list of domains used to access MinIO and Console.
	//
	Domains *TenantDomains `json:"domains,omitempty"`
	// *Optional* +
	//
	// Starts minio server with SFTP support
	EnableSFTP *bool `json:"enableSFTP,omitempty"`
}

// TenantSpec (`spec`) defines the configuration of a MinIO Tenant object. +
//
// The following parameters are specific to the `minio.min.io/v2` MinIO CRD API `spec` definition added as part of the MinIO Operator v4.0.0. +
//
// For more complete documentation on this object, see the https://min.io/docs/minio/kubernetes/upstream/operations/installation.html[MinIO Kubernetes Documentation]. +
type TenantSpec struct {
	// *Required* +
	//
	// An array of objects describing each MinIO server pool deployed in the MinIO Tenant. Each pool consists of a set of MinIO server pods which "pool" their storage resources for supporting object storage and retrieval requests. Each server pool is independent of all others and supports horizontal scaling of available storage resources in the MinIO Tenant. +
	//
	// The MinIO Tenant `spec` *must have* at least *one* element in the `pools` array. +
	//
	// See the https://min.io/docs/minio/kubernetes/upstream/operations/install-deploy-manage/deploy-minio-tenant.html[MinIO Operator CRD] reference for the `pools` object for examples and more complete documentation.
	Pools []Pool `json:"pools"`
	// *Optional* +
	//
	// The Docker image to use when deploying `minio` server pods. Defaults to {minio-image}. +
	//
	// +optional
	Image string `json:"image,omitempty"`
	// *Optional* +
	//
	// Specify the secret key to use for pulling images from a private Docker repository. +
	// +optional
	ImagePullSecret corev1.LocalObjectReference `json:"imagePullSecret,omitempty"`
	// *Optional* +
	//
	// Pod Management Policy for pod created by StatefulSet
	// +optional
	PodManagementPolicy appsv1.PodManagementPolicyType `json:"podManagementPolicy,omitempty"`
	// *optional* +
	//
	// Specify a https://kubernetes.io/docs/concepts/configuration/secret/[Kubernetes opaque secret] to use for setting the MinIO root access key and secret key. Specify the secret as `name: <secret>`. The Kubernetes secret must contain the following fields: +
	//
	// * `data.accesskey` - The access key for the root credentials +
	//
	// * `data.secretkey` - The secret key for the root credentials +
	//
	//
	// +optional
	CredsSecret *corev1.LocalObjectReference `json:"credsSecret,omitempty"`
	// *Optional* +
	//
	// If provided, the MinIO Operator adds the specified environment variables when deploying the Tenant resource.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// *Optional* +
	//
	// Enables TLS with SNI support on each MinIO pod in the tenant. If `externalCertSecret` is omitted *and* `requestAutoCert` is set to `false`, the MinIO Tenant deploys *without* TLS enabled. +
	//
	// Specify an array of https://kubernetes.io/docs/concepts/configuration/secret/[Kubernetes TLS secrets]. The MinIO Operator copies the specified certificates to every MinIO server pod in the tenant. When the MinIO pod/service responds to a TLS connection request, it uses SNI to select the certificate with matching `subjectAlternativeName`. +
	//
	// Each element in the `externalCertSecret` array is an object containing the following fields: +
	//
	// * - `name` - The name of the Kubernetes secret containing the TLS certificate. +
	//
	// * - `type` - Specify `kubernetes.io/tls` +
	//
	// See the https://min.io/docs/minio/kubernetes/upstream/operations/install-deploy-manage/deploy-minio-tenant.html#create-tenant-security-section[MinIO Operator CRD] reference for examples and more complete documentation on configuring TLS for MinIO Tenants.
	// +optional
	ExternalCertSecret []*LocalCertificateReference `json:"externalCertSecret,omitempty"`
	// *Optional* +
	//
	// Allows MinIO server pods to verify client TLS certificates signed by a Certificate Authority not in the pod's trust store. +
	//
	// Specify an array of https://kubernetes.io/docs/concepts/configuration/secret/[Kubernetes TLS secrets]. The MinIO Operator copies the specified certificates to every MinIO server pod in the tenant. +
	//
	// Each element in the `externalCertSecret` array is an object containing the following fields: +
	//
	// * - `name` - The name of the Kubernetes secret containing the Certificate Authority. +
	//
	// * - `type` - Specify `kubernetes.io/tls`. +
	//
	// See the https://min.io/docs/minio/kubernetes/upstream/operations/install-deploy-manage/deploy-minio-tenant.html#create-tenant-security-section[MinIO Operator CRD] reference for examples and more complete documentation on configuring TLS for MinIO Tenants.
	// +optional
	ExternalCaCertSecret []*LocalCertificateReference `json:"externalCaCertSecret,omitempty"`
	// *Optional* +
	//
	// Enables mTLS authentication between the MinIO Tenant pods and https://github.com/minio/kes[MinIO KES]. *Required* for enabling connectivity between the MinIO Tenant and MinIO KES. +
	//
	// Specify a https://kubernetes.io/docs/concepts/configuration/secret/[Kubernetes TLS secrets]. The MinIO Operator copies the specified certificate to every MinIO server pod in the tenant. The secret *must* contain the following fields: +
	//
	// * `name` - The name of the Kubernetes secret containing the TLS certificate. +
	//
	// * `type` - Specify `kubernetes.io/tls` +
	//
	// The specified certificate *must* correspond to an identity on the KES server. See the https://github.com/minio/kes/wiki/Configuration#policy-configuration[KES Wiki] for more information on KES identities. +
	//
	// If deploying KES with the MinIO Operator, include the hash of the certificate as part of the <<k8s-api-github-com-minio-operator-pkg-apis-minio-min-io-v2-kesconfig,`kes`>> object specification. +
	//
	// See the https://min.io/docs/minio/kubernetes/upstream/operations/install-deploy-manage/deploy-minio-tenant.html#create-tenant-security-section[MinIO Operator CRD] reference for examples and more complete documentation on configuring TLS for MinIO Tenants.
	//
	// +optional
	ExternalClientCertSecret *LocalCertificateReference `json:"externalClientCertSecret,omitempty"`
	// *Optional* +
	//
	// Provide support for mounting additional client certificate into MinIO Tenant pods
	// Multiple client certificates will be mounted using the following folder structure:
	//
	//	certs
	//		|
	//		+ client-0
	//		|			+ client.crt
	//		|			+ client.key
	//		+ client-1
	//		|			+ client.crt
	//		|			+ client.key
	//		+ client-2
	//		|			+ client.crt
	//		|			+ client.key
	//
	// Specify a https://kubernetes.io/docs/concepts/configuration/secret/[Kubernetes TLS secrets]. The MinIO Operator copies the specified certificate to every MinIO server pod in the tenant that later can be referenced using environment variables. The secret *must* contain the following fields: +
	//
	// * `name` - The name of the Kubernetes secret containing the TLS certificate. +
	//
	// * `type` - Specify `kubernetes.io/tls` +
	//
	// +optional
	ExternalClientCertSecrets []*LocalCertificateReference `json:"externalClientCertSecrets,omitempty"`
	// *Optional* +
	//
	// Mount path for MinIO volume (PV). Defaults to `/export`
	// +optional
	Mountpath string `json:"mountPath,omitempty"`
	// *Optional* +
	//
	// Subpath inside mount path. This is the directory where MinIO stores data. Default to `""`` (empty)
	// +optional
	Subpath string `json:"subPath,omitempty"`
	// *Optional* +
	//
	// Enables using https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/[Kubernetes-based TLS certificate generation] and signing for pods and services in the MinIO Tenant. +
	//
	// * Specify `true` to explicitly enable automatic certificate generate (Default). +
	//
	// * Specify `false` to disable automatic certificate generation. +
	//
	// If `requestAutoCert` is set to `false` *and* `externalCertSecret` is omitted, the MinIO Tenant deploys *without* TLS enabled.
	//
	// See the https://min.io/docs/minio/kubernetes/upstream/operations/install-deploy-manage/deploy-minio-tenant.html#create-tenant-security-section[MinIO Operator CRD] reference for examples and more complete documentation on configuring TLS for MinIO Tenants.
	// +optional
	RequestAutoCert *bool `json:"requestAutoCert,omitempty"`

	// Liveness Probe for container liveness. Container will be restarted if the probe fails.
	// +optional
	Liveness *corev1.Probe `json:"liveness,omitempty"`

	// Readiness Probe for container readiness. Container will be removed from service endpoints if the probe fails.
	// +optional
	Readiness *corev1.Probe `json:"readiness,omitempty"`

	// Startup Probe allows to configure a max grace period for a pod to start before getting traffic routed to it.
	// +optional
	Startup *corev1.Probe `json:"startup,omitempty"`

	// S3 related features can be disabled or enabled such as `bucketDNS` etc.
	Features *Features `json:"features,omitempty"`
	// *Optional* +
	//
	// Enables setting the `CommonName`, `Organization`, and `dnsName` attributes for all TLS certificates automatically generated by the Operator. Configuring this object has no effect if `requestAutoCert` is `false`. +
	// +optional
	CertConfig *CertificateConfig `json:"certConfig,omitempty"`
	// *Optional* +
	//
	// Directs the MinIO Operator to deploy the https://github.com/minio/kes[MinIO Key Encryption Service] (KES) using the specified configuration. The MinIO KES supports performing server-side encryption of objects on the MiNIO Tenant. +
	//
	//
	//+optional
	KES *KESConfig `json:"kes,omitempty"`
	// *Optional* +
	//
	// Directs the MinIO Operator to use prometheus operator. +
	//
	// Tenant scrape configuration will be added to prometheus managed by the prometheus-operator.
	//+optional
	PrometheusOperator bool `json:"prometheusOperator,omitempty"`
	// *Optional* +
	//
	// The https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/[Kubernetes Service Account] to use for running MinIO pods created as part of the Tenant. +
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// *Optional* +
	//
	// Indicates the Pod priority and therefore importance of a Pod relative to other Pods in the cluster.
	// This is applied to MinIO pods only. +
	//
	// Refer Kubernetes https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/#priorityclass[Priority Class documentation] for more complete documentation.
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// *Optional* +
	//
	// The pull policy for the MinIO Docker image. Specify one of the following: +
	//
	// * `Always` +
	//
	// * `Never` +
	//
	// * `IfNotPresent` (Default) +
	//
	// Refer Kubernetes documentation for details https://kubernetes.io/docs/concepts/containers/images#updating-images
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// *Optional* +
	//
	// A list of containers to run as sidecars along every MinIO Pod deployed in the tenant.
	// +optional
	SideCars *SideCars `json:"sideCars,omitempty"`
	// *Optional* +
	//
	// Directs the Operator to expose the MinIO and/or Console services. +
	// +optional
	ExposeServices *ExposeServices `json:"exposeServices,omitempty"`
	// *Optional* +
	//
	// Specify custom labels and annotations to append to the MinIO service and/or Console service.
	// +optional
	ServiceMetadata *ServiceMetadata `json:"serviceMetadata,omitempty"`
	// *Optional* +
	//
	// An array of https://kubernetes.io/docs/concepts/configuration/secret/[Kubernetes opaque secrets] to use for generating MinIO users during tenant provisioning. +
	//
	// Each element in the array is an object consisting of a key-value pair `name: <string>`, where the `<string>` references an opaque Kubernetes secret. +
	//
	// Each referenced Kubernetes secret must include the following fields: +
	//
	// * `CONSOLE_ACCESS_KEY` - The "Username" for the MinIO user +
	//
	// * `CONSOLE_SECRET_KEY` - The "Password" for the MinIO user +
	//
	// The Operator creates each user with the `consoleAdmin` policy by default. You can change the assigned policy after the Tenant starts. +
	// +optional
	Users []*corev1.LocalObjectReference `json:"users,omitempty"`
	// *Optional* +
	//
	// Create buckets when creating a new tenant. Skip if bucket with given name already exists
	// +optional
	Buckets []Bucket `json:"buckets,omitempty"`
	// *Optional* +
	//
	// Enable JSON, Anonymous logging for MinIO tenants.
	// +optional
	Logging *Logging `json:"logging,omitempty"`
	// *Optional* +
	//
	// Specify a secret that contains additional environment variable configurations to be used for the MinIO pools.
	// The secret is expected to have a key named config.env containing all exported environment variables for MinIO+
	// +optional
	Configuration *corev1.LocalObjectReference `json:"configuration,omitempty"`
	// *Optional* +
	//
	// Add customs initContainers to StatefulSet
	// +optional
	InitContainers []corev1.Container `json:"initContainers,omitempty"`
}

// Logging describes Logging for MinIO tenants.
type Logging struct {
	JSON      bool `json:"json,omitempty"`
	Anonymous bool `json:"anonymous,omitempty"`
	Quiet     bool `json:"quiet,omitempty"`
}

// ServiceMetadata (`serviceMetadata`) defines custom labels and annotations for the MinIO Object Storage service and/or MinIO Console service. +
type ServiceMetadata struct {
	// *Optional* +
	//
	// If provided, append these labels to the MinIO service
	// +optional
	MinIOServiceLabels map[string]string `json:"minioServiceLabels,omitempty"`
	// *Optional* +
	//
	// If provided, append these annotations to the MinIO service
	// +optional
	MinIOServiceAnnotations map[string]string `json:"minioServiceAnnotations,omitempty"`
	// *Optional* +
	//
	// If provided, append these labels to the Console service
	// +optional
	ConsoleServiceLabels map[string]string `json:"consoleServiceLabels,omitempty"`
	// *Optional* +
	//
	// If provided, append these annotations to the Console service
	// +optional
	ConsoleServiceAnnotations map[string]string `json:"consoleServiceAnnotations,omitempty"`
}

// LocalCertificateReference (`externalCertSecret`, `externalCaCertSecret`,`clientCertSecret`) contains a Kubernetes secret containing TLS certificates or Certificate Authority files for use with enabling TLS in the MinIO Tenant. +
type LocalCertificateReference struct {
	// *Required* +
	//
	// The name of the Kubernetes secret containing the TLS certificate or Certificate Authority file. +
	Name string `json:"name"`
	// *Required* +
	//
	// The type of Kubernetes secret. Specify `kubernetes.io/tls` +
	Type string `json:"type,omitempty"`
}

// ExposeServices (`exposeServices`) defines the exposure of the MinIO object storage and Console services. +
type ExposeServices struct {
	// *Optional* +
	//
	// Directs the Operator to expose the MinIO service. Defaults to `true`. +
	// +optional
	MinIO bool `json:"minio,omitempty"`
	// *Optional* +
	//
	// Directs the Operator to expose the MinIO Console service. Defaults to `true`. +
	// +optional
	Console bool `json:"console,omitempty"`
}

// CertificateStatus keeps track of all the certificates managed by the operator
type CertificateStatus struct {
	// AutoCertEnabled registers whether we know if the tenant has autocert enabled
	// +nullable
	AutoCertEnabled *bool `json:"autoCertEnabled,omitempty"`
	// Provides the output of the `client`, `minio`, and`minioCAs` custom TLS certificates manually added to the Operator.
	// +nullable
	CustomCertificates *CustomCertificates `json:"customCertificates,omitempty"`
}

// PoolState represents the state of a pool
type PoolState string

const (
	// PoolNotCreated of a pool when it's not even created yet
	PoolNotCreated PoolState = "PoolNotCreated"
	// PoolCreated indicates a pool was created
	PoolCreated PoolState = "PoolCreated"
	// PoolInitialized indicates if a pool has been observed to be online
	PoolInitialized PoolState = "PoolInitialized"
)

// PoolStatus keeps track of all the pools and their current state
type PoolStatus struct {
	SSName string    `json:"ssName"`
	State  PoolState `json:"state"`
	// LegacySecurityContext stands for Legacy SecurityContext. It represents that these pool was created before v4.2.3 when
	// we introduced the default securityContext as non-root, thus we should keep running this Pool without a
	// Security Context
	// +optional
	LegacySecurityContext bool `json:"legacySecurityContext"`
}

// HealthStatus represents whether the tenant is healthy, with decreased service or offline
type HealthStatus string

const (
	// HealthStatusGreen indicates a healthy tenant: all drives online
	HealthStatusGreen HealthStatus = "green"
	// HealthStatusYellow indicates a decreased resilience tenant, some drives offline
	HealthStatusYellow HealthStatus = "yellow"
	// HealthStatusRed indicates the tenant is offline, or lost write quorum
	HealthStatusRed HealthStatus = "red"
)

// TierUsage represents the usage from a tier setup by the tenant
type TierUsage struct {
	// Name of the tier
	Name string `json:"Name"`
	// type of the tier
	Type string `json:"Type,omitempty"`
	// TotalSize usage of the tier
	TotalSize int64 `json:"totalSize"`
}

// TenantUsage are metrics regarding the usage and capacity of the tenant
type TenantUsage struct {
	// Capacity the usage capacity of this tenant in bytes.
	// +optional
	Capacity int64 `json:"capacity,omitempty"`
	// Capacity the raw capacity of this tenant in bytes.
	// +optional
	RawCapacity int64 `json:"rawCapacity,omitempty"`
	// Usage is how much data is managed by MinIO in bytes.
	// +optional
	Usage int64 `json:"usage,omitempty"`
	// Usage is the raw usage on disks in bytes.
	// +optional
	RawUsage int64 `json:"rawUsage,omitempty"`
	// Tiers includes the usage of individual tiers in the tenant
	// +optional
	Tiers []TierUsage `json:"tiers,omitempty"`
}

// TenantStatus is the status for a Tenant resource
type TenantStatus struct {
	CurrentState      string `json:"currentState"`
	AvailableReplicas int32  `json:"availableReplicas"`
	Revision          int32  `json:"revision"`
	SyncVersion       string `json:"syncVersion"`
	// Keeps track of all the TLS certificates managed by the operator
	// +nullable
	Certificates CertificateStatus `json:"certificates"`
	// All the pools get an individual status
	// +nullable
	Pools []PoolStatus `json:"pools"`
	// *Optional* +
	//
	// Minimum number of disks that need to be online
	WriteQuorum int32 `json:"writeQuorum,omitempty"`
	// *Optional* +
	//
	// Total number of drives online for the tenant
	DrivesOnline int32 `json:"drivesOnline,omitempty"`
	// *Optional* +
	//
	// Total number of drives offline
	DrivesOffline int32 `json:"drivesOffline,omitempty"`
	// *Optional* +
	//
	// Drives with healing going on
	DrivesHealing int32 `json:"drivesHealing,omitempty"`
	// *Optional* +
	//
	// Health State of the tenant
	HealthStatus HealthStatus `json:"healthStatus,omitempty"`
	// *Optional* +
	//
	// Health Message regarding the State of the tenant
	HealthMessage string `json:"healthMessage,omitempty"`
	// *Optional* +
	//
	// If set, we will wait until cleared for up a given time
	WaitingOnReady *metav1.Time `json:"waitingOnReady,omitempty"`
	// *Optional* +
	//
	// Information about tenant usage
	Usage TenantUsage `json:"usage,omitempty"`
	// Health Message regarding the State of the tenant
	// ProvisionedUsers keeps track for telling if operator already created initial users for the tenant
	ProvisionedUsers bool `json:"provisionedUsers,omitempty"`
	// *Optional* +
	//
	// Health Message regarding the State of the tenant
	// ProvisionedBuckets keeps track for telling if operator already created initial buckets for the tenant
	ProvisionedBuckets bool `json:"provisionedBuckets,omitempty"`
}

// CertificateConfig (`certConfig`) defines controlling attributes associated to any TLS certificate automatically generated by the Operator as part of tenant creation. These fields have no effect if `spec.autoCert: false`.
type CertificateConfig struct {
	// *Optional* +
	//
	// The `CommonName` or `CN` attribute to associate to automatically generated TLS certificates. +
	CommonName string `json:"commonName,omitempty"`
	// *Optional* +
	//
	// Specify one or more `OrganizationName` or `O` attributes to associate to automatically generated TLS certificates. +
	OrganizationName []string `json:"organizationName,omitempty"`
	// *Optional* +
	//
	// Specify one or more x.509 Subject Alternative Names (SAN) to associate to automatically generated TLS certificates. MinIO Server pods use SNI to determine which certificate to respond with based on the requested hostname.
	DNSNames []string `json:"dnsNames,omitempty"`
}

// CustomCertificates (`customCertificates`) provides groupings of the TLS certificates manually added to the Operator as part of tenant creation. These fields contain no data if there are no custom TLS certificates.
type CustomCertificates struct {
	// *Optional* +
	//
	// Client
	Client []*CustomCertificateConfig `json:"client,omitempty"`
	// *Optional* +
	//
	// Minio
	Minio []*CustomCertificateConfig `json:"minio,omitempty"`
	// *Optional* +
	//
	// Certificate Authorities
	MinioCAs []*CustomCertificateConfig `json:"minioCAs,omitempty"`
}

// CustomCertificateConfig (`customCertificateConfig`) provides attributes associated of the TLS certificates manually added to the Operator as part of tenant creation. These fields contain no data if there are no custom TLS certificates.
type CustomCertificateConfig struct {
	// *Optional* +
	//
	// Output one or more `CertName` attributes associated with the manually provided TLS certificates. +
	CertName string `json:"certName,omitempty"`
	// *Optional* +
	//
	// Output one or more `Domains` attributes associated with the manually provided TLS certificates. +
	Domains []string `json:"domains,omitempty"`
	// *Optional* +
	//
	// Output one or more `Expiry` attributes associated with the manually provided TLS certificates. +
	Expiry string `json:"expiry,omitempty"`
	// *Optional* +
	//
	// Output one or more `ExpiresIn` attributes associated with the manually provided TLS certificates. +
	ExpiresIn string `json:"expiresIn,omitempty"`
	// *Optional* +
	//
	// Output one or more `SerialNo` attributes associated with the manually provided TLS certificates. +
	SerialNo string `json:"serialNo,omitempty"`
}

// Pool (`pools`) defines a MinIO server pool on a Tenant. Each pool consists of a set of MinIO server pods which "pool" their storage resources for supporting object storage and retrieval requests. Each server pool is independent of all others and supports horizontal scaling of available storage resources in the MinIO Tenant. +
//
// See the https://min.io/docs/minio/kubernetes/upstream/operations/install-deploy-manage/deploy-minio-tenant.html#procedure-command-line[MinIO Operator CRD] reference for the `pools` object for examples and more complete documentation. +
type Pool struct {
	// *Optional* +
	//
	// Specify the name of the pool. The Operator automatically generates the pool name if this field is omitted.
	// +optional
	Name string `json:"name,omitempty"`
	// *Required*
	//
	// The number of MinIO server pods to deploy in the pool. The minimum value is `2`.
	//
	// The MinIO Operator requires a minimum of `4` volumes per pool. Specifically, the result of `pools.servers X pools.volumesPerServer` must be greater than `4`. +
	Servers int32 `json:"servers"`
	// *Required* +
	//
	// The number of Persistent Volume Claims to generate for each MinIO server pod in the pool. +
	//
	// The MinIO Operator requires a minimum of `4` volumes per pool. Specifically, the result of `pools.servers X pools.volumesPerServer` must be greater than `4`. +
	VolumesPerServer int32 `json:"volumesPerServer"`
	// *Required* +
	//
	// Specify the configuration options for the MinIO Operator to use when generating Persistent Volume Claims for the MinIO tenant. +
	//
	VolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"volumeClaimTemplate"`
	// *Optional* +
	//
	// Object specification for specifying CPU and memory https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/[resource allocations] or limits in the MinIO tenant. +
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// *Optional* +
	//
	// The filter for the Operator to apply when selecting which nodes on which to deploy pods in the pool. The Operator only selects those nodes whose labels match the specified selector. +
	//
	// See the Kubernetes documentation on https://kubernetes.io/docs/concepts/configuration/assign-pod-node/[Assigning Pods to Nodes] for more information.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// *Optional* +
	//
	// Specify node affinity, pod affinity, and pod anti-affinity for pods in the MinIO pool. +
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// *Optional* +
	//
	// Specify one or more https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/[Kubernetes tolerations] to apply to pods deployed in the MinIO pool.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// *Optional* +
	//
	// Specify one or more https://kubernetes.io/docs/concepts/workloads/pods/pod-topology-spread-constraints/[Kubernetes Topology Spread Constraints] to apply to pods deployed in the MinIO pool.
	// +optional
	TopologySpreadConstraints []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// *Optional* +
	//
	// Specify the https://kubernetes.io/docs/tasks/configure-pod-container/security-context/[Security Context] of pods in the pool. The Operator supports only the following pod security fields: +
	//
	// * `fsGroup` +
	//
	// * `fsGroupChangePolicy` +
	//
	// * `runAsGroup` +
	//
	// * `runAsNonRoot` +
	//
	// * `runAsUser` +
	//
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	// Specify the https://kubernetes.io/docs/tasks/configure-pod-container/security-context/[Security Context] of containers in the pool. The Operator supports only the following container security fields: +
	//
	// * `runAsGroup` +
	//
	// * `runAsNonRoot` +
	//
	// * `runAsUser` +
	//
	// +optional
	ContainerSecurityContext *corev1.SecurityContext `json:"containerSecurityContext,omitempty"`
	// *Optional* +
	//
	// Specify custom labels and annotations to append to the Pool.
	// +optional
	// *Optional* +
	//
	// If provided, use these annotations for the Pool Objects Meta annotations (Statefulset and Pod template)
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// *Optional* +
	//
	// If provided, use these labels for the Pool Objects Meta annotations (Statefulset and Pod template)
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// *Optional* +
	//
	// If provided, each pod on the Statefulset will run with the specified RuntimeClassName, for more info https://kubernetes.io/docs/concepts/containers/runtime-class/
	// +optional
	RuntimeClassName *string `json:"runtimeClassName,omitempty"`
}

// EqualImage returns true if config image and current input image are same
func (c *KESConfig) EqualImage(currentImage string) bool {
	if c == nil {
		return false
	}
	return c.Image == currentImage
}

// AuditConfig defines configuration parameters for Audit (type) logs
type AuditConfig struct {
	// *Required* +
	//
	// Specify the amount of storage to request in Gigabytes (GB) for storing audit logs.
	// +optional
	DiskCapacityGB *int `json:"diskCapacityGB,omitempty"`
}

// KESConfig (`kes`) defines the configuration of the https://github.com/minio/kes[MinIO Key Encryption Service] (KES) StatefulSet deployed as part of the MinIO Tenant. KES supports Server-Side Encryption of objects using an external Key Management Service (KMS). +
type KESConfig struct {
	// *Optional* +
	//
	// Specify the number of replica KES pods to deploy in the tenant. Defaults to `2`.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// *Optional* +
	//
	// The Docker image to use for deploying MinIO KES. Defaults to {kes-image}. +
	// +optional
	Image string `json:"image,omitempty"`
	// *Optional* +
	//
	// The pull policy for the MinIO Docker image. Specify one of the following: +
	//
	// * `Always` +
	//
	// * `Never` +
	//
	// * `IfNotPresent` (Default) +
	//
	// Refer to the Kubernetes documentation for details https://kubernetes.io/docs/concepts/containers/images#updating-images
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// *Optional* +
	//
	// The https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/[Kubernetes Service Account] to use for running MinIO KES pods created as part of the Tenant. +
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// *Required* +
	//
	// Specify a https://kubernetes.io/docs/concepts/configuration/secret/[Kubernetes opaque secret] which contains environment variables to use for setting up the MinIO KES service. +
	//
	// See the https://github.com/minio/operator/blob/master/examples/kes-secret.yaml[MinIO Operator `console-secret.yaml`] for an example.
	Configuration *corev1.LocalObjectReference `json:"kesSecret"`
	// *Optional* +
	//
	// Enables TLS with SNI support on each MinIO KES pod in the tenant. If `externalCertSecret` is omitted *and* `spec.requestAutoCert` is set to `false`, MinIO KES pods deploy *without* TLS enabled. +
	//
	// Specify a https://kubernetes.io/docs/concepts/configuration/secret/[Kubernetes TLS secret]. The MinIO Operator copies the specified certificate to every MinIO pod in the tenant. When the MinIO pod/service responds to a TLS connection request, it uses SNI to select the certificate with matching `subjectAlternativeName`. +
	//
	// Specify an object containing the following fields: +
	//
	// * - `name` - The name of the Kubernetes secret containing the TLS certificate. +
	//
	// * - `type` - Specify `kubernetes.io/tls` +
	//
	// See the https://min.io/docs/minio/kubernetes/upstream/operations/install-deploy-manage/deploy-minio-tenant.html#procedure-command-line[MinIO Operator CRD] reference for examples and more complete documentation on configuring TLS for MinIO Tenants.
	// +optional
	ExternalCertSecret *LocalCertificateReference `json:"externalCertSecret,omitempty"`
	// *Optional* +
	//
	// Specify a a https://kubernetes.io/docs/concepts/configuration/secret/[Kubernetes TLS secret] containing a custom root Certificate Authority and x.509 certificate to use for performing mTLS authentication with an external Key Management Service, such as Hashicorp Vault. +
	//
	// Specify an object containing the following fields: +
	//
	// * - `name` - The name of the Kubernetes secret containing the Certificate Authority and x.509 Certificate. +
	//
	// * - `type` - Specify `kubernetes.io/tls` +
	// +optional
	ClientCertSecret *LocalCertificateReference `json:"clientCertSecret,omitempty"`
	// *Optional* +
	//
	//  Specify the GCP default credentials to be used for KES to authenticate to GCP key store
	//
	// +optional
	GCPCredentialSecretName string `json:"gcpCredentialSecretName,omitempty"`
	// *Optional* +
	//
	//  Specify the name of the workload identity pool (This is required for generating service account token)
	//
	// +optional
	GCPWorkloadIdentityPool string `json:"gcpWorkloadIdentityPool,omitempty"`
	// *Optional* +
	//
	// If provided, use these annotations for KES Object Meta annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// *Optional* +
	//
	// If provided, use these labels for KES Object Meta labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// *Optional* +
	//
	// Object specification for specifying CPU and memory https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/[resource allocations] or limits in the MinIO tenant. +
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// *Optional* +
	//
	// The filter for the Operator to apply when selecting which nodes on which to deploy MinIO KES pods. The Operator only selects those nodes whose labels match the specified selector. +
	//
	// See the Kubernetes documentation on https://kubernetes.io/docs/concepts/configuration/assign-pod-node/[Assigning Pods to Nodes] for more information.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// *Optional* +
	//
	// Specify one or more https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/[Kubernetes tolerations] to apply to MinIO KES pods.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// *Optional* +
	//
	// Specify node affinity, pod affinity, and pod anti-affinity for the KES pods. +
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// *Optional* +
	//
	// Specify one or more https://kubernetes.io/docs/concepts/workloads/pods/pod-topology-spread-constraints/[Kubernetes Topology Spread Constraints] to apply to pods deployed in the MinIO pool.
	// +optional
	TopologySpreadConstraints []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	// *Optional* +
	//
	// If provided, use this as the name of the key that KES creates on the KMS backend
	// +optional
	KeyName string `json:"keyName,omitempty"`
	// Specify the https://kubernetes.io/docs/tasks/configure-pod-container/security-context/[Security Context] of MinIO KES pods. The Operator supports only the following pod security fields: +
	//
	// * `fsGroup` +
	//
	// * `fsGroupChangePolicy` +
	//
	// * `runAsGroup` +
	//
	// * `runAsNonRoot` +
	//
	// * `runAsUser` +
	//
	// * `seLinuxOptions` +
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	// *Optional* +
	//
	// If provided, the MinIO Operator adds the specified environment variables when deploying the KES resource.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TenantList is a list of Tenant resources
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Tenant `json:"items"`
}

// SideCars (`sidecars`) defines a list of containers that the Operator attaches to each MinIO server pods in the `pool`.
type SideCars struct {
	// *Optional* +
	//
	// List of containers to run inside the Pod
	// +patchMergeKey=name
	// +patchStrategy=merge
	Containers []corev1.Container `json:"containers" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=containers"`
	// *Optional* +
	//
	// volumeClaimTemplates is a list of claims that pods are allowed to reference.
	// The StatefulSet controller is responsible for mapping network identities to
	// claims in a way that maintains the identity of a pod. Every claim in
	// this list must have at least one matching (by name) volumeMount in one
	// container in the template. A claim in this list takes precedence over
	// any volumes in the template, with the same name.
	// +TODO: Define the behavior if a claim already exists with the same name.
	// +optional
	VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty" protobuf:"bytes,4,rep,name=volumeClaimTemplates"`
	// *Optional* +
	//
	// List of volumes that can be mounted by containers belonging to the pod.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Volumes []corev1.Volume `json:"volumes,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name" protobuf:"bytes,1,rep,name=volumes"`
}
