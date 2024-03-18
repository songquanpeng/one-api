package sigv4

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"net/http"
	"net/textproto"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// HTTPSigner is an AWS SigV4 signer that can sign HTTP requests.
type HTTPSigner interface {
	// Sign AWS v4 requests with the provided payload hash, service name, region
	// the request is made to, and time the request is signed at. Set sigtime
	// to the future to create a request that cannot be used until the future time.
	//
	// payloadHash is the hex encoded SHA-256 hash of the request payload, and must
	// not be empty, even if the request has no payload (aka body). If the request
	// has no payload, use the hex encoded SHA-256 of an empty string, or the constant
	// EmptyStringSHA256. You can use the utility function ContentSHA256Sum to
	// calculate the hash of a http.Request body.
	//
	// Some services such as Amazon S3 accept alternative values for the payload
	// hash, such as "UNSIGNED-PAYLOAD" for requests where the body will not be
	// protected by sigv4. See https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-header-based-auth.html.
	//
	// Sign differs from Presign in that it will sign the request using HTTP headers.
	// The passed in request r will be modified in place: modified fields include
	// r.Host and r.Header.
	Sign(r *http.Request, payloadHash string, sigtime Time) error
	// Presign is like Sign, but does not modify request r. It returns a copy of
	// r.URL with additional query parameters that contains signing information.
	// The URL can be used to recreate an authenticated request without specifying
	// headers. It also returns http.Header as a second result, which must be
	// included in the reconstructed request.
	//
	// Header hoisting: use WithHeaderHoisting option function to specify whether
	// headers in request r should be added as query parameters. Some headers cannot
	// be hoisted, and are returned as the second result.
	//
	// Presign will not set the expires time of the presigned request automatically.
	// To specify the expire duration for a request, add the "X-Amz-Expires" query
	// parameter on the request with the value as the duration in seconds the
	// presigned URL should be considered valid for. This parameter is not used
	// by all AWS services, and is most notable used by Amazon S3 APIs.
	//
	//     expires := 20*time.Minute
	//     query := req.URL.Query()
	//     query.Set("X-Amz-Expires", strconv.FormatInt(int64(expires/time.Second), 10)
	//     req.URL.RawQuery = query.Encode()
	Presign(r *http.Request, payloadHash string, sigtime Time) (*url.URL, http.Header, error)
}

// HTTPSignerOption is an option parameter for HTTPSigner constructor function.
type HTTPSignerOption func(HTTPSigner) error

// ErrInvalidOption means the option parameter is incompatible with the HTTPSigner.
var ErrInvalidOption = errors.New("cannot apply option to HTTPSigner")

// httpV4Signer is the default implementation of HTTPSigner.
type httpV4Signer struct {
	KeyDeriver     keyDeriver
	AccessKey      string
	Secret         string
	SessionToken   string
	Service        string
	Region         string
	HeaderHoisting bool
	EscapeURLPath  bool
}

// WithCredential sets HTTPSigner credential fields.
func WithCredential(accessKey, secret, sessionToken string) HTTPSignerOption {
	return func(signer HTTPSigner) error {
		if sigv4, ok := signer.(*httpV4Signer); ok {
			sigv4.AccessKey = accessKey
			sigv4.Secret = secret
			sigv4.SessionToken = sessionToken
			return nil
		}
		return ErrInvalidOption
	}
}

// WithHeaderHoisting specifies whether HTTPSigner automatically hoist headers.
// Default is enabled.
func WithHeaderHoisting(enable bool) HTTPSignerOption {
	return func(signer HTTPSigner) error {
		if sigv4, ok := signer.(*httpV4Signer); ok {
			sigv4.HeaderHoisting = enable
			return nil
		}
		return ErrInvalidOption
	}
}

