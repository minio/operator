package cluster

import (
	"context"
	"encoding/xml"
	"errors"
	"net/http"
	"sync"

	"github.com/minio/minio-go/v7/pkg/credentials"
	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	xhttp "github.com/minio/operator/pkg/internal"
	"k8s.io/klog/v2"
)

// STS Handler constants
const (
	assumeRole          = "AssumeRole"
	webIdentity         = "AssumeRoleWithWebIdentity"
	stsAPIVersion       = "2011-06-15"
	stsVersion          = "Version"
	stsAction           = "Action"
	stsPolicy           = "Policy"
	stsToken            = "Token"
	stsRoleArn          = "RoleArn"
	stsWebIdentityToken = "WebIdentityToken"
	stsDurationSeconds  = "DurationSeconds"
	AmzRequestID        = "x-amz-request-id"
	stsRequestBodyLimit = 10 * (1 << 20) // 10 MiB

)

type contextKeyType string

const ContextLogKey = contextKeyType("operatorlog")

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

type stsErrorCodeMap map[STSErrorCode]STSError

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

// STSErrorCode type of error status.
type STSErrorCode int

// STSError structure
type STSError struct {
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

// writeSTSErrorRespone writes error headers
func writeSTSErrorResponse(ctx context.Context, w http.ResponseWriter, isErrCodeSTS bool, errCode STSErrorCode, errCtxt error) {
	var err STSError
	if isErrCodeSTS {
		err = stsErrCodes.ToSTSErr(errCode)
	}
	//TODO Parse error from Minio Instance
	// if err.Code == "InternalError" || !isErrCodeSTS {
	// 	aerr := getAPIError(APIErrorCode(errCode))
	// 	if aerr.Code != "InternalError" {
	// 		err.Code = aerr.Code
	// 		err.Description = aerr.Description
	// 		err.HTTPStatusCode = aerr.HTTPStatusCode
	// 	}
	// }
	// Generate error response.
	stsErrorResponse := STSErrorResponse{}
	stsErrorResponse.Error.Code = err.Code
	stsErrorResponse.RequestID = w.Header().Get(AmzRequestID)
	stsErrorResponse.Error.Message = err.Descripton
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

func (e stsErrorCodeMap) ToSTSErr(errCode STSErrorCode) STSError {
	apiErr, ok := e[errCode]
	if !ok {
		return e[ErrSTSInternalError]
	}
	return apiErr
}

func AssumeRole(ctx context.Context, c *Controller, tenant *miniov2.Tenant, sessionPolicy string) (bool, error) {

	tenantConfiguration, err := c.getTenantCredentials(ctx, tenant)
	transport := c.getTransport()
	if err != nil {
		return false, err
	}

	host := tenant.MinIOServerHostAddress()
	if host == "" {
		return false, errors.New("MinIO server host is empty")
	}

	accessKey, ok := tenantConfiguration["accesskey"]
	if !ok {
		return false, errors.New("MinIO server accesskey not set")
	}

	secretKey, ok := tenantConfiguration["secretkey"]
	if !ok {
		return false, errors.New("MinIO server secretkey not set")
	}

	stsOptions := credentials.STSAssumeRoleOptions{
		AccessKey:       string(accessKey),
		SecretKey:       string(secretKey),
		Policy:          sessionPolicy,
		DurationSeconds: 600, //TODO: get duration and calculate maxium
	}

	// creds, err := credentials.NewSTSAssumeRole(host, stsOptions)
	// if err != nil {
	// 	return false, err
	// }

	// opts := &minio.Options{
	// 	Transport: transport,
	// 	Secure:    tenant.TLS(),
	// 	Creds:     creds,
	// }

	// minioClient, err := minio.New(host, opts)
	// if err != nil {
	// 	return false, err
	// }

	// minioClient.ListBuckets(ctx)
	client := &http.Client{
		Transport: transport,
	}

	stsAssumeRole := &credentials.STSAssumeRole{
		Client:      client,
		STSEndpoint: host,
		Options:     stsOptions,
	}
	p := &credentials.STSWebIdentity{
		Client:      http.DefaultClient,
		STSEndpoint: host,
	}

	credentials, err := p.Retrieve()

	if err != nil {
		return false, err
	}

	//minioClient, err := tenant.NewMinIOUser(tenantConfiguration, c.getTransport())

	adminClient, err := tenant.NewMinIOAdmin(tenantConfiguration, c.getTransport())
	if err != nil {
		klog.Errorf("Error instantiating madmin: %v", err)
		return false, err
	}
	//adminClient.AddUser(ctx, ac, sk)
	//iampolicy.SessionPolicyName
	adminClient, err := tenant.NewMinIOAdmin(tenantConfiguration, c.getTransport())
	if err != nil {
		klog.Errorf("Error instantiating madmin: %v", err)
		return
	}
}
