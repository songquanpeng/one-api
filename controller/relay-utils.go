package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/common/requester"
	"one-api/model"
	"one-api/providers"
	providersBase "one-api/providers/base"
	"one-api/types"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func getProvider(c *gin.Context, modeName string) (provider providersBase.ProviderInterface, newModelName string, fail bool) {
	channel, fail := fetchChannel(c, modeName)
	if fail {
		return
	}

	provider = providers.GetProvider(channel, c)
	if provider == nil {
		common.AbortWithMessage(c, http.StatusNotImplemented, "channel not found")
		fail = true
		return
	}

	newModelName, err := provider.ModelMappingHandler(modeName)
	if err != nil {
		common.AbortWithMessage(c, http.StatusInternalServerError, err.Error())
		fail = true
		return
	}

	return
}

func GetValidFieldName(err error, obj interface{}) string {
	getObj := reflect.TypeOf(obj)
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			if f, exist := getObj.Elem().FieldByName(e.Field()); exist {
				return f.Name
			}
		}
	}
	return err.Error()
}

func fetchChannel(c *gin.Context, modelName string) (channel *model.Channel, fail bool) {
	channelId, ok := c.Get("channelId")
	if ok {
		channel, fail = fetchChannelById(c, channelId.(int))
		if fail {
			return
		}

	}
	channel, fail = fetchChannelByModel(c, modelName)
	if fail {
		return
	}

	c.Set("channel_id", channel.Id)

	return
}

func fetchChannelById(c *gin.Context, channelId any) (*model.Channel, bool) {
	id, err := strconv.Atoi(channelId.(string))
	if err != nil {
		common.AbortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
		return nil, true
	}
	channel, err := model.GetChannelById(id, true)
	if err != nil {
		common.AbortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
		return nil, true
	}
	if channel.Status != common.ChannelStatusEnabled {
		common.AbortWithMessage(c, http.StatusForbidden, "该渠道已被禁用")
		return nil, true
	}

	return channel, false
}

func fetchChannelByModel(c *gin.Context, modelName string) (*model.Channel, bool) {
	group := c.GetString("group")
	channel, err := model.CacheGetRandomSatisfiedChannel(group, modelName)
	if err != nil {
		message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", group, modelName)
		if channel != nil {
			common.SysError(fmt.Sprintf("渠道不存在：%d", channel.Id))
			message = "数据库一致性已被破坏，请联系管理员"
		}
		common.AbortWithMessage(c, http.StatusServiceUnavailable, message)
		return nil, true
	}

	return channel, false
}

func shouldDisableChannel(err *types.OpenAIError, statusCode int) bool {
	if !common.AutomaticDisableChannelEnabled {
		return false
	}
	if err == nil {
		return false
	}
	if statusCode == http.StatusUnauthorized {
		return true
	}
	if err.Type == "insufficient_quota" || err.Code == "invalid_api_key" || err.Code == "account_deactivated" {
		return true
	}
	return false
}

func shouldEnableChannel(err error, openAIErr *types.OpenAIError) bool {
	if !common.AutomaticEnableChannelEnabled {
		return false
	}
	if err != nil {
		return false
	}
	if openAIErr != nil {
		return false
	}
	return true
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
