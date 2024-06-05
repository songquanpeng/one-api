package types

// 支付网关的通用配置
type PayConfig struct {
	NotifyURL string  `json:"notify_url"`
	ReturnURL string  `json:"return_url"`
	TradeNo   string  `json:"trade_no"`
	Money     float64 `json:"money"`
}

// 请求支付时的数据结构
type PayRequest struct {
	Type int            `json:"type"` // 支付类型 1 url 2 qrcode
	Data PayRequestData `json:"data"`
}

type PayRequestData struct {
	URL    string `json:"url"`
	Method string `json:"method,omitempty"`
	Params any    `json:"params,omitempty"`
}

// 支付回调时的数据结构
type PayNotify struct {
	TradeNo   string `json:"trade_no"`
	GatewayNo string `json:"gateway_no"`
}
