package http

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

var (
	// De-facto standard header keys.
	xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
	xRealIP       = http.CanonicalHeaderKey("X-Real-IP")
	// RFC7239 defines a new "Forwarded: " header designed to replace the
	// existing use of X-Forwarded-* headers.
	// e.g. Forwarded: for=192.0.2.60;proto=https;by=203.0.113.43
	forwarded = http.CanonicalHeaderKey("Forwarded")
	// Allows for a sub-match of the first value after 'for=' to the next
	// comma, semi-colon or space. The match is case-insensitive.
	forRegex = regexp.MustCompile(`(?i)(?:for=)([^(;|,| )]+)(.*)`)
)

// Standard S3 HTTP response constants
const (
	ContentType   = "Content-Type"
	ContentLength = "Content-Length"
	AcceptRanges  = "Accept-Ranges"
	ServerInfo    = "Server"
)

// mimeType represents various MIME type used API responses.
type mimeType string

const (
	// MimeNone Means no response type.
	MimeNone mimeType = ""
	// MimeXML Means response type is XML.
	MimeXML mimeType = "application/xml"
)

// UnescapeQueryPath URL unencode the path
func UnescapeQueryPath(ep string) (string, error) {
	ep, err := url.QueryUnescape(ep)
	if err != nil {
		return "", err
	}
	return TrimLeadingSlash(ep), nil
}

// TrimLeadingSlash Cleans and ensure there is a leading slash path in the URL
func TrimLeadingSlash(ep string) string {
	if len(ep) > 0 && ep[0] == '/' {
		// Path ends with '/' preserve it
		if ep[len(ep)-1] == '/' && len(ep) > 1 {
			ep = path.Clean(ep)
			ep += "/"
		} else {
			ep = path.Clean(ep)
		}
		ep = ep[1:]
	}
	return ep
}

// GetSourceIPFromHeaders retrieves the IP from the X-Forwarded-For, X-Real-IP
// and RFC7239 Forwarded headers (in that order)
func GetSourceIPFromHeaders(r *http.Request) string {
	var addr string

	if fwd := r.Header.Get(xForwardedFor); fwd != "" {
		// Only grab the first (client) address. Note that '192.168.0.1,
		// 10.1.1.1' is a valid key for X-Forwarded-For where addresses after
		// the first may represent forwarding proxies earlier in the chain.
		s := strings.Index(fwd, ", ")
		if s == -1 {
			s = len(fwd)
		}
		addr = fwd[:s]
	} else if fwd := r.Header.Get(xRealIP); fwd != "" {
		// X-Real-IP should only contain one IP address (the client making the
		// request).
		addr = fwd
	} else if fwd := r.Header.Get(forwarded); fwd != "" {
		// match should contain at least two elements if the protocol was
		// specified in the Forwarded header. The first element will always be
		// the 'for=' capture, which we ignore. In the case of multiple IP
		// addresses (for=8.8.8.8, 8.8.4.4, 172.16.1.20 is valid) we only
		// extract the first, which should be the client IP.
		if match := forRegex.FindStringSubmatch(fwd); len(match) > 1 {
			// IPv6 addresses in Forwarded headers are quoted-strings. We strip
			// these quotes.
			addr = strings.Trim(match[1], `"`)
		}
	}

	if addr != "" {
		return addr
	}
	// Default to remote address if headers not set.
	addr, _, _ = net.SplitHostPort(r.RemoteAddr)
	if strings.ContainsRune(addr, ':') {
		return "[" + addr + "]"
	}
	return addr
}

// EncodeResponse Encodes the response headers into XML format.
func EncodeResponse(response interface{}) []byte {
	var bytesBuffer bytes.Buffer
	bytesBuffer.WriteString(xml.Header)
	e := xml.NewEncoder(&bytesBuffer)
	e.Encode(response)
	return bytesBuffer.Bytes()
}

// ParseForm Parses form fields
func ParseForm(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	for k, v := range r.PostForm {
		if _, ok := r.Form[k]; !ok {
			r.Form[k] = v
		}
	}
	return nil
}

// WriteResponse writes ressponse to http.ResponseWriter
func WriteResponse(w http.ResponseWriter, statusCode int, response []byte, mType mimeType) {
	if statusCode == 0 {
		statusCode = 200
	}
	// Similar check to http.checkWriteHeaderCode
	if statusCode < 100 || statusCode > 999 {
		klog.Errorf(fmt.Sprintf("invalid WriteHeader code %v", statusCode))
		statusCode = http.StatusInternalServerError
	}
	SetCommonHeaders(w)
	if mType != MimeNone {
		w.Header().Set(ContentType, string(mType))
	}
	w.Header().Set(ContentLength, strconv.Itoa(len(response)))
	w.WriteHeader(statusCode)
	if response != nil {
		w.Write(response)
	}
}

// SetCommonHeaders writes http common headers
func SetCommonHeaders(w http.ResponseWriter) {
	// Set the "Server" http header.
	w.Header().Set(ServerInfo, "MinIO")
	w.Header().Set(AcceptRanges, "bytes")
}
