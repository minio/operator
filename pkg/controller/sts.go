package controller

import (
	"context"
	"encoding/xml"
	"errors"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/minio/madmin-go/v2"
	"github.com/minio/minio-go/v7/pkg/credentials"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	xhttp "github.com/minio/operator/pkg/internal"
	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// STS Handler constants
const (
	webIdentity         = "AssumeRoleWithWebIdentity"
	stsAPIVersion       = "2011-06-15"
	stsVersion          = "Version"
	stsAction           = "Action"
	stsPolicy           = "Policy"
	stsWebIdentityToken = "WebIdentityToken"
	stsDurationSeconds  = "DurationSeconds"
	AmzRequestID        = "x-amz-request-id"
	// stsRoleArn          = "RoleArn"
)

// STS API constants
const (
	STSDefaultPort = "4223"
	STSEndpoint    = "/sts"
)

const (
	// STSEnabled Env variable name to turn on and off the STS Service is enabled, disabled by default
	STSEnabled = "OPERATOR_STS_ENABLED"

	// STSTLSSecretName is the name of secret created for the Operator STS TLS certs
	STSTLSSecretName = "sts-tls"
)

type contextKeyType string

// Error codes, non exhaustive list - http://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRoleWithSAML.html
const (
	ErrSTSNone STSErrorCode = iota
	ErrSTSAccessDenied
	ErrSTSInvalidIdentityToken
	ErrSTSMissingParameter
	ErrSTSInvalidParameterValue
	ErrSTSWebIdentityExpiredToken
	ErrSTSClientGrantsExpiredToken
	ErrSTSInvalidClientGrantsToken
	ErrSTSMalformedPolicyDocument
	ErrSTSInsecureConnection
	ErrSTSInvalidClientCertificate
	ErrSTSNotInitialized
	ErrSTSUpstreamError
	ErrSTSInternalError
	ErrSTSIDPCommunicationError
	ErrSTSPackedPolicyTooLarge
)

type stsErrorCodeMap map[STSErrorCode]APIError

// ReqInfo stores the request info.
// Reading/writing directly to struct requires appropriate R/W lock.
type ReqInfo struct {
	RemoteHost      string // Client Host/IP
	Host            string // Node Host/IP
	UserAgent       string // User Agent
	RequestID       string // x-amz-request-id
	API             string // API name
	AccessKey       string // Access Key
	TenantNamespace string // tenant namespace
	sync.RWMutex
}

// Credentials holds access and secret keys.
type Credentials struct {
	AccessKey    string                 `xml:"AccessKeyId" json:"accessKey,omitempty"`
	SecretKey    string                 `xml:"SecretAccessKey" json:"secretKey,omitempty"`
	Expiration   time.Time              `xml:"Expiration" json:"expiration,omitempty"`
	SessionToken string                 `xml:"SessionToken" json:"sessionToken,omitempty"`
	Status       string                 `xml:"-" json:"status,omitempty"`
	ParentUser   string                 `xml:"-" json:"parentUser,omitempty"`
	Groups       []string               `xml:"-" json:"groups,omitempty"`
	Claims       map[string]interface{} `xml:"-" json:"claims,omitempty"`
}

// STSErrorCode type of error status.
type STSErrorCode int

// APIError structure
type APIError struct {
	Code           string
	Description    string
	HTTPStatusCode int
}

// STSErrorResponse - error response format
type STSErrorResponse struct {
	XMLName xml.Name `xml:"https://sts.amazonaws.com/doc/2011-06-15/ ErrorResponse" json:"-"`
	Error   struct {
		Type    string `xml:"Type"`
		Code    string `xml:"Code"`
		Message string `xml:"Message"`
	} `xml:"Error"`
	RequestID string `xml:"RequestId"`
}

// error code to STSError structure, these fields carry respective
// descriptions for all the error responses.
var stsErrCodes = stsErrorCodeMap{
	ErrSTSAccessDenied: {
		Code:           "AccessDenied",
		Description:    "Generating temporary credentials not allowed for this request.",
		HTTPStatusCode: http.StatusForbidden,
	},
	ErrSTSInvalidIdentityToken: {
		Code:           "InvalidIdentityToken",
		Description:    "The web identity token that was passed could not be validated. Get a new identity token from the identity provider and then retry the request.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrSTSMissingParameter: {
		Code:           "MissingParameter",
		Description:    "A required parameter for the specified action is not supplied.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrSTSInvalidParameterValue: {
		Code:           "InvalidParameterValue",
		Description:    "An invalid or out-of-range value was supplied for the input parameter.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrSTSWebIdentityExpiredToken: {
		Code:           "ExpiredToken",
		Description:    "The web identity token that was passed is expired or is not valid. Get a new identity token from the identity provider and then retry the request.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrSTSClientGrantsExpiredToken: {
		Code:           "ExpiredToken",
		Description:    "The client grants that was passed is expired or is not valid. Get a new client grants token from the identity provider and then retry the request.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrSTSInvalidClientGrantsToken: {
		Code:           "InvalidClientGrantsToken",
		Description:    "The client grants token that was passed could not be validated by MinIO.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrSTSMalformedPolicyDocument: {
		Code:           "MalformedPolicyDocument",
		Description:    "The request was rejected because the policy document was malformed.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrSTSInsecureConnection: {
		Code:           "InsecureConnection",
		Description:    "The request was made over a plain HTTP connection. A TLS connection is required.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrSTSInvalidClientCertificate: {
		Code:           "InvalidClientCertificate",
		Description:    "The provided client certificate is invalid. Retry with a different certificate.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrSTSNotInitialized: {
		Code:           "STSNotInitialized",
		Description:    "STS API not initialized, please try again.",
		HTTPStatusCode: http.StatusServiceUnavailable,
	},
	ErrSTSUpstreamError: {
		Code:           "InternalError",
		Description:    "An upstream service required for this operation failed - please try again or contact an administrator.",
		HTTPStatusCode: http.StatusInternalServerError,
	},
	ErrSTSInternalError: {
		Code:           "InternalError",
		Description:    "We encountered an internal error generating credentials, please try again.",
		HTTPStatusCode: http.StatusInternalServerError,
	},
	ErrSTSIDPCommunicationError: {
		Code:           "IDPCommunicationError",
		Description:    "The request could not be fulfilled because the identity provider (IDP) that was asked to verify the incoming identity token could not be reached.",
		HTTPStatusCode: http.StatusBadRequest,
	},
	ErrSTSPackedPolicyTooLarge: {
		Code:           "PackedPolicyTooLarge",
		Description:    "The request was rejected because the total packed size of the session policies and session tags combined was too large",
		HTTPStatusCode: http.StatusBadRequest,
	},
}

// AssumedRoleUser - The identifiers for the temporary security credentials that
// the operation returns. Please also see https://docs.aws.amazon.com/goto/WebAPI/sts-2011-06-15/AssumedRoleUser
type AssumedRoleUser struct {
	Arn           string
	AssumedRoleID string `xml:"AssumeRoleId"`
}

// WebIdentityResult - Contains the response to a successful AssumeRoleWithWebIdentity
// request, including temporary credentials that can be used to make MinIO API requests.
type WebIdentityResult struct {
	// The identifiers for the temporary security credentials that the operation
	// returns.
	AssumedRoleUser AssumedRoleUser `xml:",omitempty"`

	// The intended audience (also known as client ID) of the web identity token.
	// This is traditionally the client identifier issued to the application that
	// requested the client grants.
	Audience string `xml:",omitempty"`

	// The temporary security credentials, which include an access key ID, a secret
	// access key, and a security (or session) token.
	//
	// Note: The size of the security token that STS APIs return is not fixed. We
	// strongly recommend that you make no assumptions about the maximum size. As
	// of this writing, the typical size is less than 4096 bytes, but that can vary.
	// Also, future updates to AWS might require larger sizes.
	Credentials Credentials `xml:",omitempty"`

	// A percentage value that indicates the size of the policy in packed form.
	// The service rejects any policy with a packed size greater than 100 percent,
	// which means the policy exceeded the allowed space.
	PackedPolicySize int `xml:",omitempty"`

	// The issuing authority of the web identity token presented. For OpenID Connect
	// ID tokens, this contains the value of the iss field. For OAuth 2.0 id_tokens,
	// this contains the value of the ProviderId parameter that was passed in the
	// AssumeRoleWithWebIdentity request.
	Provider string `xml:",omitempty"`

	// The unique user identifier that is returned by the identity provider.
	// This identifier is associated with the Token that was submitted
	// with the AssumeRoleWithWebIdentity call. The identifier is typically unique to
	// the user and the application that acquired the WebIdentityToken (pairwise identifier).
	// For OpenID Connect ID tokens, this field contains the value returned by the identity
	// provider as the token's sub (Subject) claim.
	SubjectFromWebIdentityToken string `xml:",omitempty"`
}

// AssumeRoleWithWebIdentityResponse contains the result of successful AssumeRoleWithWebIdentity request.
type AssumeRoleWithWebIdentityResponse struct {
	XMLName          xml.Name          `xml:"https://sts.amazonaws.com/doc/2011-06-15/ AssumeRoleWithWebIdentityResponse" json:"-"`
	Result           WebIdentityResult `xml:"AssumeRoleWithWebIdentityResult"`
	ResponseMetadata struct {
		RequestID string `xml:"RequestId,omitempty"`
	} `xml:"ResponseMetadata,omitempty"`
}

func configureSTSServer(c *Controller) *http.Server {
	router := mux.NewRouter().SkipClean(true).UseEncodedPath()

	router.Methods(http.MethodPost).
		Path(STSEndpoint + "/{tenantNamespace}").
		HandlerFunc(c.AssumeRoleWithWebIdentityHandler)

	router.NotFoundHandler = http.NotFoundHandler()

	s := &http.Server{
		Addr:           ":" + STSDefaultPort,
		Handler:        router,
		ReadTimeout:    time.Minute,
		WriteTimeout:   time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	return s
}

// writeSTSErrorRespone writes error headers
func writeSTSErrorResponse(w http.ResponseWriter, isErrCodeSTS bool, errCode STSErrorCode, errCtxt error) {
	var err APIError
	if isErrCodeSTS {
		err = stsErrCodes.ToSTSErr(errCode)
	}

	stsErrorResponse := STSErrorResponse{}
	stsErrorResponse.Error.Code = err.Code
	stsErrorResponse.RequestID = w.Header().Get(AmzRequestID)
	stsErrorResponse.Error.Message = err.Description
	if errCtxt != nil {
		stsErrorResponse.Error.Message = errCtxt.Error()
	}
	switch errCode {
	case ErrSTSInternalError, ErrSTSNotInitialized, ErrSTSUpstreamError:
		klog.Errorf("Error:%s/%s, err:%s", err.Code, stsErrorResponse.RequestID, errCtxt)
	}
	encodedErrorResponse := xhttp.EncodeResponse(stsErrorResponse)
	xhttp.WriteResponse(w, err.HTTPStatusCode, encodedErrorResponse, xhttp.MimeXML)
}

func writeSuccessResponseXML(w http.ResponseWriter, response []byte) {
	xhttp.WriteResponse(w, http.StatusOK, response, xhttp.MimeXML)
}

func (e stsErrorCodeMap) ToSTSErr(errCode STSErrorCode) APIError {
	apiErr, ok := e[errCode]
	if !ok {
		return e[ErrSTSInternalError]
	}
	return apiErr
}

// GetPolicy returns a tenant Policy by Name
func GetPolicy(ctx context.Context, adminClient *madmin.AdminClient, policyName string) (*madmin.PolicyInfo, error) {
	policy, err := adminClient.InfoCannedPolicyV2(ctx, policyName)
	return policy, err
}

// AssumeRole invokes the AssumeRole method in the Minio Tenant
func AssumeRole(ctx context.Context, c *Controller, tenant *miniov2.Tenant, sessionPolicy string, duration int) (*credentials.Value, error) {
	client, accessKey, secretKey, err := getTenantClient(ctx, c, tenant)
	if err != nil {
		return nil, err
	}

	host := tenant.MinIOServerEndpoint()
	if host == "" {
		return nil, errors.New("MinIO server host is empty")
	}

	stsOptions := credentials.STSAssumeRoleOptions{
		AccessKey:       accessKey,
		SecretKey:       secretKey,
		Policy:          sessionPolicy,
		DurationSeconds: duration,
	}

	stsAssumeRole := &credentials.STSAssumeRole{
		Client:      client,
		STSEndpoint: host,
		Options:     stsOptions,
	}

	stsCredentialsResponse, err := stsAssumeRole.Retrieve()
	if err != nil {
		return nil, err
	}
	return &stsCredentialsResponse, nil
}

// getTenantClient returns an http client that can be used to connect with the tenant
func getTenantClient(ctx context.Context, c *Controller, tenant *miniov2.Tenant) (*http.Client, string, string, error) {
	tenantConfiguration, err := c.getTenantCredentials(ctx, tenant)
	transport := c.getTransport()
	if err != nil {
		return nil, "", "", err
	}

	accessKey, ok := tenantConfiguration["accesskey"]
	if !ok {
		return nil, "", "", errors.New("MinIO server accesskey not set")
	}

	secretKey, ok := tenantConfiguration["secretkey"]
	if !ok {
		return nil, "", "", errors.New("MinIO server secretkey not set")
	}

	client := &http.Client{
		Transport: transport,
	}
	return client, string(accessKey), string(secretKey), nil
}

// ValidateServiceAccountJWT Executes a call to TokenReview  API to verify if the JWT Token received from the client
// is a valid Service Account JWT Token
func (c *Controller) ValidateServiceAccountJWT(ctx *context.Context, token string) (*authv1.TokenReview, error) {
	tr := authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token: token,
		},
	}

	tokenReviewResult, err := c.kubeClientSet.AuthenticationV1().TokenReviews().Create(*ctx, &tr, metav1.CreateOptions{})
	if err != nil {
		klog.Fatalf("Error building Kubernetes clientset: %s", err.Error())
		return nil, err
	}

	return tokenReviewResult, nil
}

// IsSTSEnabled Validates if the STS API is turned on, STS is disabled by default
// **WARNING** This will change and will be default to "on" in operator v5
func IsSTSEnabled() bool {
	value, set := os.LookupEnv(STSEnabled)
	return (set && value == "on")
}

// generateConsoleTLSCert Issues the Operator Console TLS Certificate
func (c *Controller) generateSTSTLSCert() (*string, *string) {
	return c.generateTLSCert("sts", STSTLSSecretName, getOperatorDeploymentName())
}

// waitSTSTLSCert Waits for the Operator leader to issue the TLS Certificate for STS
func (c *Controller) waitSTSTLSCert() (*string, *string) {
	return c.waitForCertSecretReady("sts", STSTLSSecretName)
}