// WithEscapeURLPath specifies whether HTTPSigner automatically escapes URL paths.
// Default is enabled.
func WithEscapeURLPath(enable bool) HTTPSignerOption {
	return func(signer HTTPSigner) error {
		if sigv4, ok := signer.(*httpV4Signer); ok {
			sigv4.EscapeURLPath = enable
			return nil
		}
		return ErrInvalidOption
	}
}

// WithRegionService sets HTTPSigner region and service fields.
func WithRegionService(region, service string) HTTPSignerOption {
	return func(signer HTTPSigner) error {
		if sigv4, ok := signer.(*httpV4Signer); ok {
			sigv4.Region = region
			sigv4.Service = service
			return nil
		}
		return ErrInvalidOption
	}
}

// New creates a HTTPSigner.
func New(opts ...HTTPSignerOption) (HTTPSigner, error) {
	sigv4 := &httpV4Signer{
		KeyDeriver:     newKeyDeriver(),
		EscapeURLPath:  true,
		HeaderHoisting: true,
	}
	for _, o := range opts {
		if o == nil {
			continue
		}
		if err := o(sigv4); err != nil {
			return nil, err
		}
	}
	return sigv4, nil
}

// Sign implements HTTPSigner.
func (s *httpV4Signer) Sign(r *http.Request, payloadHash string, sigtime Time) error {
	if payloadHash == "" {
		var err error
		payloadHash, err = ContentSHA256Sum(r)
		if err != nil {
			return err
		}
	}

	// add mandatory headers to r.Header
	setRequiredSigningHeaders(r.Header, sigtime, s.SessionToken)
	// remove port in r.Host if any
	r.Host = sanitizeHostForHeader(r)

	// parse URL query only once
	query := r.URL.Query()
	// sigBuf is used to act as a sha256 hash buffer
	sigBuf := make([]byte, 0, sha256.Size)

	//hasher := &debugHasher{}
	hasher := sha256.New()
	reqhash, signedHeaderStr := canonicalRequestHash(hasher, r, r.Header, query,
		r.Host, payloadHash, s.EscapeURLPath, false, sigBuf)

	credentialScope := strings.Join([]string{
		sigtime.ShortTimeFormat(),
		s.Region,
		s.Service,
		"aws4_request",
	}, "/")

	keyBytes := s.KeyDeriver.DeriveKey(s.AccessKey, s.Secret, s.Service,
		s.Region, sigtime)
	sigHasher := hmac.New(sha256.New, keyBytes)
	signature := authorizationSignature(sigHasher, sigtime, credentialScope, reqhash, sigBuf)

	writeAuthorizationHeader(r.Header, s.AccessKey+"/"+credentialScope,
		signedHeaderStr, signature)

	// done
	return nil
}

