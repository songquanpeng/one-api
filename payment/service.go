package payment

import (
	"errors"
	"fmt"
	"one-api/common/config"
	"one-api/common/logger"
	"one-api/model"
	"one-api/payment/types"
	"strings"

	"github.com/gin-gonic/gin"
)

type PaymentService struct {
	Payment *model.Payment
	gateway PaymentProcessor
}

type PayMoney struct {
	Amount   float64
	Currency model.CurrencyType
}

func NewPaymentService(uuid string) (*PaymentService, error) {
	payment, err := model.GetPaymentByUUID(uuid)
	if err != nil {
		return nil, errors.New("payment not found")
	}

	gateway, ok := Gateways[payment.Type]
	if !ok {
		return nil, errors.New("payment gateway not found")
	}

	return &PaymentService{
		Payment: payment,
		gateway: gateway,
	}, nil
}

func (s *PaymentService) Pay(tradeNo string, amount float64) (*types.PayRequest, error) {
	config := &types.PayConfig{
		Money:     amount,
		TradeNo:   tradeNo,
		NotifyURL: s.getNotifyURL(),
		ReturnURL: s.getReturnURL(),
	}

	payRequest, err := s.gateway.Pay(config, s.Payment.Config)
	if err != nil {
		return nil, err
	}

	return payRequest, nil
}

func (s *PaymentService) HandleCallback(c *gin.Context, gatewayConfig string) (*types.PayNotify, error) {
	payNotify, err := s.gateway.HandleCallback(c, gatewayConfig)
	if err != nil {
		logger.SysError(fmt.Sprintf("%s payment callback error: %v", s.gateway.Name(), err))

	}

	return payNotify, err
}

func (s *PaymentService) getNotifyURL() string {
	notifyDomain := s.Payment.NotifyDomain
	if notifyDomain == "" {
		notifyDomain = config.ServerAddress
	}

	notifyDomain = strings.TrimSuffix(notifyDomain, "/")
	return fmt.Sprintf("%s/api/payment/notify/%s", notifyDomain, s.Payment.UUID)
}

func (s *PaymentService) getReturnURL() string {
	serverAdd := strings.TrimSuffix(config.ServerAddress, "/")
	return fmt.Sprintf("%s/panel/log", serverAdd)
}
