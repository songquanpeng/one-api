package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/ssrf"
)

var HTTPClient *http.Client
var ImpatientHTTPClient *http.Client

// UserContentRequestHTTPClient is the HTTP client used to fetch user content (images, files, etc.)
// SSRF protection is applied to this client.
var UserContentRequestHTTPClient *http.Client

var userContentTransport *http.Transport
var userContentTransportMu sync.RWMutex

func Init() {
	ctx := context.Background()
	initUserContentRequestClient(ctx)
	initRelayClient(ctx)
}

func initUserContentRequestClient(ctx context.Context) {
	var transport *http.Transport

	if config.UserContentRequestProxy != "" {
		logger.SysLog(fmt.Sprintf("using %s as proxy to fetch user content", config.UserContentRequestProxy))
		proxyURL, err := url.Parse(config.UserContentRequestProxy)
		if err != nil {
			logger.FatalLog(fmt.Sprintf("USER_CONTENT_REQUEST_PROXY set but invalid: %s", config.UserContentRequestProxy))
		}
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	} else {
		transport = &http.Transport{}
	}

	protection := ssrf.GetGlobalProtection()
	safeTransport := protection.CreateSafeTransport(ctx, transport)

	userContentTransportMu.Lock()
	userContentTransport = safeTransport
	userContentTransportMu.Unlock()

	UserContentRequestHTTPClient = &http.Client{
		Transport:     safeTransport,
		Timeout:       time.Second * time.Duration(config.UserContentRequestTimeout),
		CheckRedirect: protection.CheckRedirect(ctx),
	}
}

func CloseUserContentIdleConnections() {
	userContentTransportMu.RLock()
	transport := userContentTransport
	userContentTransportMu.RUnlock()

	if transport != nil {
		transport.CloseIdleConnections()
	}
}

func initRelayClient(ctx context.Context) {
	var transport http.RoundTripper
	if config.RelayProxy != "" {
		logger.SysLog(fmt.Sprintf("using %s as api relay proxy", config.RelayProxy))
		proxyURL, err := url.Parse(config.RelayProxy)
		if err != nil {
			logger.FatalLog(fmt.Sprintf("USER_CONTENT_REQUEST_PROXY set but invalid: %s", config.UserContentRequestProxy))
		}
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	if config.RelayTimeout == 0 {
		HTTPClient = &http.Client{
			Transport: transport,
		}
	} else {
		HTTPClient = &http.Client{
			Timeout:   time.Duration(config.RelayTimeout) * time.Second,
			Transport: transport,
		}
	}

	ImpatientHTTPClient = &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	}
}
