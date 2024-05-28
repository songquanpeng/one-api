package cloudflareAI

import (
	"encoding/base64"
	"io"
	"net/http"
	"one-api/common"
	"one-api/common/storage"
	"one-api/common/utils"
	"one-api/types"
	"time"
)

func (p *CloudflareAIProvider) CreateImageGenerations(request *types.ImageRequest) (*types.ImageResponse, *types.OpenAIErrorWithStatusCode) {
	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(request.Model)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapper(nil, "invalid_cloudflare_ai_config", http.StatusInternalServerError)
	}

	// 获取请求头
	headers := p.GetRequestHeaders()
	cfRequest := convertFromIamgeOpenai(request)

	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(cfRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}
	defer req.Body.Close()

	resp, errWithCode := p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, errWithCode
	}

	defer resp.Body.Close()

	if resp.Header.Get("Content-Type") != "image/png" {
		return nil, common.StringErrorWrapper("invalid_image_response", "invalid_image_response", http.StatusInternalServerError)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, common.ErrorWrapper(err, "read_response_failed", http.StatusInternalServerError)
	}

	url := ""
	if request.ResponseFormat == "" || request.ResponseFormat == "url" {
		url = storage.Upload(body, utils.GetUUID()+".png")
	}

	openaiResponse := &types.ImageResponse{
		Created: time.Now().Unix(),
	}

	if url == "" {
		base64Image := base64.StdEncoding.EncodeToString(body)
		openaiResponse.Data = []types.ImageResponseDataInner{{B64JSON: base64Image}}
	} else {
		openaiResponse.Data = []types.ImageResponseDataInner{{URL: url}}
	}

	p.Usage.PromptTokens = 1000

	return openaiResponse, nil
}

func convertFromIamgeOpenai(request *types.ImageRequest) *ImageRequest {
	return &ImageRequest{
		Prompt: request.Prompt,
	}
}
