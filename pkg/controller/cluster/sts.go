package cluster

import (
	"context"
	"encoding/xml"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
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
)

type contextKeyType string

//go:generate stringer -type=STSErrorCode -trimprefix=Err $GOFILE

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
	RemoteHost string // Client Host/IP
	Host       string // Node Host/IP
	UserAgent  string // User Agent
	RequestID  string // x-amz-request-id
	API        string // API name - GetObject PutObject NewMultipartUpload etc.
	AccessKey  string // Access Key
	ObjectName string // Object name
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

// AssumeRoleResult - Contains the response to a successful AssumeRole
// request, including temporary credentials that can be used to make
// MinIO API requests.
type AssumeRoleResult struct {
	AssumedRoleUser  AssumedRoleUser `xml:",omitempty"`
	Credentials      Credentials     `xml:",omitempty"`
	PackedPolicySize int             `xml:",omitempty"`
}

// AssumeRoleResponse contains the result of successful AssumeRole request.
type AssumeRoleResponse struct {
	XMLName xml.Name `xml:"https://sts.amazonaws.com/doc/2011-06-15/ AssumeRoleResponse" json:"-"`

	Result           AssumeRoleResult `xml:"AssumeRoleResult"`
	ResponseMetadata struct {
		RequestID string `xml:"RequestId,omitempty"`
	} `xml:"ResponseMetadata,omitempty"`
}

func configureSTSServer(c *Controller) *http.Server {
	router := mux.NewRouter().SkipClean(true).UseEncodedPath()

	router.Methods(http.MethodPost).
		Path(miniov2.STSEndpoint).
		Queries(stsAction, webIdentity).
		Queries(stsVersion, stsAPIVersion).
		HandlerFunc(c.AssumeRoleWithWebIdentityHandler)

	router.NotFoundHandler = http.NotFoundHandler()

	s := &http.Server{
		Addr:           ":" + miniov2.STSDefaultPort,
		Handler:        router,
		ReadTimeout:    time.Minute,
		WriteTimeout:   time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	return s
}

// writeSTSErrorRespone writes error headers
func writeSTSErrorResponse(ctx context.Context, w http.ResponseWriter, isErrCodeSTS bool, errCode STSErrorCode, errCtxt error) {
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

// AssumeRole invokes the AssumeRole method in the Minio Tenant
func AssumeRole(ctx context.Context, c *Controller, tenant *miniov2.Tenant, sessionPolicy string, duration int) (*credentials.Value, error) {
	tenantConfiguration, err := c.getTenantCredentials(ctx, tenant)
	transport := c.getTransport()
	if err != nil {
		return nil, err
	}

	host := tenant.MinIOServerHostAddress()
	if host == "" {
		return nil, errors.New("MinIO server host is empty")
	}

	accessKey, ok := tenantConfiguration["accesskey"]
	if !ok {
		return nil, errors.New("MinIO server accesskey not set")
	}

	secretKey, ok := tenantConfiguration["secretkey"]
	if !ok {
		return nil, errors.New("MinIO server secretkey not set")
	}

	stsOptions := credentials.STSAssumeRoleOptions{
		AccessKey:       string(accessKey),
		SecretKey:       string(secretKey),
		Policy:          sessionPolicy,
		DurationSeconds: duration,
	}

	client := &http.Client{
		Transport: transport,
	}

	stsAssumeRole := &credentials.STSAssumeRole{
		Client:      client,
		STSEndpoint: host, // TODO: Set the protocol right before the host var
		Options:     stsOptions,
	}

	stsCredentialsResponse, err := stsAssumeRole.Retrieve()
	if err != nil {
		return nil, err
	}
	return &stsCredentialsResponse, nil
}

// ValidateServiceAccountJWT Executes a call to TokenReview  API to verify if the JWT Token received from the client
// is a valid Service Account JWT Token
func (c *Controller) ValidateServiceAccountJWT(ctx *context.Context, token string) (*authv1.TokenReview, error) {
	tr := authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token:     token,
			Audiences: []string{"server"},
		},
	}

	tokenReviewResult, err := c.kubeClientSet.AuthenticationV1().TokenReviews().Create(*ctx, &tr, metav1.CreateOptions{})
	if err != nil {
		klog.Fatalf("Error building Kubernetes clientset: %s", err.Error())
		return nil, err
	}

	return tokenReviewResult, nil
}
