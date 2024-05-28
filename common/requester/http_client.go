package requester

import (
	"net/http"
	"one-api/common/utils"
	"time"
)

var HTTPClient *http.Client

func InitHttpClient() {
	trans := &http.Transport{
		DialContext: utils.Socks5ProxyFunc,
		Proxy:       utils.ProxyFunc,
	}

	HTTPClient = &http.Client{
		Transport: trans,
	}

	relayTimeout := utils.GetOrDefault("relay_timeout", 600)
	if relayTimeout != 0 {
		HTTPClient.Timeout = time.Duration(relayTimeout) * time.Second
	}
}
