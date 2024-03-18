package sigv4

import (
	"crypto/sha256"
	"strings"
	"sync"
	"time"
	gotime "time"
)

var credScopeSuffixBytes = []byte{'a', 'w', 's', '4', '_', 'r', 'e', 'q', 'u', 'e', 's', 't'}

// deriveKey calculates the signing key. See https://docs.aws.amazon.com/general/latest/gr/create-signed-request.html.
func deriveKey(secret, service, region string, t Time) []byte {
	// enc(
	//     enc(
	//         enc(
	//             enc(AWS4<secret>, <short_time>),
	//             <region>),
	//         <service>),
	//    "aws4_request")

	// https://en.wikipedia.org/wiki/HMAC
	// HMAC_SHA256 produces 32 bytes output

	f1 := len(secret) + 4
	f2 := f1 + len(t.ShortTimeFormat())
	f3 := f2 + len(region)
	f4 := f3 + len(service)

	qs := make([]byte, 0, f4)
	qs = append(qs, "AWS4"...)
	qs = append(qs, secret...)
	qs = append(qs, t.ShortTimeFormat()...)
	qs = append(qs, region...)
	qs = append(qs, service...)

	buf := make([]byte, 0, sha256.BlockSize)
	buf = hmacsha256(qs[:f1], qs[f1:f2], buf)
	buf = hmacsha256(buf, qs[f2:f3], buf[:0])
	buf = hmacsha256(buf, qs[f3:], buf[:0])
	return hmacsha256(buf, credScopeSuffixBytes, buf[:0])
}

// keyDeriver returns a signing key based on parameters such as credentials.
type keyDeriver interface {
	DeriveKey(accessKey, secret, service, region string, sigtime Time) []byte
}

// signingKeyDeriver is the default implementation of keyDerivator.
type signingKeyDeriver struct {
	cache derivedKeyCache
}

// newKeyDeriver creates a keyDeriver using the default implementation. The
// signing key is cached per region/service, and updated when accessKey changes
// or signingTime is not on the same day for that region/service.
func newKeyDeriver() keyDeriver {
	return &signingKeyDeriver{cache: newDerivedKeyCache()}
}

// DeriveKey returns a derived signing key from the given credentials to be used
// with SigV4 signing.
func (k *signingKeyDeriver) DeriveKey(accessKey, secret, service, region string, sigtime Time) []byte {
	return k.cache.Get(accessKey, secret, service, region, sigtime)
}

type derivedKeyCache struct {
	mutex   sync.RWMutex
	values  map[string]derivedKey
	nowFunc func() gotime.Time
}

type derivedKey struct {
	Date       gotime.Time
	Credential []byte
}

func newDerivedKeyCache() derivedKeyCache {
	return derivedKeyCache{
		values:  make(map[string]derivedKey),
		nowFunc: gotime.Now,
	}
}

// Get returns key from cache or creates a new one.
func (s *derivedKeyCache) Get(accessKey, secret, service, region string, sigtime Time) []byte {
	// <accessKey>/<YYYYMMDD>/<region>/<service>
	key := strings.Join([]string{accessKey, sigtime.ShortTimeFormat(), region, service}, "/")

	s.mutex.RLock()
	cred, status := s.getFromCache(key)
	s.mutex.RUnlock()
	if status == 0 {
		return cred
	}

	cred = deriveKey(secret, service, region, sigtime)

	s.mutex.Lock()
	if status == -1 {
		delete(s.values, key)
	}
	s.values[key] = derivedKey{
		Date:       sigtime.Time,
		Credential: cred,
	}
	s.mutex.Unlock()

	return cred
}

// getFromCache returns s.values[key]. Second result is 1 if key was not found,
// or -1 if the cached value has expired.
func (s *derivedKeyCache) getFromCache(key string) ([]byte, int) {
	v, ok := s.values[key]
	if !ok {
		return nil, 1
	}
	// evict from cache if item is a day older than system time
	if s.nowFunc().Sub(v.Date) > 24*time.Hour {
		return nil, -1
	}
	return v.Credential, 0
}
