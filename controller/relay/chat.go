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

type relayChat struct {
	relayBase
	chatRequest types.ChatCompletionRequest
}

func NewRelayChat(c *gin.Context) *relayChat {
	relay := &relayChat{}
	relay.c = c
	return relay
}

func (r *relayChat) setRequest() error {
	if err := common.UnmarshalBodyReusable(r.c, &r.chatRequest); err != nil {
		return err
	}

	if r.chatRequest.MaxTokens < 0 || r.chatRequest.MaxTokens > math.MaxInt32/2 {
		return errors.New("max_tokens is invalid")
	}

	r.originalModel = r.chatRequest.Model

	return nil
}

func (r *relayChat) getPromptTokens() (int, error) {
	return common.CountTokenMessages(r.chatRequest.Messages, r.modelName), nil
}

func (r *relayChat) send() (err *types.OpenAIErrorWithStatusCode, done bool) {
	chatProvider, ok := r.provider.(providersBase.ChatInterface)
	if !ok {
		err = common.StringErrorWrapper("channel not implemented", "channel_error", http.StatusServiceUnavailable)
		done = true
		return
	}

	r.chatRequest.Model = r.modelName

	if r.chatRequest.Stream {
		var response requester.StreamReaderInterface[string]
		response, err = chatProvider.CreateChatCompletionStream(&r.chatRequest)
		if err != nil {
			return
		}

		err = responseStreamClient(r.c, response)
	} else {
		var response *types.ChatCompletionResponse
		response, err = chatProvider.CreateChatCompletion(&r.chatRequest)
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
