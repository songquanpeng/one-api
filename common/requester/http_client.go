package requester

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"one-api/common"
	"time"

	"golang.org/x/net/proxy"
)

type ContextKey string

const ProxyHTTPAddrKey ContextKey = "proxyHttpAddr"
const ProxySock5AddrKey ContextKey = "proxySock5Addr"

func proxyFunc(req *http.Request) (*url.URL, error) {
	proxyAddr := req.Context().Value(ProxyHTTPAddrKey)
	if proxyAddr == nil {
		return nil, nil
	}

	proxyURL, err := url.Parse(proxyAddr.(string))
	if err != nil {
		return nil, fmt.Errorf("error parsing proxy address: %w", err)
	}

	switch proxyURL.Scheme {
	case "http", "https":
		return proxyURL, nil
	}

	return nil, fmt.Errorf("unsupported proxy scheme: %s", proxyURL.Scheme)
}

func socks5ProxyFunc(ctx context.Context, network, addr string) (net.Conn, error) {
	// 设置TCP超时
	dialer := &net.Dialer{
		Timeout:   time.Duration(common.GetOrDefault("connect_timeout", 5)) * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// 从上下文中获取代理地址
	proxyAddr, ok := ctx.Value(ProxySock5AddrKey).(string)
	if !ok {
		return dialer.DialContext(ctx, network, addr)
	}

	proxyURL, err := url.Parse(proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("error parsing proxy address: %w", err)
	}

	proxyDialer, err := proxy.SOCKS5("tcp", proxyURL.Host, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("error creating socks5 dialer: %w", err)
	}

	return proxyDialer.Dial(network, addr)
}

var HTTPClient *http.Client

func InitHttpClient() {
	trans := &http.Transport{
		DialContext: socks5ProxyFunc,
		Proxy:       proxyFunc,
	}

	HTTPClient = &http.Client{
		Transport: trans,
	}

	relayTimeout := common.GetOrDefault("relay_timeout", 600)
	if relayTimeout != 0 {
		HTTPClient.Timeout = time.Duration(relayTimeout) * time.Second
	}
}
