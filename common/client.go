package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/types"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var HttpClient *http.Client

func init() {
	if RelayTimeout == 0 {
		HttpClient = &http.Client{}
	} else {
		HttpClient = &http.Client{
			Timeout: time.Duration(RelayTimeout) * time.Second,
		}
	}
}

type Client struct {
	requestBuilder    RequestBuilder
	CreateFormBuilder func(io.Writer) FormBuilder
}

func NewClient() *Client {
	return &Client{
		requestBuilder: NewRequestBuilder(),
		CreateFormBuilder: func(body io.Writer) FormBuilder {
			return NewFormBuilder(body)
		},
	}
}

type requestOptions struct {
	body   any
	header http.Header
}

type requestOption func(*requestOptions)

type Stringer interface {
	GetString() *string
}

func WithBody(body any) requestOption {
	return func(args *requestOptions) {
		args.body = body
	}
}

func WithHeader(header map[string]string) requestOption {
	return func(args *requestOptions) {
		for k, v := range header {
			args.header.Set(k, v)
		}
	}
}

func WithContentType(contentType string) requestOption {
	return func(args *requestOptions) {
		args.header.Set("Content-Type", contentType)
	}
}

type RequestError struct {
	HTTPStatusCode int
	Err            error
}

func (c *Client) NewRequest(method, url string, setters ...requestOption) (*http.Request, error) {
	// Default Options
	args := &requestOptions{
		body:   nil,
		header: make(http.Header),
	}
	for _, setter := range setters {
		setter(args)
	}
	req, err := c.requestBuilder.Build(method, url, args.body, args.header)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func SendRequest(req *http.Request, response any, outputResp bool) (*http.Response, *types.OpenAIErrorWithStatusCode) {
	// 发送请求
	resp, err := HttpClient.Do(req)
	if err != nil {
		return nil, ErrorWrapper(err, "http_request_failed", http.StatusInternalServerError)
	}

	if !outputResp {
		defer resp.Body.Close()
	}

	// 处理响应
	if IsFailureStatusCode(resp) {
		return nil, HandleErrorResp(resp)
	}

	// 解析响应
	if outputResp {
		var buf bytes.Buffer
		tee := io.TeeReader(resp.Body, &buf)
		err = DecodeResponse(tee, response)

		// 将响应体重新写入 resp.Body
		resp.Body = io.NopCloser(&buf)
	} else {
		err = DecodeResponse(resp.Body, response)
	}
	if err != nil {
		return nil, ErrorWrapper(err, "decode_response_failed", http.StatusInternalServerError)
	}

	if outputResp {
		return resp, nil
	}

	return nil, nil
}

// 处理错误响应
func HandleErrorResp(resp *http.Response) (openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
	openAIErrorWithStatusCode = &types.OpenAIErrorWithStatusCode{
		StatusCode: resp.StatusCode,
		OpenAIError: types.OpenAIError{
			Message: fmt.Sprintf("bad response status code %d", resp.StatusCode),
			Type:    "upstream_error",
			Code:    "bad_response_status_code",
			Param:   strconv.Itoa(resp.StatusCode),
		},
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = resp.Body.Close()
	if err != nil {
		return
	}
	var errorResponse types.OpenAIErrorResponse
	err = json.Unmarshal(responseBody, &errorResponse)
	if err != nil {
		return
	}
	if errorResponse.Error.Type != "" {
		openAIErrorWithStatusCode.OpenAIError = errorResponse.Error
	} else {
		openAIErrorWithStatusCode.OpenAIError.Message = string(responseBody)
	}
	return
}

func (c *Client) SendRequestRaw(req *http.Request) (body io.ReadCloser, err error) {
	resp, err := HttpClient.Do(req)
	if err != nil {
		return
	}

	return resp.Body, nil
}

func IsFailureStatusCode(resp *http.Response) bool {
	return resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest
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

func SetEventStreamHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
}
