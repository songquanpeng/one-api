package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
    "io/ioutil"
    "context"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/apitype"
	"github.com/songquanpeng/one-api/relay/billing"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

func RelayTextHelper(c *gin.Context) *model.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := meta.GetByContext(c)

    // Read the original request body
    bodyBytes, err := ioutil.ReadAll(c.Request.Body)
    if err != nil {
        logger.Errorf(ctx, "Failed to read request body: %s", err.Error())
        return openai.ErrorWrapper(err, "invalid_request_body", http.StatusBadRequest)
    }

    // Restore the request body for `getAndValidateTextRequest`
    c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

    // Call `getAndValidateTextRequest`
    textRequest, err := getAndValidateTextRequest(c, meta.Mode)
    if err != nil {
        logger.Errorf(ctx, "getAndValidateTextRequest failed: %s", err.Error())
        return openai.ErrorWrapper(err, "invalid_text_request", http.StatusBadRequest)
    }
    meta.IsStream = textRequest.Stream

    // Parse the request body into a map
    var rawRequest map[string]interface{}
    if err := json.Unmarshal(bodyBytes, &rawRequest); err != nil {
        logger.Errorf(ctx, "Failed to parse request body into map: %s", err.Error())
        return openai.ErrorWrapper(err, "invalid_json", http.StatusBadRequest)
    }

    // Apply parameter overrides
    applyParameterOverrides(ctx, meta, textRequest, rawRequest)

	// map model name
	meta.OriginModelName = textRequest.Model
	textRequest.Model, _ = getMappedModelName(textRequest.Model, meta.ModelMapping)
	meta.ActualModelName = textRequest.Model
	// get model ratio & group ratio
	modelRatio := billingratio.GetModelRatio(textRequest.Model, meta.ChannelType)
	groupRatio := billingratio.GetGroupRatio(meta.Group)
	ratio := modelRatio * groupRatio
	// pre-consume quota
	promptTokens := getPromptTokens(textRequest, meta.Mode)
	meta.PromptTokens = promptTokens
	preConsumedQuota, bizErr := preConsumeQuota(ctx, textRequest, promptTokens, ratio, meta)
	if bizErr != nil {
		logger.Warnf(ctx, "preConsumeQuota failed: %+v", *bizErr)
		return bizErr
	}

	adaptor := relay.GetAdaptor(meta.APIType)
	if adaptor == nil {
		return openai.ErrorWrapper(fmt.Errorf("invalid api type: %d", meta.APIType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(meta)

	// get request body
	requestBody, err := getRequestBody(c, meta, textRequest, adaptor)
	if err != nil {
		return openai.ErrorWrapper(err, "convert_request_failed", http.StatusInternalServerError)
	}

	// do request
	resp, err := adaptor.DoRequest(c, meta, requestBody)
	if err != nil {
		logger.Errorf(ctx, "DoRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}
	if isErrorHappened(meta, resp) {
		billing.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
		return RelayErrorHandler(resp)
	}

	// do response
	usage, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		logger.Errorf(ctx, "respErr is not nil: %+v", respErr)
		billing.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
		return respErr
	}
	// post-consume quota
	go postConsumeQuota(ctx, usage, meta, textRequest, ratio, preConsumedQuota, modelRatio, groupRatio)
	return nil
}

func getRequestBody(c *gin.Context, meta *meta.Meta, textRequest *model.GeneralOpenAIRequest, adaptor adaptor.Adaptor) (io.Reader, error) {
	if meta.APIType == apitype.OpenAI && meta.OriginModelName == meta.ActualModelName && meta.ChannelType != channeltype.Baichuan {
		// no need to convert request for openai
		return c.Request.Body, nil
	}

	// get request body
	var requestBody io.Reader
	convertedRequest, err := adaptor.ConvertRequest(c, meta.Mode, textRequest)
	if err != nil {
		logger.Debugf(c.Request.Context(), "converted request failed: %s\n", err.Error())
		return nil, err
	}
	jsonData, err := json.Marshal(convertedRequest)
	if err != nil {
		logger.Debugf(c.Request.Context(), "converted request json_marshal_failed: %s\n", err.Error())
		return nil, err
	}
	logger.Debugf(c.Request.Context(), "converted request: \n%s", string(jsonData))
	requestBody = bytes.NewBuffer(jsonData)
	return requestBody, nil
}

func applyParameterOverrides(ctx context.Context, meta *meta.Meta, textRequest *model.GeneralOpenAIRequest, rawRequest map[string]interface{}) {
    if meta.ParamsOverride != nil {
        modelName := meta.OriginModelName
        if overrideParams, exists := meta.ParamsOverride[modelName]; exists {
            logger.Infof(ctx, "Applying parameter overrides for model %s on channel %d", modelName, meta.ChannelId)
            for key, value := range overrideParams {
                if _, userSpecified := rawRequest[key]; !userSpecified {
                    // Apply the override since the user didn't specify this parameter
                    switch key {
                    case "temperature":
                        if v, ok := value.(float64); ok {
                            textRequest.Temperature = v
                        } else if v, ok := value.(int); ok {
                            textRequest.Temperature = float64(v)
                        }
                    case "max_tokens":
                        if v, ok := value.(float64); ok {
                            textRequest.MaxTokens = int(v)
                        } else if v, ok := value.(int); ok {
                            textRequest.MaxTokens = v
                        }
                    case "top_p":
                        if v, ok := value.(float64); ok {
                            textRequest.TopP = v
                        } else if v, ok := value.(int); ok {
                            textRequest.TopP = float64(v)
                        }
                    case "frequency_penalty":
                        if v, ok := value.(float64); ok {
                            textRequest.FrequencyPenalty = v
                        } else if v, ok := value.(int); ok {
                            textRequest.FrequencyPenalty = float64(v)
                        }
                    case "presence_penalty":
                        if v, ok := value.(float64); ok {
                            textRequest.PresencePenalty = v
                        } else if v, ok := value.(int); ok {
                            textRequest.PresencePenalty = float64(v)
                        }
                    case "stop":
                        textRequest.Stop = value
                    case "n":
                        if v, ok := value.(float64); ok {
                            textRequest.N = int(v)
                        } else if v, ok := value.(int); ok {
                            textRequest.N = v
                        }
                    case "stream":
                        if v, ok := value.(bool); ok {
                            textRequest.Stream = v
                        }
                    case "num_ctx":
                        if v, ok := value.(float64); ok {
                            textRequest.NumCtx = int(v)
                        } else if v, ok := value.(int); ok {
                            textRequest.NumCtx = v
                        }
                    // Handle other parameters as needed
                    default:
                        logger.Warnf(ctx, "Unknown parameter override key: %s", key)
                    }
                }
            }
        }
    }
}