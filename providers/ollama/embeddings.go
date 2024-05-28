package ollama

import (
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/types"
)

func (p *OllamaProvider) CreateEmbeddings(request *types.EmbeddingRequest) (*types.EmbeddingResponse, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeEmbeddings)
	if errWithCode != nil {
		return nil, errWithCode
	}
	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url, request.Model)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapper(nil, "invalid_ollama_config", http.StatusInternalServerError)
	}

	// 获取请求头
	headers := p.GetRequestHeaders()

	ollamaRequest := &EmbeddingRequest{
		Model:  request.Model,
		Prompt: request.ParseInputString(),
	}

	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(ollamaRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}
	defer req.Body.Close()

	ollamaResponse := &EmbeddingResponse{}

	// 发送请求
	var res *http.Response
	res, errWithCode = p.Requester.SendRequest(req, ollamaResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	errWithOP := errorHandle(&ollamaResponse.OllamaError)
	if errWithOP != nil {
		return nil, &types.OpenAIErrorWithStatusCode{
			OpenAIError: *errWithOP,
			StatusCode:  res.StatusCode,
		}
	}

	response := &types.EmbeddingResponse{
		Object: "list",
		Model:  request.Model,
		Data: []types.Embedding{{
			Object:    "embedding",
			Index:     0,
			Embedding: ollamaResponse.Embedding,
		}},
		Usage: &types.Usage{
			TotalTokens:      0,
			CompletionTokens: 0,
			PromptTokens:     0,
		},
	}

	*p.Usage = *response.Usage

	return response, nil
}
