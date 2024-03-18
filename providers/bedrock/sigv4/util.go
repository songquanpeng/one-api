package sigv4

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

// ContentSHA256Sum calculates the hex-encoded SHA256 checksum of r.Body. It returns
// EmptyStringSHA256 if r.Body is nil, r.Method is TRACE or r.ContentLength is zero.
// Returns non-nil error if r.Body cannot be read.
func ContentSHA256Sum(r *http.Request) (string, error) {
	// We need to check r.Body is non-nil, because io.Copy(dst, nil) panics.
	// This is not documented in https://pkg.go.dev/io#Copy.
	if r.Method == http.MethodTrace || r.ContentLength == 0 || r.Body == nil {
		return EmptyStringSHA256, nil
	}

	h := sha256.New()
	_, err := io.Copy(h, r.Body)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
