package vertexai

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/adaptor/gemini"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/relaymode"

	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

var ModelList = []string{
	"gemini-pro", "gemini-pro-vision",
	"gemini-exp-1206",
	"gemini-1.5-pro-001", "gemini-1.5-pro-002",
	"gemini-1.5-flash-001", "gemini-1.5-flash-002",
	"gemini-2.0-flash-exp", "gemini-2.0-flash-001",
	"gemini-2.0-flash-lite-preview-02-05",
	"gemini-2.0-flash-thinking-exp-01-21",
}

type Adaptor struct {
}

func (a *Adaptor) parseGeminiChatGenerationThinking(model string) (string, *gemini.ThinkingConfig) {
	thinkingConfig := &gemini.ThinkingConfig{
		IncludeThoughts: false,
		ThinkingBudget:  0,
	}
	modelName := model
	if strings.Contains(model, "?") {
		parts := strings.Split(model, "?")
		_modelName := parts[0]
		if len(parts) >= 2 {
			modelOptions, err := url.ParseQuery(parts[1])
			if err == nil && modelOptions != nil {
				modelName = _modelName
				hasThinkingFlag := modelOptions.Has("thinking")
				if hasThinkingFlag {
					thinkingConfig.IncludeThoughts = modelOptions.Get("thinking") == "1"
				}
				thinkingBudget := modelOptions.Get("thinking_budget")
				if thinkingBudget != "" {
					thinkingBudgetInt, err := strconv.Atoi(thinkingBudget)
					if err == nil {
						thinkingConfig.ThinkingBudget = thinkingBudgetInt
					}
				}
			}
		}
	}
	if strings.HasPrefix(modelName, "gemini-2.5") {
		// 目前2.5的模型支持传递thinking config，且默认开启了thinking，不希望进入thinking模式需要显式传递thinkingConfig来关闭
		return modelName, thinkingConfig
	} else {
		// 其他模型暂时不支持
		if thinkingConfig != nil && (thinkingConfig.IncludeThoughts || thinkingConfig.ThinkingBudget > 0) {
			// 为了后续一旦有其他模型支持了thinking，这里指定可以指定参数开启
			return modelName, thinkingConfig
		}
		return modelName, nil
	}
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	modelName, thinkingConfig := a.parseGeminiChatGenerationThinking(request.Model)
	request.Model = modelName
	geminiRequest := gemini.ConvertRequest(*request)
	if thinkingConfig != nil {
		geminiRequest.GenerationConfig.ThinkingConfig = thinkingConfig
	}
	c.Set(ctxkey.RequestModel, modelName)
	c.Set(ctxkey.ConvertedRequest, geminiRequest)
	return geminiRequest, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		var responseText string
		err, responseText = gemini.StreamHandler(c, resp)
		usage = openai.ResponseText2Usage(responseText, meta.ActualModelName, meta.PromptTokens)
	} else {
		switch meta.Mode {
		case relaymode.Embeddings:
			err, usage = gemini.EmbeddingHandler(c, resp)
		default:
			err, usage = gemini.Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
		}
	}
	return
}
