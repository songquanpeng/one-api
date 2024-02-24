package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/channel/openai"
	"github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/helper"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/util"
	"io"
	"net/http"
	"strings"
)

func RelayTextHelper(c *gin.Context) *model.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := util.GetRelayMeta(c)
	// get & validate textRequest
	textRequest, err := getAndValidateTextRequest(c, meta.Mode)
	if err != nil {
		logger.Errorf(ctx, "getAndValidateTextRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "invalid_text_request", http.StatusBadRequest)
	}

	memory := GetMemory(meta.UserId, meta.TokenName)
	reqMsg := textRequest.Messages[0]

	//fmt.Printf("----req msg %v \n", textRequest.Messages)
	//fmt.Printf("----memory%v \n", memory)
	//fmt.Println("-------------------------------------------------------------")

	memory = append(memory, textRequest.Messages...)
	textRequest.Messages = memory

	meta.IsStream = textRequest.Stream

	// map model name
	var isModelMapped bool
	meta.OriginModelName = textRequest.Model
	textRequest.Model, isModelMapped = util.GetMappedModelName(textRequest.Model, meta.ModelMapping)
	meta.ActualModelName = textRequest.Model
	// get model ratio & group ratio
	modelRatio := common.GetModelRatio(textRequest.Model)
	groupRatio := common.GetGroupRatio(meta.Group)
	ratio := modelRatio * groupRatio
	// pre-consume quota
	promptTokens := getPromptTokens(textRequest, meta.Mode)
	meta.PromptTokens = promptTokens
	preConsumedQuota, bizErr := preConsumeQuota(ctx, textRequest, promptTokens, ratio, meta)
	if bizErr != nil {
		logger.Warnf(ctx, "preConsumeQuota failed: %+v", *bizErr)
		return bizErr
	}

	adaptor := helper.GetAdaptor(meta.APIType)
	if adaptor == nil {
		return openai.ErrorWrapper(fmt.Errorf("invalid api type: %d", meta.APIType), "invalid_api_type", http.StatusBadRequest)
	}

	// get request body
	var requestBody io.Reader
	if meta.APIType == constant.APITypeOpenAI {
		// no need to convert request for openai
		if isModelMapped {
			jsonStr, err := json.Marshal(textRequest)
			if err != nil {
				return openai.ErrorWrapper(err, "json_marshal_failed", http.StatusInternalServerError)
			}
			requestBody = bytes.NewBuffer(jsonStr)
		} else {
			requestBody = c.Request.Body
		}
	} else {
		convertedRequest, err := adaptor.ConvertRequest(c, meta.Mode, textRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "convert_request_failed", http.StatusInternalServerError)
		}
		jsonData, err := json.Marshal(convertedRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "json_marshal_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonData)
	}

	// do request
	resp, err := adaptor.DoRequest(c, meta, requestBody)
	if err != nil {
		logger.Errorf(ctx, "DoRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}
	meta.IsStream = meta.IsStream || strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream")
	if resp.StatusCode != http.StatusOK {
		util.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
		return util.RelayErrorHandler(resp)
	}

	// do response
	usage, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		logger.Errorf(ctx, "respErr is not nil: %+v", respErr)
		util.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
		return respErr
	}

	if !meta.IsStream {
		//非流式，保存历史记录
		respMsg := adaptor.GetLastTextResp()
		SaveMemory(meta.UserId, meta.TokenName, respMsg, reqMsg)

	}
	// post-consume quota
	go postConsumeQuota(ctx, usage, meta, textRequest, ratio, preConsumedQuota, modelRatio, groupRatio)
	return nil
}

func SaveMemory(userId int, tokenName, resp string, req model.Message) {
	if len(resp) < 1 {
		return
	}

	msgs := []model.Message{}
	msgs = append(msgs, req)
	msgs = append(msgs, model.Message{Role: "assistant", Content: resp})

	v, _ := json.Marshal(&msgs)

	key := fmt.Sprintf("one_api_memory:%d:%s", userId, tokenName)

	common.RedisLPush(key, string(v))
}

func GetMemory(userId int, tokenName string) []model.Message {

	key := fmt.Sprintf("one_api_memory:%d:%s", userId, tokenName)

	ss := common.RedisLRange(key, 0, int64(config.MemoryMaxNum))
	var memory []model.Message

	i := len(ss) - 1

	for i >= 0 {
		s := ss[i]
		var msgItem []model.Message
		if e := json.Unmarshal([]byte(s), &msgItem); e != nil {
			continue
		}

		for _, v := range msgItem {
			if v.Content != nil && len(v.Content.(string)) > 0 {
				memory = append(memory, v)
			}
		}
		i--
	}
	return memory
}
