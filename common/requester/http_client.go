package requester

import (
	"fmt"
	"net/http"
	"net/url"
	"one-api/common"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

type HTTPClient struct{}

var clientPool = &sync.Pool{
	New: func() interface{} {
		return &http.Client{}
	},
}

func (h *HTTPClient) getClientFromPool(proxyAddr string) *http.Client {
	client := clientPool.Get().(*http.Client)

	if common.RelayTimeout > 0 {
		client.Timeout = time.Duration(common.RelayTimeout) * time.Second
	}

	if proxyAddr != "" {
		err := h.setProxy(client, proxyAddr)
		if err != nil {
			common.SysError(err.Error())
			return client
		}
	}

	return client
}

func (h *HTTPClient) returnClientToPool(client *http.Client) {
	clientPool.Put(client)
}

func (h *HTTPClient) setProxy(client *http.Client, proxyAddr string) error {
	proxyURL, err := url.Parse(proxyAddr)
	if err != nil {
		return fmt.Errorf("error parsing proxy address: %w", err)
	}

	switch proxyURL.Scheme {
	case "http", "https":
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	case "socks5":
		dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, nil, proxy.Direct)
		if err != nil {
			return fmt.Errorf("error creating socks5 dialer: %w", err)
		}
		client.Transport = &http.Transport{
			Dial: dialer.Dial,
		}
	default:
		return fmt.Errorf("unsupported proxy scheme: %s", proxyURL.Scheme)
	}

	return nil
}
