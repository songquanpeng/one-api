package sigv4

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

var (
	awsURLNoEscTable [256]bool
	awsURLEscTable   [256][2]byte
)

func init() {
	for i := 0; i < len(awsURLNoEscTable); i++ {
		// every char except these must be escaped
		awsURLNoEscTable[i] = (i >= 'A' && i <= 'Z') ||
			(i >= 'a' && i <= 'z') ||
			(i >= '0' && i <= '9') ||
			i == '-' ||
			i == '.' ||
			i == '_' ||
			i == '~'
		// %<hex><hex>
		encoded := fmt.Sprintf("%02X", i)
		awsURLEscTable[i] = [2]byte{encoded[0], encoded[1]}
	}
}

// hmacsha256 computes a HMAC-SHA256 of data given the provided key.
func hmacsha256(key, data, buf []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(buf)
}

// hasPrefixFold tests whether the string s begins with prefix, interpreted as
// UTF-8 strings, under Unicode case-folding.
func hasPrefixFold(s, prefix string) bool {
	return len(s) >= len(prefix) &&
		strings.EqualFold(s[0:len(prefix)], prefix)
}

// hostOrURLHost returns r.Host, or if empty, r.URL.Host.
func hostOrURLHost(r *http.Request) string {
	if r.Host != "" {
		return r.Host
	}
	return r.URL.Host
}

// parsePort returns the port part of u.Host, without the leading colon. Returns
// an empty string if u.Host  doesn't contain port.
//
// Adapted from the Go 1.8 standard library (net/url).
func parsePort(hostport string) string {
	colon := strings.IndexByte(hostport, ':')
	if colon == -1 || colon == len(hostport)-1 {
		return ""
	}

	// take care of ipv6 syntax: [a:b::]:<port>
	const ipv6Sep = "]:"
	if i := strings.Index(hostport, ipv6Sep); i != -1 {
		return hostport[i+len(ipv6Sep):]
	}
	if strings.Contains(hostport, "]") {
		return ""
	}

	return hostport[colon+1:]
}

// stripPort returns Hostname portion of u.Host, i.e. without any port number.
//
// If hostport is an IPv6 literal with a port number, returns the IPv6 literal
// without the square brackets. IPv6 literals may include a zone identifier.
//
// Adapted from the Go 1.8 standard library (net/url).
func stripPort(hostport string) string {
	colon := strings.IndexByte(hostport, ':')
	if colon == -1 {
		return hostport
	}
	// ipv6: remove the []
	if i := strings.IndexByte(hostport, ']'); i != -1 {
		return strings.TrimPrefix(hostport[:i], "[")
	}
	return hostport[:colon]
}

// isDefaultPort returns true if the specified URI is using the standard port
// (i.e. port 80 for HTTP URIs or 443 for HTTPS URIs).
func isDefaultPort(scheme, port string) bool {
	switch strings.ToLower(scheme) {
	case "http":
		return port == "80"
	case "https":
		return port == "443"
	default:
		return false
	}
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	u2 := new(url.URL)
	*u2 = *u
	if u.User != nil {
		u2.User = new(url.Userinfo)
		*u2.User = *u.User
	}
	return u2
}

// writeAWSURIPath writes the escaped URI component from the specified URL (using
// AWS canonical URI specification) into w. URI component is path without query
// string.
func writeAWSURIPath(w *bufio.Writer, u *url.URL, encodeSep bool, isEscaped bool) {
	const schemeSep, pathSep, queryStart = "//", "/", "?"

	var p string
	if u.Opaque == "" {
		p = u.EscapedPath()
	} else {
		opaque := u.Opaque
		// discard query string if any
		if i := strings.Index(opaque, queryStart); i != -1 {
			opaque = opaque[:i]
		}
		// if has scheme separator as prefix, discard it
		if strings.HasPrefix(opaque, schemeSep) {
			opaque = opaque[len(schemeSep):]
		}

		// everything after the first /, including the /
		if i := strings.Index(opaque, pathSep); i != -1 {
			p = opaque[i:]
		}
	}

	if p == "" {
		w.WriteByte('/')
		return
	}

	if isEscaped {
		w.WriteString(p)
		return
	}

	// Loop thru first like in https://cs.opensource.google/go/go/+/refs/tags/go1.20.2:/src/net/url/url.go.
	// It may add ~800ns but we save on memory alloc and catches cases where there
	// is no need to escape.
	plen := len(p)
	strlen := plen
	for i := 0; i < plen; i++ {
		c := p[i]
		if awsURLNoEscTable[c] || (c == '/' && !encodeSep) {
			continue
		}
		strlen += 2
	}

	// path already canonical, no need to escape
	if plen == strlen {
		w.WriteString(p)
		return
	}

	for i := 0; i < plen; i++ {
		c := p[i]
		if awsURLNoEscTable[c] || (c == '/' && !encodeSep) {
			w.WriteByte(c)
			continue
		}
		w.Write([]byte{'%', awsURLEscTable[c][0], awsURLEscTable[c][1]})
	}
}

// writeCanonicalQueryParams builds the canonical form of query and writes to w.
//
// Side effect: query values are sorted after this function returns.
func writeCanonicalQueryParams(w *bufio.Writer, query url.Values) {
	qlen := len(query)
	if qlen == 0 {
		return
	}

	keys := make([]string, 0, qlen)
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		keyEscaped := strings.Replace(url.QueryEscape(k), "+", "%20", -1)
		vs := query[k]

		if i != 0 {
			w.WriteByte('&')
		}

		if len(vs) == 0 {
			w.WriteString(keyEscaped)
			w.WriteByte('=')
			continue
		}

		sort.Strings(vs)
		for j, v := range vs {
			if j != 0 {
				w.WriteByte('&')
			}
			w.WriteString(keyEscaped)
			w.WriteByte('=')
			if v != "" {
				w.WriteString(strings.Replace(url.QueryEscape(v), "+", "%20", -1))
			}
		}
	}
}

// writeCanonicalString removes leading and trailing whitespaces (as defined by Unicode)
// in s, replaces consecutive spaces (' ') in s with a single space, and then
// write the result to w.
func writeCanonicalString(w *bufio.Writer, s string) {
	const dblSpace = "  "

	s = strings.TrimSpace(s)

	// bail if str doesn't contain "  "
	j := strings.Index(s, dblSpace)
	if j < 0 {
		w.WriteString(s)
		return
	}

	w.WriteString(s[:j])

	// replace all "  " with " " in a performant way
	var lastIsSpace bool
	for i, l := j, len(s); i < l; i++ {
		if s[i] == ' ' {
			if !lastIsSpace {
				w.WriteByte(' ')
				lastIsSpace = true
			}
			continue
		}
		lastIsSpace = false
		w.WriteByte(s[i])
	}
}
