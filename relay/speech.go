package relay

import (
	"net/http"
	"one-api/common"
	providersBase "one-api/providers/base"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

type relaySpeech struct {
	relayBase
	request types.SpeechAudioRequest
}

func NewRelaySpeech(c *gin.Context) *relaySpeech {
	relay := &relaySpeech{}
	relay.c = c
	return relay
}

func (r *relaySpeech) setRequest() error {
	if err := common.UnmarshalBodyReusable(r.c, &r.request); err != nil {
		return err
	}

	r.originalModel = r.request.Model

	return nil
}

func (r *relaySpeech) getPromptTokens() (int, error) {
	return len(r.request.Input), nil
}

func (r *relaySpeech) send() (err *types.OpenAIErrorWithStatusCode, done bool) {
	provider, ok := r.provider.(providersBase.SpeechInterface)
	if !ok {
		err = common.StringErrorWrapper("channel not implemented", "channel_error", http.StatusServiceUnavailable)
		done = true
		return
	}

	r.request.Model = r.modelName

	response, err := provider.CreateSpeech(&r.request)
	if err != nil {
		return
	}
	err = responseMultipart(r.c, response)

	if err != nil {
		done = true
	}

	return
}
