package common

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/songquanpeng/one-api/common/ctxkey"
)

func GetRequestBody(c *gin.Context) (requestBody []byte, err error) {
	if requestBodyCache, _ := c.Get(ctxkey.KeyRequestBody); requestBodyCache != nil {
		return requestBodyCache.([]byte), nil
	}
	requestBody, err = io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	_ = c.Request.Body.Close()
	c.Set(ctxkey.KeyRequestBody, requestBody)

	return requestBody, nil
}

func UnmarshalBodyReusable(c *gin.Context, v any) error {
	requestBody, err := GetRequestBody(c)
	if err != nil {
		return err
	}

	// check v should be a pointer
	if v == nil || reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.Errorf("UnmarshalBodyReusable only accept pointer, got %v", reflect.TypeOf(v))
	}

	contentType := c.Request.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		err = json.Unmarshal(requestBody, v)
	} else {
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		err = c.ShouldBind(v)
	}
	if err != nil {
		return errors.Wrap(err, "unmarshal request body failed")
	}
	// Reset request body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

	return nil
}

func SetEventStreamHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
}
