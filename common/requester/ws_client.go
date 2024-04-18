package requester

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"one-api/common"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/net/proxy"
)

func GetWSClient(proxyAddr string) *websocket.Dialer {
	dialer := &websocket.Dialer{
		HandshakeTimeout: time.Duration(common.GetOrDefault("connect_timeout", 5)) * time.Second,
	}

	if proxyAddr != "" {
		err := setWSProxy(dialer, proxyAddr)
		if err != nil {
			common.SysError(err.Error())
			return dialer
		}
	}

	return dialer
}

func setWSProxy(dialer *websocket.Dialer, proxyAddr string) error {
	proxyURL, err := url.Parse(proxyAddr)
	if err != nil {
		return fmt.Errorf("error parsing proxy address: %w", err)
	}

	switch proxyURL.Scheme {
	case "http", "https":
		dialer.Proxy = http.ProxyURL(proxyURL)
	case "socks5":
		socks5Proxy, err := proxy.SOCKS5("tcp", proxyURL.Host, nil, proxy.Direct)
		if err != nil {
			return fmt.Errorf("error creating socks5 dialer: %w", err)
		}
		dialer.NetDial = func(network, addr string) (net.Conn, error) {
			return socks5Proxy.Dial(network, addr)
		}
	default:
		return fmt.Errorf("unsupported proxy scheme: %s", proxyURL.Scheme)
	}

	return nil
}
