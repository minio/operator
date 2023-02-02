package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEnsureDefaults(t *testing.T) {
	mt := Tenant{}
	mt.EnsureDefaults()

	t.Run("defaults", func(t *testing.T) {
		assert.Equal(t, mt.Spec.Image, DefaultMinIOImage)
		assert.Equal(t, mt.Spec.Mountpath, MinIOVolumeMountPath)
		assert.Equal(t, mt.Spec.Subpath, MinIOVolumeSubPath)
		// default behavior is autoCert will be enabled
		assert.True(t, mt.AutoCert())
		require.True(t, mt.HasCertConfig())
		require.NotNil(t, mt.Spec.CertConfig)
	})

	t.Run("enable and disable autoCert", func(t *testing.T) {
		// disable autoCert explicitly
		autoCertEnabled := false
		mt.Spec.RequestAutoCert = &autoCertEnabled

		mt.EnsureDefaults()

		assert.False(t, mt.AutoCert())
		assert.True(t, mt.HasCertConfig())
		require.NotNil(t, mt.Spec.CertConfig)

		// enable autoCert explicitly
		autoCertEnabled = true
		mt.Spec.RequestAutoCert = &autoCertEnabled

		mt.EnsureDefaults()

		assert.True(t, mt.AutoCert())
		assert.True(t, mt.HasCertConfig())
		require.NotNil(t, mt.Spec.CertConfig)
	})

	t.Run("defaults don't override", func(t *testing.T) {
		newImage := "mtnio/mtnio:latest"
		mt.Spec.Image = newImage
		mt.EnsureDefaults()

		assert.Equal(t, newImage, mt.Spec.Image)
	})
}

func TestTemplateVariables(t *testing.T) {
	servers := 2
	mt := Tenant{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: TenantSpec{
			Pools: []Pool{
				{
					Name:                "single",
					Servers:             int32(servers),
					VolumesPerServer:    4,
					VolumeClaimTemplate: nil,
					Resources:           corev1.ResourceRequirements{},
					NodeSelector:        nil,
					Affinity:            nil,
					Tolerations:         nil,
				},
			},
		},
	}
	mt.EnsureDefaults()

	t.Run("StatefulSet", func(t *testing.T) {
		hosts := mt.TemplatedMinIOHosts("{{.StatefulSet}}")
		assert.Contains(t, hosts, mt.MinIOStatefulSetNameForPool(&mt.Spec.Pools[0]))
	})

	t.Run("CIService", func(t *testing.T) {
		hosts := mt.TemplatedMinIOHosts("{{.CIService}}")
		assert.Contains(t, hosts, mt.MinIOCIServiceName())
	})

	t.Run("HLService", func(t *testing.T) {
		hosts := mt.TemplatedMinIOHosts("{{.HLService}}")
		assert.Contains(t, hosts, mt.MinIOHLServiceName())
	})

	t.Run("Ellipsis", func(t *testing.T) {
		hosts := mt.TemplatedMinIOHosts("{{.Ellipsis}}")
		assert.Contains(t, hosts, genEllipsis(0, servers-1))
	})

	t.Run("Domain", func(t *testing.T) {
		hosts := mt.TemplatedMinIOHosts("{{.Domain}}")
		assert.Contains(t, hosts, GetClusterDomain())
	})
}