// Presign implements HTTPSigner.
func (s *httpV4Signer) Presign(r *http.Request, payloadHash string, sigtime Time) (*url.URL, http.Header, error) {
	if payloadHash == "" {
		var err error
		payloadHash, err = ContentSHA256Sum(r)
		if err != nil {
			return nil, nil, err
		}
	}

	query := r.URL.Query()
	setRequiredSigningQuery(query, sigtime, s.SessionToken)
	// sort each query key's values
	for key := range query {
		sort.Strings(query[key])
	}

	credentialScope := strings.Join([]string{
		sigtime.ShortTimeFormat(),
		s.Region,
		s.Service,
		"aws4_request",
	}, "/")
	credentialStr := s.AccessKey + "/" + credentialScope
	query.Set(AmzCredentialKey, credentialStr)

	var headersLeft http.Header
	if s.HeaderHoisting {
		headersLeft = make(http.Header, len(r.Header))
		for k, v := range r.Header {
			if isAllowQueryHoisting(k) {
				query[k] = v
			} else {
				headersLeft[k] = v
			}
		}
	}

	// sigBuf is used to act as a sha256 hash buffer
	sigBuf := make([]byte, 0, sha256.Size)

	hasher := sha256.New()
	reqhash, signedHeaderStr := canonicalRequestHash(hasher, r, headersLeft,
		query, sanitizeHostForHeader(r), payloadHash, s.EscapeURLPath, true, sigBuf)

	keyBytes := s.KeyDeriver.DeriveKey(s.AccessKey, s.Secret, s.Service,
		s.Region, sigtime)
	sigHasher := hmac.New(sha256.New, keyBytes)
	signature := authorizationSignature(sigHasher, sigtime, credentialScope, reqhash, sigBuf)
	query.Set(AmzSignatureKey, signature)

	u := cloneURL(r.URL)
	u.RawQuery = strings.Replace(query.Encode(), "+", "%20", -1)

	// For the signed headers we canonicalize the header keys in the returned map.
	// This avoids situations where standard library can sometimes add double
	// headers. For example, the standard library will set the Host header,
	// even if it is present in lower-case form.
	signedHeader := strings.Split(signedHeaderStr, ";")
	canonHeader := make(http.Header, len(signedHeader))
	for _, k := range signedHeader {
		canonKey := textproto.CanonicalMIMEHeaderKey(k)
		switch k {
		case "host":
			canonHeader[canonKey] = []string{sanitizeHostForHeader(r)}
		case "content-length":
			canonHeader[canonKey] = []string{strconv.FormatInt(r.ContentLength, 10)}
		default:
			canonHeader[canonKey] = append(canonHeader[canonKey], headersLeft[http.CanonicalHeaderKey(k)]...)
		}
	}
	return u, canonHeader, nil
}

// authorizationSignature returns `sig` as documented in step 4 of algorithm
// documentation. key is hSig in step 4. It calculates the result of step 3
// internally.
func authorizationSignature(hasher hash.Hash, sigtime Time, credScope, requestHash string, buf []byte) string {
	w := bufio.NewWriterSize(hasher, sha256.BlockSize)

	w.WriteString(SigningAlgorithm)
	w.WriteByte('\n')
	w.WriteString(sigtime.TimeFormat())
	w.WriteByte('\n')
	w.WriteString(credScope)
	w.WriteByte('\n')
	w.WriteString(requestHash)

	w.Flush() // VERY IMPORTANT! Don't forget to flush remaining buffer
	//hasher.Println()
	return hex.EncodeToString(hasher.Sum(buf[:0]))
}

// canonicalRequestHash returns the hex-encoded sha256 sum of the canonical
// request string. Refer to step 2 of algorithm documentation. Expect hasher to
// be sha256.New.
func canonicalRequestHash(
	hasher hash.Hash, r *http.Request, headers http.Header, query url.Values,
	hostname, hashcode string, escapeURL, isPresign bool, buf []byte,
) (string, string) {
	w := bufio.NewWriterSize(hasher, sha256.BlockSize)

	signedHeaders := make([]string, 0, len(headers)+2)
	signedHeaders = append(signedHeaders, "host")
	if r.ContentLength > 0 {
		signedHeaders = append(signedHeaders, "content-length")
	}
	for k := range headers {
		if strings.EqualFold(k, "content-length") || strings.EqualFold(k, "host") || isIgnoredHeader(k) {
			continue
		}
		signedHeaders = append(signedHeaders, strings.ToLower(k))
	}
	sort.Strings(signedHeaders)
	signedHeaderStr := strings.Join(signedHeaders, ";")

	// for presigned requests, we need to add X-Amz-SignedHeaders to calculate the
	// correct hash
	if isPresign {
		query.Set(AmzSignedHeadersKey, signedHeaderStr)
	}

	// <METHOD>\n<URI>\n<QUERY>\n<HEADERS>\n<SIGNED_HEADERS>\n<PAYLOAD_HASH>

	// HTTP_METHOD
	w.WriteString(r.Method)
	w.WriteByte('\n')
	// CANONICAL_URI
	writeAWSURIPath(w, r.URL, false, !escapeURL)
	w.WriteByte('\n')
	// CANONICAL_QUERY_PARAMS
	writeCanonicalQueryParams(w, query)
	w.WriteByte('\n')
	// CANONICAL_HEADERS
	for _, head := range signedHeaders {
		switch head {
		case "host":
			w.WriteString(head)
			w.WriteByte(':')
			writeCanonicalString(w, hostname)
			w.WriteByte('\n')
		case "content-length":
			w.WriteString(head)
			w.WriteByte(':')
			w.WriteString(strconv.FormatInt(r.ContentLength, 10))
			w.WriteByte('\n')
		default:
			w.WriteString(head)
			w.WriteByte(':')
			values := headers[http.CanonicalHeaderKey(head)]
			for i, v := range values {
				if i != 0 {
					w.WriteByte(',')
				}
				writeCanonicalString(w, v)
			}
			w.WriteByte('\n')
		}
	}
	w.WriteByte('\n')
	// SIGNED_HEADERS
	w.WriteString(signedHeaderStr)
	w.WriteByte('\n')
	// PAYLOAD_HASH
	w.WriteString(hashcode)

	w.Flush() // VERY IMPORTANT! Don't forget to flush remaining buffer
	//hasher.Println()
	return hex.EncodeToString(hasher.Sum(buf[:0])), signedHeaderStr
}

