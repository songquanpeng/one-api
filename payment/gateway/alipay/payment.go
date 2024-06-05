package alipay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/smartwalle/alipay/v3"
	"net/http"
	"net/url"
	sysconfig "one-api/common/config"
	"one-api/payment/types"
	"strconv"
)

type Alipay struct{}

type AlipayConfig struct {
	AppID      string  `json:"app_id"`
	PrivateKey string  `json:"private_key"`
	PublicKey  string  `json:"public_key"`
	PayType    PayType `json:"pay_type"`
}

var client *alipay.Client

const isProduction bool = true

func (a *Alipay) Name() string {
	return "支付宝当面付"
}

func (a *Alipay) InitClient(config *AlipayConfig) error {
	var err error
	client, err = alipay.New(config.AppID, config.PrivateKey, isProduction)
	if err != nil {
		return err
	}
	return client.LoadAliPayPublicKey(config.PublicKey)
}

func (a *Alipay) Pay(config *types.PayConfig, gatewayConfig string) (*types.PayRequest, error) {
	alipayConfig, err := getAlipayConfig(gatewayConfig)
	if err != nil {
		return nil, err
	}

	if client == nil {
		err := a.InitClient(alipayConfig)
		if err != nil {
			return nil, err
		}
	}

	if alipayConfig.PayType != PagePay {
		var p = alipay.TradePreCreate{}
		p.OutTradeNo = config.TradeNo
		p.TotalAmount = strconv.FormatFloat(config.Money, 'f', 2, 64)
		p.Subject = sysconfig.SystemName + "-Token充值:" + p.TotalAmount
		p.NotifyURL = config.NotifyURL
		p.ReturnURL = config.ReturnURL
		ctx := context.Background()
		alipayRes, err := client.TradePreCreate(ctx, p)
		if err != nil {
			return nil, fmt.Errorf("alipay trade precreate failed: %s", alipayRes.Msg)
		}
		if !alipayRes.IsSuccess() {
			return nil, fmt.Errorf("alipay trade precreate failed: %s", alipayRes.Msg)
		}
		if alipayRes.Code != "10000" {
			return nil, fmt.Errorf("alipay trade precreate failed: %s", alipayRes.Msg)
		}
		payRequest := &types.PayRequest{
			Type: 2,
			Data: types.PayRequestData{
				URL:    alipayRes.QRCode,
				Method: http.MethodGet,
			},
		}
		return payRequest, nil
	} else {
		var p = alipay.TradePagePay{}
		p.OutTradeNo = config.TradeNo
		p.TotalAmount = strconv.FormatFloat(config.Money, 'f', 2, 64)
		p.Subject = sysconfig.SystemName + "-Token充值:" + p.TotalAmount
		p.NotifyURL = config.NotifyURL
		p.ReturnURL = config.ReturnURL
		alipayRes, err := client.TradePagePay(p)
		if err != nil {
			return nil, fmt.Errorf("alipay trade precreate failed: %s", err.Error())
		}
		payUrl, parms, err := extractURLAndParams(alipayRes.String())
		if err != nil {
			return nil, fmt.Errorf("alipay trade precreate failed: %s", err.Error())
		}
		payRequest := &types.PayRequest{
			Type: 1,
			Data: types.PayRequestData{
				URL:    payUrl,
				Params: parms,
				Method: http.MethodGet,
			},
		}
		return payRequest, nil
	}

}

func (a *Alipay) HandleCallback(c *gin.Context, gatewayConfig string) (*types.PayNotify, error) {
	// 获取通知参数
	params := c.Request.URL.Query()
	if err := c.Request.ParseForm(); err != nil {
		c.Writer.Write([]byte("failure"))
		return nil, fmt.Errorf("Alipay params failed: %v", err)
	}
	for k, v := range c.Request.PostForm {
		params[k] = v
	}
	// 验证通知签名
	if err := client.VerifySign(params); err != nil {
		c.Writer.Write([]byte("failure"))
		return nil, fmt.Errorf("Alipay Signature verification failed: %v", err)
	}
	//解析通知内容
	var noti, err = client.DecodeNotification(params)
	if err != nil {
		c.Writer.Write([]byte("failure"))
		return nil, fmt.Errorf("Alipay Error decoding notification: %v", err)
	}

	if noti.TradeStatus == alipay.TradeStatusSuccess {
		payNotify := &types.PayNotify{
			TradeNo:   noti.OutTradeNo,
			GatewayNo: noti.TradeNo,
		}
		alipay.ACKNotification(c.Writer)
		return payNotify, nil
	}
	c.Writer.Write([]byte("failure"))
	return nil, fmt.Errorf("trade status not success")
}

func getAlipayConfig(gatewayConfig string) (*AlipayConfig, error) {
	var alipayConfig AlipayConfig
	if err := json.Unmarshal([]byte(gatewayConfig), &alipayConfig); err != nil {
		return nil, errors.New("config error")
	}

	return &alipayConfig, nil
}

// extractURLAndParams 从给定的原始 URL 中提取网址和参数，并将参数转换为 map[string]string
func extractURLAndParams(rawURL string) (string, map[string]string, error) {
	// 解析 URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", nil, err
	}

	// 提取网址
	baseURL := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, parsedURL.Path)

	// 提取参数并转换成 map[string]string
	params := parsedURL.Query()
	paramMap := make(map[string]string)
	for key, values := range params {
		// 由于 URL 参数可能有多个值，这里只取第一个值
		paramMap[key] = values[0]
	}

	return baseURL, paramMap, nil
}