func TestTenant_KESServiceEndpoint(t1 *testing.T) {
	type fields struct {
		TypeMeta   metav1.TypeMeta
		ObjectMeta metav1.ObjectMeta
		Scheduler  TenantScheduler
		Spec       TenantSpec
		Status     TenantStatus
	}
	autoCertEnabled := true
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Success",
			fields: fields{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kes",
					Namespace: "namespace",
				},
				Spec: TenantSpec{
					RequestAutoCert: &autoCertEnabled,
				},
			},
			want: "https://kes" + KESHLSvcNameSuffix + ".namespace.svc.cluster.local:7373",
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tenant{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Scheduler:  tt.fields.Scheduler,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if got := t.KESServiceEndpoint(); got != tt.want {
				t1.Errorf("KESServiceEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareEnvs(t *testing.T) {
	type args struct {
		old map[string]string
		new map[string]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal",
			args: args{
				old: map[string]string{
					"TEST_ENV1": "test_val1",
					"TEST_ENV2": "test_val2",
				},
				new: map[string]string{
					"TEST_ENV1": "test_val1",
					"TEST_ENV2": "test_val2",
				},
			},
			want: false,
		},
		{
			name: "not_sorted",
			args: args{
				old: map[string]string{
					"TEST_ENV2": "test_val2",
					"TEST_ENV1": "test_val1",
				},
				new: map[string]string{
					"TEST_ENV1": "test_val1",
					"TEST_ENV2": "test_val2",
				},
			},
			want: false,
		},
		{
			name: "unequal_length",
			args: args{
				old: map[string]string{
					"TEST_ENV1": "test_val1",
					"TEST_ENV2": "test_val2",
				},
				new: map[string]string{
					"TEST_ENV1": "test_val1",
					"TEST_ENV2": "test_val2",
					"TEST_ENV3": "test_val3",
				},
			},
			want: true,
		},
		{
			name: "unequal_values",
			args: args{
				old: map[string]string{
					"TEST_ENV1": "test_val1",
					"TEST_ENV2": "test_val2",
				},
				new: map[string]string{
					"TEST_ENV1": "test_val1",
					"TEST_ENV2": "test_val3",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEnvUpdated(tt.args.old, tt.args.new); got != tt.want {
				t.Errorf("Test case = %s, CompareEnvs() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestParseRawConfiguration(t *testing.T) {
	type args struct {
		configuration []byte
	}
	tests := []struct {
		name       string
		args       args
		wantConfig map[string][]byte
	}{
		{
			name: "ldap-configuration",
			args: args{
				configuration: []byte(`
export MINIO_ROOT_USER=minio
export MINIO_ROOT_PASSWORD=minio123

export MINIO_IDENTITY_LDAP_SERVER_ADDR=localhost:389
export MINIO_IDENTITY_LDAP_LOOKUP_BIND_DN="cn=admin,dc=min,dc=io"
export MINIO_IDENTITY_LDAP_LOOKUP_BIND_PASSWORD="admin"
export MINIO_IDENTITY_LDAP_USER_DN_SEARCH_BASE_DN="dc=min,dc=io"
export MINIO_IDENTITY_LDAP_USER_DN_SEARCH_FILTER="(uid=%s)"
export MINIO_IDENTITY_LDAP_GROUP_SEARCH_BASE_DN="ou=swengg,dc=min,dc=io"
export MINIO_IDENTITY_LDAP_GROUP_SEARCH_FILTER="(&(objectclass=groupOfNames)(member=%d))"
export MINIO_IDENTITY_LDAP_SERVER_INSECURE="on"

export MINIO_BROWSER_REDIRECT_URL=http://localhost:9001
export MINIO_SERVER_URL=http://localhost:9000`),
			},
			wantConfig: map[string][]byte{
				"accesskey":                                  []byte("minio"),
				"secretkey":                                  []byte("minio123"),
				"MINIO_ROOT_USER":                            []byte("minio"),
				"MINIO_ROOT_PASSWORD":                        []byte("minio123"),
				"MINIO_IDENTITY_LDAP_SERVER_ADDR":            []byte("localhost:389"),
				"MINIO_IDENTITY_LDAP_LOOKUP_BIND_DN":         []byte("cn=admin,dc=min,dc=io"),
				"MINIO_IDENTITY_LDAP_LOOKUP_BIND_PASSWORD":   []byte("admin"),
				"MINIO_IDENTITY_LDAP_USER_DN_SEARCH_BASE_DN": []byte("dc=min,dc=io"),
				"MINIO_IDENTITY_LDAP_USER_DN_SEARCH_FILTER":  []byte("(uid=%s)"),
				"MINIO_IDENTITY_LDAP_GROUP_SEARCH_BASE_DN":   []byte("ou=swengg,dc=min,dc=io"),
				"MINIO_IDENTITY_LDAP_GROUP_SEARCH_FILTER":    []byte("(&(objectclass=groupOfNames)(member=%d))"),
				"MINIO_IDENTITY_LDAP_SERVER_INSECURE":        []byte("on"),
				"MINIO_BROWSER_REDIRECT_URL":                 []byte("http://localhost:9001"),
				"MINIO_SERVER_URL":                           []byte("http://localhost:9000"),
			},
		},
		{
			name: "oidc-configuration",
			args: args{
				configuration: []byte(`
#!/bin/bash

# // Auth0 Rule
# function (user, context, callback) {
#   const namespace = 'https://min.io/';
#   context.accessToken[namespace + 'policy'] = 'mcsAdmin';
#   context.idToken[namespace + 'policy'] = 'mcsAdmin';
#   callback(null, user, context);
# }

export MINIO_ROOT_USER=minio
export MINIO_ROOT_PASSWORD=minio123
export MINIO_IDENTITY_OPENID_CONFIG_URL=https://*******************/.well-known/openid-configuration
export MINIO_IDENTITY_OPENID_CLIENT_ID="****************************"
export MINIO_IDENTITY_OPENID_CLIENT_SECRET="********************"
export MINIO_IDENTITY_OPENID_SCOPES="openid,profile,email"
export MINIO_IDENTITY_OPENID_CLAIM_NAME="https://min.io/policy"
export MINIO_BROWSER_REDIRECT_URL=http://localhost:9001
export MINIO_SERVER_URL=http://localhost:9000
./minio server ~/Data --console-address ":9001"`),
			},
			wantConfig: map[string][]byte{
				"accesskey":                           []byte("minio"),
				"secretkey":                           []byte("minio123"),
				"MINIO_ROOT_USER":                     []byte("minio"),
				"MINIO_ROOT_PASSWORD":                 []byte("minio123"),
				"MINIO_IDENTITY_OPENID_CONFIG_URL":    []byte("https://*******************/.well-known/openid-configuration"),
				"MINIO_IDENTITY_OPENID_CLIENT_ID":     []byte("****************************"),
				"MINIO_IDENTITY_OPENID_CLIENT_SECRET": []byte("********************"),
				"MINIO_IDENTITY_OPENID_SCOPES":        []byte("openid,profile,email"),
				"MINIO_IDENTITY_OPENID_CLAIM_NAME":    []byte("https://min.io/policy"),
				"MINIO_BROWSER_REDIRECT_URL":          []byte("http://localhost:9001"),
				"MINIO_SERVER_URL":                    []byte("http://localhost:9000"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedValues := ParseRawConfiguration(tt.args.configuration)
			assert.Equalf(t, len(tt.wantConfig), len(parsedValues), "ParseRawConfiguration(%v)", string(tt.args.configuration))
			for k := range tt.wantConfig {
				assert.Equalf(t, string(tt.wantConfig[k]), string(parsedValues[k]), "ParseRawConfiguration(%v)", string(tt.args.configuration))
			}
		})
	}
}

func TestTenant_GetDomainHosts(t1 *testing.T) {
	type fields struct {
		TypeMeta   metav1.TypeMeta
		ObjectMeta metav1.ObjectMeta
		Scheduler  TenantScheduler
		Spec       TenantSpec
		Status     TenantStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "List of domains to host list",
			fields: fields{
				Spec: TenantSpec{
					Features: &Features{
						Domains: &TenantDomains{
							Minio: []string{
								"domain1.com:8080",
								"domain2.com",
							},
						},
					},
				},
			},
			want: []string{
				"domain1.com",
				"domain2.com",
			},
		},
		{
			name: "Empty hosts",
			fields: fields{
				Spec: TenantSpec{
					Features: &Features{},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tenant{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Scheduler:  tt.fields.Scheduler,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			assert.Equalf(t1, tt.want, t.GetDomainHosts(), "GetDomainHosts()")
		})
	}
}

func TestTenant_HasEnv(t1 *testing.T) {
	type fields struct {
		TypeMeta   metav1.TypeMeta
		ObjectMeta metav1.ObjectMeta
		Scheduler  TenantScheduler
		Spec       TenantSpec
		Status     TenantStatus
	}
	type args struct {
		envName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Contains env",
			fields: fields{
				Spec: TenantSpec{
					Env: []corev1.EnvVar{
						{
							Name:  "ENV1",
							Value: "whatever",
						},
					},
				},
			},
			args: args{
				envName: "ENV1",
			},
			want: true,
		},
		{
			name: "Does not Contains env",
			fields: fields{
				Spec: TenantSpec{
					Env: []corev1.EnvVar{
						{
							Name:  "ENV1",
							Value: "whatever",
						},
					},
				},
			},
			args: args{
				envName: "ENV2",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tenant{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Scheduler:  tt.fields.Scheduler,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			assert.Equalf(t1, tt.want, t.HasEnv(tt.args.envName), "HasEnv(%v)", tt.args.envName)
		})
	}
}

func TestTenant_ValidateDomains(t1 *testing.T) {
	type fields struct {
		TypeMeta   metav1.TypeMeta
		ObjectMeta metav1.ObjectMeta
		Scheduler  TenantScheduler
		Spec       TenantSpec
		Status     TenantStatus
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Valid Domains",
			fields: fields{
				Spec: TenantSpec{
					Features: &Features{
						Domains: &TenantDomains{
							Minio: []string{
								"domain1.com:8080",
								"domain2.com",
								"http://domain3.com:8080",
								"https://domain4.com",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid Domains",
			fields: fields{
				Spec: TenantSpec{
					Features: &Features{
						Domains: &TenantDomains{
							Minio: []string{
								"http s://domain1.com:8080",
								"http://domain2.com",
								"httx://domain3.com",
								":8080",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Duplicate Domains",
			fields: fields{
				Spec: TenantSpec{
					Features: &Features{
						Domains: &TenantDomains{
							Minio: []string{
								"domain2.com",
								"other.domain2.com:8080",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tenant{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Scheduler:  tt.fields.Scheduler,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if tt.wantErr {
				if err := t.ValidateDomains(); err == nil {
					assert.Failf(t1, "Test %s did not return error", tt.name)
				}
			}
		})
	}
}
