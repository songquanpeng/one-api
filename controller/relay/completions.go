package relay

import (
	"errors"
	"math"
	"net/http"
	"one-api/common"
	"one-api/common/requester"
	providersBase "one-api/providers/base"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

type relayCompletions struct {
	relayBase
	request types.CompletionRequest
}

func NewRelayCompletions(c *gin.Context) *relayCompletions {
	relay := &relayCompletions{}
	relay.c = c
	return relay
}

func (r *relayCompletions) setRequest() error {
	if err := common.UnmarshalBodyReusable(r.c, &r.request); err != nil {
		return err
	}

	if r.request.MaxTokens < 0 || r.request.MaxTokens > math.MaxInt32/2 {
		return errors.New("max_tokens is invalid")
	}

	r.originalModel = r.request.Model

	return nil
}

func (r *relayCompletions) getPromptTokens() (int, error) {
	return common.CountTokenInput(r.request.Prompt, r.modelName), nil
}

func (r *relayCompletions) send() (err *types.OpenAIErrorWithStatusCode, done bool) {
	provider, ok := r.provider.(providersBase.CompletionInterface)
	if !ok {
		err = common.StringErrorWrapper("channel not implemented", "channel_error", http.StatusServiceUnavailable)
		done = true
		return
	}

	r.request.Model = r.modelName

	if r.request.Stream {
		var response requester.StreamReaderInterface[string]
		response, err = provider.CreateCompletionStream(&r.request)
		if err != nil {
			return
		}

		err = responseStreamClient(r.c, response)
	} else {
		var response *types.CompletionResponse
		response, err = provider.CreateCompletion(&r.request)
		if err != nil {
			return
		}
		err = responseJsonClient(r.c, response)
	}

	if err != nil {
		done = true
	}

	return
}
