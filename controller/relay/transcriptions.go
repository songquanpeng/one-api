package relay

import (
	"net/http"
	"one-api/common"
	providersBase "one-api/providers/base"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

type relayTranscriptions struct {
	relayBase
	request types.AudioRequest
}

func NewRelayTranscriptions(c *gin.Context) *relayTranscriptions {
	relay := &relayTranscriptions{}
	relay.c = c
	return relay
}

func (r *relayTranscriptions) setRequest() error {
	if err := common.UnmarshalBodyReusable(r.c, &r.request); err != nil {
		return err
	}

	r.originalModel = r.request.Model

	return nil
}

func (r *relayTranscriptions) getPromptTokens() (int, error) {
	return 0, nil
}

func (r *relayTranscriptions) send() (err *types.OpenAIErrorWithStatusCode, done bool) {
	provider, ok := r.provider.(providersBase.TranscriptionsInterface)
	if !ok {
		err = common.StringErrorWrapper("channel not implemented", "channel_error", http.StatusServiceUnavailable)
		done = true
		return
	}

	r.request.Model = r.modelName

	response, err := provider.CreateTranscriptions(&r.request)
	if err != nil {
		return
	}
	err = responseCustom(r.c, response)

	if err != nil {
		done = true
	}

	return
}
