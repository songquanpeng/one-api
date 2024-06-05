package payment

import (
	"one-api/payment/gateway/alipay"
	"one-api/payment/gateway/epay"
	"one-api/payment/types"

	"github.com/gin-gonic/gin"
)

type PaymentProcessor interface {
	Name() string
	Pay(config *types.PayConfig, gatewayConfig string) (*types.PayRequest, error)
	HandleCallback(c *gin.Context, gatewayConfig string) (*types.PayNotify, error)
}

var Gateways = make(map[string]PaymentProcessor)

func init() {
	Gateways["epay"] = &epay.Epay{}
	Gateways["alipay"] = &alipay.Alipay{}
}
