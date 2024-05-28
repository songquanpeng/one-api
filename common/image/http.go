package image

import (
	"net/http"
	"one-api/common/config"
	"one-api/common/utils"
	"time"
)

var ImageHttpClients = &http.Client{
	Transport: &http.Transport{
		DialContext: utils.Socks5ProxyFunc,
		Proxy:       utils.ProxyFunc,
	},
	Timeout: 15 * time.Second,
}

func requestImage(url, method string) (*http.Response, error) {
	res, err := utils.RequestBuilder(utils.SetProxy(config.ChatImageRequestProxy, nil), method, url, nil, nil)

	if err != nil {
		return nil, err
	}

	return ImageHttpClients.Do(res)
}
