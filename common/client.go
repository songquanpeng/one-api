package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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
	createFormBuilder func(io.Writer) FormBuilder
}

func NewClient() *Client {
	return &Client{
		requestBuilder: NewRequestBuilder(),
		createFormBuilder: func(body io.Writer) FormBuilder {
			return NewFormBuilder(body)
		},
	}
}

type requestOptions struct {
	body   any
	header http.Header
}

type requestOption func(*requestOptions)

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

func (c *Client) SendRequest(req *http.Request, response any) error {

	// 发送请求
	resp, err := HttpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// 处理响应
	if IsFailureStatusCode(resp) {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	// 解析响应
	err = DecodeResponse(resp.Body, response)
	if err != nil {
		return err
	}

	return nil
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
