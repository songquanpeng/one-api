package epay

type PayType string

var (
	EpayPay PayType = ""       // 网关
	Alipay  PayType = "alipay" // 支付宝
	Wechat  PayType = "wxpay"  // 微信
	QQ      PayType = "qqpay"  // QQ
	Bank    PayType = "bank"   // 银行
	JD      PayType = "jdpay"  // 京东
	PayPal  PayType = "paypal" // PayPal
	USDT    PayType = "usdt"   // USDT
)

// type DeviceType string

// var (
// 	PC     DeviceType = "pc"     // PC
// 	Mobile DeviceType = "mobile" // 移动端
// )

const (
	FormArgsSignType   = "MD5"
	FormSubmitUrl      = "/submit.php"
	TradeStatusSuccess = "TRADE_SUCCESS"
)

type PayArgs struct {
	Type       PayType `json:"type,omitempty"`
	OutTradeNo string  `json:"out_trade_no"`
	NotifyUrl  string  `json:"notify_url"`
	ReturnUrl  string  `json:"return_url"`
	Name       string  `json:"name"`
	Money      string  `json:"money"`
}

type PaymentResult struct {
	Type        PayType `mapstructure:"type"`
	TradeNo     string  `mapstructure:"trade_no"`
	OutTradeNo  string  `mapstructure:"out_trade_no"`
	Name        string  `mapstructure:"name"`
	Money       string  `mapstructure:"money"`
	TradeStatus string  `mapstructure:"trade_status"`
}
