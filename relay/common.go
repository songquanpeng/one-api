package relay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/common/requester"
	"one-api/controller"
	"one-api/model"
	"one-api/providers"
	providersBase "one-api/providers/base"
	"one-api/types"
	"strings"

	"github.com/gin-gonic/gin"
)

func Path2Relay(c *gin.Context, path string) RelayBaseInterface {
	if strings.HasPrefix(path, "/v1/chat/completions") {
		return NewRelayChat(c)
	} else if strings.HasPrefix(path, "/v1/completions") {
		return NewRelayCompletions(c)
	} else if strings.HasPrefix(path, "/v1/embeddings") {
		return NewRelayEmbeddings(c)
	} else if strings.HasPrefix(path, "/v1/moderations") {
		return NewRelayModerations(c)
	} else if strings.HasPrefix(path, "/v1/images/generations") {
		return NewRelayImageGenerations(c)
	} else if strings.HasPrefix(path, "/v1/images/edits") {
		return NewRelayImageEdits(c)
	} else if strings.HasPrefix(path, "/v1/images/variations") {
		return NewRelayImageVariations(c)
	} else if strings.HasPrefix(path, "/v1/audio/speech") {
		return NewRelaySpeech(c)
	} else if strings.HasPrefix(path, "/v1/audio/transcriptions") {
		return NewRelayTranscriptions(c)
	} else if strings.HasPrefix(path, "/v1/audio/translations") {
		return NewRelayTranslations(c)
	}

	return nil
}

func getProvider(c *gin.Context, modeName string) (provider providersBase.ProviderInterface, newModelName string, fail error) {
	channel, fail := fetchChannel(c, modeName)
	if fail != nil {
		return
	}
	c.Set("channel_id", channel.Id)

	provider = providers.GetProvider(channel, c)
	if provider == nil {
		fail = errors.New("channel not found")
		return
	}
	provider.SetOriginalModel(modeName)

	newModelName, fail = provider.ModelMappingHandler(modeName)
	if fail != nil {
		return
	}

	return
}

func fetchChannel(c *gin.Context, modelName string) (channel *model.Channel, fail error) {
	channelId := c.GetInt("specific_channel_id")
	if channelId > 0 {
		return fetchChannelById(channelId)
	}

	return fetchChannelByModel(c, modelName)
}

func fetchChannelById(channelId int) (*model.Channel, error) {
	channel, err := model.GetChannelById(channelId, true)
	if err != nil {
		return nil, errors.New("无效的渠道 Id")
	}
	if channel.Status != common.ChannelStatusEnabled {
		return nil, errors.New("该渠道已被禁用")
	}

	return channel, nil
}

func fetchChannelByModel(c *gin.Context, modelName string) (*model.Channel, error) {
	group := c.GetString("group")
	channel, err := model.ChannelGroup.Next(group, modelName)
	if err != nil {
		message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", group, modelName)
		if channel != nil {
			common.SysError(fmt.Sprintf("渠道不存在：%d", channel.Id))
			message = "数据库一致性已被破坏，请联系管理员"
		}
		return nil, errors.New(message)
	}

	return channel, nil
}

func responseJsonClient(c *gin.Context, data interface{}) *types.OpenAIErrorWithStatusCode {
	// 将data转换为 JSON
	responseBody, err := json.Marshal(data)
	if err != nil {
		return common.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError)
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(http.StatusOK)
	_, err = c.Writer.Write(responseBody)
	if err != nil {
		return common.ErrorWrapper(err, "write_response_body_failed", http.StatusInternalServerError)
	}

	return nil
}

func responseStreamClient(c *gin.Context, stream requester.StreamReaderInterface[string]) *types.OpenAIErrorWithStatusCode {
	requester.SetEventStreamHeaders(c)
	dataChan, errChan := stream.Recv()

	defer stream.Close()
	c.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			fmt.Fprintln(w, "data: "+data+"\n")
			return true
		case err := <-errChan:
			if !errors.Is(err, io.EOF) {
				fmt.Fprintln(w, "data: "+err.Error()+"\n")
			}

			fmt.Fprintln(w, "data: [DONE]")
			return false
		}
	})

	return nil
}

func responseMultipart(c *gin.Context, resp *http.Response) *types.OpenAIErrorWithStatusCode {
	defer resp.Body.Close()

	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}

	c.Writer.WriteHeader(resp.StatusCode)

	_, err := io.Copy(c.Writer, resp.Body)
	if err != nil {
		return common.ErrorWrapper(err, "write_response_body_failed", http.StatusInternalServerError)
	}

	return nil
}

func responseCustom(c *gin.Context, response *types.AudioResponseWrapper) *types.OpenAIErrorWithStatusCode {
	for k, v := range response.Headers {
		c.Writer.Header().Set(k, v)
	}
	c.Writer.WriteHeader(http.StatusOK)

	_, err := c.Writer.Write(response.Body)
	if err != nil {
		return common.ErrorWrapper(err, "write_response_body_failed", http.StatusInternalServerError)
	}

	return nil
}

func shouldRetry(c *gin.Context, statusCode int) bool {
	channelId := c.GetInt("specific_channel_id")
	if channelId > 0 {
		return false
	}
	if statusCode == http.StatusTooManyRequests {
		return true
	}
	if statusCode/100 == 5 {
		return true
	}
	if statusCode == http.StatusBadRequest {
		return false
	}
	if statusCode/100 == 2 {
		return false
	}
	return true
}

func processChannelRelayError(ctx context.Context, channelId int, channelName string, err *types.OpenAIErrorWithStatusCode) {
	common.LogError(ctx, fmt.Sprintf("relay error (channel #%d(%s)): %s", channelId, channelName, err.Message))
	if controller.ShouldDisableChannel(&err.OpenAIError, err.StatusCode) {
		controller.DisableChannel(channelId, channelName, err.Message)
	}
}
