package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/billing"
	"github.com/songquanpeng/one-api/relay/model"
)

type Options struct {
	EnableBilling bool
}

// RelayInstance is the interface for relay controller
type RelayInstance interface {
	RelayTextHelper(c *gin.Context) *model.ErrorWithStatusCode
	RelayImageHelper(c *gin.Context, relayMode int) *model.ErrorWithStatusCode
	RelayAudioHelper(c *gin.Context, relayMode int) *model.ErrorWithStatusCode
}

type defaultRelay struct {
	billing.Bookkeeper
}

func NewRelayInstance(opts Options) RelayInstance {
	relay := &defaultRelay{}
	if opts.EnableBilling {
		relay.Bookkeeper = billing.NewBookkeeper()
	}
	return relay
}
