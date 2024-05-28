package requester

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/common/utils"
	"one-api/types"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type HttpErrorHandler func(*http.Response) *types.OpenAIError

type HTTPRequester struct {
	// requestBuilder    utils.RequestBuilder
	CreateFormBuilder func(io.Writer) FormBuilder
	ErrorHandler      HttpErrorHandler
	proxyAddr         string
	Context           context.Context
	IsOpenAI          bool
}

// NewHTTPRequester 创建一个新的 HTTPRequester 实例。
// proxyAddr: 是代理服务器的地址。
// errorHandler: 是一个错误处理函数，它接收一个 *http.Response 参数并返回一个 *types.OpenAIErrorResponse。
// 如果 errorHandler 为 nil，那么会使用一个默认的错误处理函数。
func NewHTTPRequester(proxyAddr string, errorHandler HttpErrorHandler) *HTTPRequester {
	return &HTTPRequester{
		CreateFormBuilder: func(body io.Writer) FormBuilder {
			return NewFormBuilder(body)
		},
		ErrorHandler: errorHandler,
		proxyAddr:    proxyAddr,
		Context:      context.Background(),
		IsOpenAI:     true,
	}
}

type requestOptions struct {
	body   any
	header http.Header
}

type requestOption func(*requestOptions)

func (r *HTTPRequester) setProxy() context.Context {
	return utils.SetProxy(r.proxyAddr, r.Context)
}

// 创建请求
func (r *HTTPRequester) NewRequest(method, url string, setters ...requestOption) (*http.Request, error) {
	args := &requestOptions{
		body:   nil,
		header: make(http.Header),
	}
	for _, setter := range setters {
		setter(args)
	}
	req, err := utils.RequestBuilder(r.setProxy(), method, url, args.body, args.header)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// 发送请求
func (r *HTTPRequester) SendRequest(req *http.Request, response any, outputResp bool) (*http.Response, *types.OpenAIErrorWithStatusCode) {
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, common.ErrorWrapper(err, "http_request_failed", http.StatusInternalServerError)
	}

	if !outputResp {
		defer resp.Body.Close()
	}

	// 处理响应
	if r.IsFailureStatusCode(resp) {
		return nil, HandleErrorResp(resp, r.ErrorHandler, r.IsOpenAI)
	}

	// 解析响应
	if response == nil {
		return resp, nil
	}

	if outputResp {
		var buf bytes.Buffer
		tee := io.TeeReader(resp.Body, &buf)
		err = DecodeResponse(tee, response)

		// 将响应体重新写入 resp.Body
		resp.Body = io.NopCloser(&buf)
	} else {
		err = json.NewDecoder(resp.Body).Decode(response)
	}

	if err != nil {
		return nil, common.ErrorWrapper(err, "decode_response_failed", http.StatusInternalServerError)
	}

	return resp, nil
}

// 发送请求 RAW
func (r *HTTPRequester) SendRequestRaw(req *http.Request) (*http.Response, *types.OpenAIErrorWithStatusCode) {
	// 发送请求
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, common.ErrorWrapper(err, "http_request_failed", http.StatusInternalServerError)
	}

	// 处理响应
	if r.IsFailureStatusCode(resp) {
		return nil, HandleErrorResp(resp, r.ErrorHandler, r.IsOpenAI)
	}

	return resp, nil
}

// 获取流式响应
func RequestStream[T streamable](requester *HTTPRequester, resp *http.Response, handlerPrefix HandlerPrefix[T]) (*streamReader[T], *types.OpenAIErrorWithStatusCode) {
	// 如果返回的头是json格式 说明有错误
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		return nil, HandleErrorResp(resp, requester.ErrorHandler, requester.IsOpenAI)
	}

	stream := &streamReader[T]{
		reader:        bufio.NewReader(resp.Body),
		response:      resp,
		handlerPrefix: handlerPrefix,

		DataChan: make(chan T),
		ErrChan:  make(chan error),
	}

	return stream, nil
}

// 设置请求体
func (r *HTTPRequester) WithBody(body any) requestOption {
	return func(args *requestOptions) {
		args.body = body
	}
}

// 设置请求头
func (r *HTTPRequester) WithHeader(header map[string]string) requestOption {
	return func(args *requestOptions) {
		for k, v := range header {
			args.header.Set(k, v)
		}
	}
}

// 设置Content-Type
func (r *HTTPRequester) WithContentType(contentType string) requestOption {
	return func(args *requestOptions) {
		args.header.Set("Content-Type", contentType)
	}
}

// 判断是否为失败状态码
func (r *HTTPRequester) IsFailureStatusCode(resp *http.Response) bool {
	return resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest
}

// 处理错误响应
func HandleErrorResp(resp *http.Response, toOpenAIError HttpErrorHandler, isPrefix bool) *types.OpenAIErrorWithStatusCode {

	openAIErrorWithStatusCode := &types.OpenAIErrorWithStatusCode{
		StatusCode: resp.StatusCode,
		OpenAIError: types.OpenAIError{
			Message: "",
			Type:    "upstream_error",
			Code:    "bad_response_status_code",
			Param:   strconv.Itoa(resp.StatusCode),
		},
	}

	defer resp.Body.Close()

	if toOpenAIError != nil {
		errorResponse := toOpenAIError(resp)

		if errorResponse != nil && errorResponse.Message != "" {
			openAIErrorWithStatusCode.OpenAIError = *errorResponse
			if isPrefix {
				openAIErrorWithStatusCode.OpenAIError.Message = fmt.Sprintf("Provider API error: %s", openAIErrorWithStatusCode.OpenAIError.Message)
			}
		}
	}

	if openAIErrorWithStatusCode.OpenAIError.Message == "" {
		openAIErrorWithStatusCode.OpenAIError.Message = fmt.Sprintf("Provider API error: bad response status code %d", resp.StatusCode)
		if isPrefix {
			openAIErrorWithStatusCode.OpenAIError.Message = fmt.Sprintf("Provider API error: bad response status code %d", resp.StatusCode)
		} else {
			openAIErrorWithStatusCode.OpenAIError.Message = fmt.Sprintf("bad response status code %d", resp.StatusCode)
		}
	}

	return openAIErrorWithStatusCode
}

func SetEventStreamHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
}

func GetJsonHeaders() map[string]string {
	return map[string]string{
		"Content-type": "application/json",
	}
}

type Stringer interface {
	GetString() *string
}

func DecodeResponse(body io.Reader, v any) error {
	if v == nil {
		return nil
	}

	if result, ok := v.(*string); ok {
		return DecodeString(body, result)
	}

	if stringer, ok := v.(Stringer); ok {
		return DecodeString(body, stringer.GetString())
	}

	return json.NewDecoder(body).Decode(v)
}

func DecodeString(body io.Reader, output *string) error {
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	*output = string(b)
	return nil
}
