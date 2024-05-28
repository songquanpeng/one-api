package requester

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"one-api/common/logger"
	"one-api/common/utils"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/net/proxy"
)

func GetWSClient(proxyAddr string) *websocket.Dialer {
	dialer := &websocket.Dialer{
		HandshakeTimeout: time.Duration(utils.GetOrDefault("connect_timeout", 5)) * time.Second,
	}

	if proxyAddr != "" {
		err := setWSProxy(dialer, proxyAddr)
		if err != nil {
			logger.SysError(err.Error())
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
		proxyDialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			return fmt.Errorf("error creating proxy dialer: %w", err)
		}
		originalNetDial := dialer.NetDial
		dialer.NetDial = func(network, addr string) (net.Conn, error) {
			if originalNetDial != nil {
				return originalNetDial(network, addr)
			}
			return proxyDialer.Dial(network, addr)
		}
	default:
		return fmt.Errorf("unsupported proxy scheme: %s", proxyURL.Scheme)
	}

	return nil
}
