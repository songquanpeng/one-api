package sigv4

// ignoredHeaders is a list of headers that are always ignored during signing.
var ignoreHeaders = map[string]struct{}{
	"Authorization":   {},
	"User-Agent":      {},
	"X-Amzn-Trace-Id": {},
	// also include lower case canonical versions
	"authorization":   {},
	"user-agent":      {},
	"x-amzn-trace-id": {},
}

// requiredHeaderPrefix are header name prefixes that are mandatory for signing.
// If a header has one of these prefixes, it is a mandatory header.
var requiredHeaderPrefix = []string{"X-Amz-Object-Lock-", "X-Amz-Meta-"}

// requiredHeaders is a list of headers that are mandatory for signing.
var requiredHeaders = map[string]struct{}{
	"Cache-Control":                         {},
	"Content-Disposition":                   {},
	"Content-Encoding":                      {},
	"Content-Language":                      {},
	"Content-Md5":                           {},
	"Content-Type":                          {},
	"Expires":                               {},
	"If-Match":                              {},
	"If-Modified-Since":                     {},
	"If-None-Match":                         {},
	"If-Unmodified-Since":                   {},
	"Range":                                 {},
	"X-Amz-Acl":                             {},
	"X-Amz-Copy-Source":                     {},
	"X-Amz-Copy-Source-If-Match":            {},
	"X-Amz-Copy-Source-If-Modified-Since":   {},
	"X-Amz-Copy-Source-If-None-Match":       {},
	"X-Amz-Copy-Source-If-Unmodified-Since": {},
	"X-Amz-Copy-Source-Range":               {},
	"X-Amz-Copy-Source-Server-Side-Encryption-Customer-Algorithm": {},
	"X-Amz-Copy-Source-Server-Side-Encryption-Customer-Key":       {},
	"X-Amz-Copy-Source-Server-Side-Encryption-Customer-Key-Md5":   {},
	"X-Amz-Grant-Full-control":                                    {},
	"X-Amz-Grant-Read":                                            {},
	"X-Amz-Grant-Read-Acp":                                        {},
	"X-Amz-Grant-Write":                                           {},
	"X-Amz-Grant-Write-Acp":                                       {},
	"X-Amz-Metadata-Directive":                                    {},
	"X-Amz-Mfa":                                                   {},
	"X-Amz-Request-Payer":                                         {},
	"X-Amz-Server-Side-Encryption":                                {},
	"X-Amz-Server-Side-Encryption-Aws-Kms-Key-Id":                 {},
	"X-Amz-Server-Side-Encryption-Customer-Algorithm":             {},
	"X-Amz-Server-Side-Encryption-Customer-Key":                   {},
	"X-Amz-Server-Side-Encryption-Customer-Key-Md5":               {},
	"X-Amz-Storage-Class":                                         {},
	"X-Amz-Website-Redirect-Location":                             {},
	"X-Amz-Content-Sha256":                                        {},
	"X-Amz-Tagging":                                               {},
}

// isIgnoredHeader returns true if header must be ignored during signing.
func isIgnoredHeader(header string) bool {
	_, ok := ignoreHeaders[header]
	return ok
}

// isRequiredHeader returns true if header is mandatory for signing.
func isRequiredHeader(header string) bool {
	_, ok := requiredHeaders[header]
	if ok {
		return true
	}
	for _, v := range requiredHeaderPrefix {
		if hasPrefixFold(header, v) {
			return true
		}
	}
	return false
}

// isAllowQueryHoisting is a allowed list for Build query headers.
func isAllowQueryHoisting(header string) bool {
	if isRequiredHeader(header) {
		return false
	}
	return hasPrefixFold(header, "X-Amz-")
}
