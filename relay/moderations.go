package relay

import (
	"net/http"
	"one-api/common"
	providersBase "one-api/providers/base"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

type relayModerations struct {
	relayBase
	request types.ModerationRequest
}

func NewRelayModerations(c *gin.Context) *relayModerations {
	relay := &relayModerations{}
	relay.c = c
	return relay
}

func (r *relayModerations) setRequest() error {
	if err := common.UnmarshalBodyReusable(r.c, &r.request); err != nil {
		return err
	}

	if r.request.Model == "" {
		r.request.Model = "text-moderation-stable"
	}

	r.originalModel = r.request.Model

	return nil
}

func (r *relayModerations) getPromptTokens() (int, error) {
	return common.CountTokenInput(r.request.Input, r.modelName), nil
}

func (r *relayModerations) send() (err *types.OpenAIErrorWithStatusCode, done bool) {
	provider, ok := r.provider.(providersBase.ModerationInterface)
	if !ok {
		err = common.StringErrorWrapper("channel not implemented", "channel_error", http.StatusServiceUnavailable)
		done = true
		return
	}

	r.request.Model = r.modelName

	response, err := provider.CreateModeration(&r.request)
	if err != nil {
		return
	}
	err = responseJsonClient(r.c, response)

	if err != nil {
		done = true
	}

	return
}
