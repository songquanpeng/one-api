package bedrock

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"one-api/common/requester"
	"one-api/model"
	"one-api/providers/base"
	"one-api/types"
	"strings"
	"time"

	"one-api/providers/bedrock/category"
	"one-api/providers/bedrock/sigv4"
)

type BedrockProviderFactory struct{}

// 创建 BedrockProvider
func (f BedrockProviderFactory) Create(channel *model.Channel) base.ProviderInterface {

	bedrockProvider := &BedrockProvider{
		BaseProvider: base.BaseProvider{
			Config:    getConfig(),
			Channel:   channel,
			Requester: requester.NewHTTPRequester(*channel.Proxy, requestErrorHandle),
		},
	}

	getKeyConfig(bedrockProvider)

	return bedrockProvider
}

type BedrockProvider struct {
	base.BaseProvider
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Category        *category.Category
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         "https://bedrock-runtime.%s.amazonaws.com",
		ChatCompletions: "/model/%s/invoke",
	}
}

// 请求错误处理
func requestErrorHandle(resp *http.Response) *types.OpenAIError {
	bedrockError := &BedrockError{}
	err := json.NewDecoder(resp.Body).Decode(bedrockError)
	if err != nil {
		return nil
	}

	return errorHandle(bedrockError)
}

// 错误处理
func errorHandle(bedrockError *BedrockError) *types.OpenAIError {
	if bedrockError.Message == "" {
		return nil
	}
	return &types.OpenAIError{
		Message: bedrockError.Message,
		Type:    "Bedrock Error",
	}
}

func (p *BedrockProvider) GetFullRequestURL(requestURL string, modelName string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

	return fmt.Sprintf(baseURL+requestURL, p.Region, modelName)
}

func (p *BedrockProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)
	headers["Accept"] = "*/*"

	return headers
}

func getKeyConfig(bedrock *BedrockProvider) {
	keys := strings.Split(bedrock.Channel.Key, "|")
	if len(keys) < 3 {
		return
	}

	bedrock.Region = keys[0]
	bedrock.AccessKeyID = keys[1]
	bedrock.SecretAccessKey = keys[2]
	if len(keys) == 4 && keys[3] != "" {
		bedrock.SessionToken = keys[3]
	}
}

func (p *BedrockProvider) Sign(req *http.Request) error {
	var body []byte
	if req.Body == nil {
		body = []byte("")
	} else {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return errors.New("error getting request body: " + err.Error())
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
	}
	sig, err := sigv4.New(sigv4.WithCredential(p.AccessKeyID, p.SecretAccessKey, p.SessionToken), sigv4.WithRegionService(p.Region, awsService))
	if err != nil {
		return err
	}

	reqBodyHashHex := fmt.Sprintf("%x", sha256.Sum256(body))
	sig.Sign(req, reqBodyHashHex, sigv4.NewTime(time.Now()))

	return nil
}
