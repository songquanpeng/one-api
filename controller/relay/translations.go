package relay

import (
	"net/http"
	"one-api/common"
	providersBase "one-api/providers/base"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

type relayTranslations struct {
	relayBase
	request types.AudioRequest
}

func NewRelayTranslations(c *gin.Context) *relayTranslations {
	relay := &relayTranslations{}
	relay.c = c
	return relay
}

func (r *relayTranslations) setRequest() error {
	if err := common.UnmarshalBodyReusable(r.c, &r.request); err != nil {
		return err
	}

	r.originalModel = r.request.Model

	return nil
}

func (r *relayTranslations) getPromptTokens() (int, error) {
	return 0, nil
}

func (r *relayTranslations) send() (err *types.OpenAIErrorWithStatusCode, done bool) {
	provider, ok := r.provider.(providersBase.TranslationInterface)
	if !ok {
		err = common.StringErrorWrapper("channel not implemented", "channel_error", http.StatusServiceUnavailable)
		done = true
		return
	}

	r.request.Model = r.modelName

	response, err := provider.CreateTranslation(&r.request)
	if err != nil {
		return
	}
	err = responseCustom(r.c, response)

	if err != nil {
		done = true
	}

	return
}
