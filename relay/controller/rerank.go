package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
	"io"
	"net/http"
)

func RerankHelper(c *gin.Context, relayMode int) *relaymodel.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := meta.GetByContext(c)
	rerankRequest, err := getRerankRequest(c, meta.Mode)
	if err != nil {
		logger.Errorf(ctx, "getRerankRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "invalid_rerank_request", http.StatusBadRequest)
	}

	// Map model name
	var isModelMapped bool
	meta.OriginModelName = rerankRequest.Model
	rerankRequest.Model, isModelMapped = getMappedModelName(rerankRequest.Model, meta.ModelMapping)
	meta.ActualModelName = rerankRequest.Model

	var requestBody io.Reader
	if isModelMapped {
		jsonStr, err := json.Marshal(rerankRequest)
		if err != nil {
			return openai.ErrorWrapper(err, "marshal_rerank_request_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonStr)
	} else {
		requestBody = c.Request.Body
	}

	adaptor := relay.GetAdaptor(meta.APIType)
	if adaptor == nil {
		return openai.ErrorWrapper(fmt.Errorf("invalid api type: %d", meta.APIType), "invalid_api_type", http.StatusBadRequest)
	}

	resp, err := adaptor.DoRequest(c, meta, requestBody)
	if err != nil {
		logger.Errorf(ctx, "DoRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}

	// do response
	_, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		logger.Errorf(ctx, "respErr is not nil: %+v", respErr)
		return respErr
	}

	return nil
}