// writeAuthorizationHeader writes the Authorization header into header:
//
//	AWS4-HMAC-SHA256 Credential=<cred>, SignedHeaders=<signed_headers>, Signature=<sig>
func writeAuthorizationHeader(headers http.Header, credentialStr, signedHeaders, signature string) {
	const credentialPrefix = "Credential="
	const signedHeadersPrefix = "SignedHeaders="
	const signaturePrefix = "Signature="
	const commaSpace = ", "

	var parts strings.Builder
	parts.Grow(len(SigningAlgorithm) + 1 +
		len(credentialPrefix) + len(credentialStr) + 2 +
		len(signedHeadersPrefix) + len(signedHeaders) + 2 +
		len(signaturePrefix) + len(signature))

	parts.WriteString(SigningAlgorithm)
	parts.WriteRune(' ')
	parts.WriteString(credentialPrefix)
	parts.WriteString(credentialStr)
	parts.WriteString(commaSpace)
	parts.WriteString(signedHeadersPrefix)
	parts.WriteString(signedHeaders)
	parts.WriteString(commaSpace)
	parts.WriteString(signaturePrefix)
	parts.WriteString(signature)

	headers[authorizationHeader] = append(headers[authorizationHeader][:0],
		parts.String())
}

// helpers

// sanitizeHostForHeader is like hostOrURLHost, but without port if port is the
// default port for the scheme. For example, it removes ":80" suffix if the scheme
// is "http".
func sanitizeHostForHeader(r *http.Request) string {
	host := hostOrURLHost(r)
	port := parsePort(host)
	if port != "" && isDefaultPort(r.URL.Scheme, port) {
		return stripPort(host)
	}
	return host
}

// setRequiredSigningHeaders modifies headers: sets X-Amz-Date to sigtime, and
// if credToken is non-empty, set X-Amz-Security-Token to credToken. This function
// overwrites existing headers values with the same key.
func setRequiredSigningHeaders(headers http.Header, sigtime Time, sessionToken string) {
	amzDate := sigtime.TimeFormat()
	headers[AmzDateKey] = append(headers[AmzDateKey][:0], amzDate)
	if sessionToken != "" {
		headers[AmzSecurityTokenKey] = append(headers[AmzSecurityTokenKey][:0],
			sessionToken)
	}
}

// setRequiredSigningQuery is like setRequiredSigningHeaders, but modifies
// query values. This is used for presign requests.
func setRequiredSigningQuery(query url.Values, sigtime Time, sessionToken string) {
	query.Set(AmzAlgorithmKey, SigningAlgorithm)

	amzDate := sigtime.TimeFormat()
	query.Set(AmzDateKey, amzDate)

	if sessionToken != "" {
		query.Set(AmzSecurityTokenKey, sessionToken)
	}
}
