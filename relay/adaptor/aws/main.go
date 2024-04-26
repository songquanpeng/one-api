// Package aws provides the AWS adaptor for the relay service.
package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/adaptor/anthropic"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

func wrapErr(err error) *relaymodel.ErrorWithStatusCode {
	return &relaymodel.ErrorWithStatusCode{
		StatusCode: http.StatusInternalServerError,
		Error: relaymodel.Error{
			Message: fmt.Sprintf("%s", err.Error()),
		},
	}
}

// https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html
var awsModelIDMap = map[string]string{
	"claude-instant-1.2":       "anthropic.claude-instant-v1",
	"claude-2.0":               "anthropic.claude-v2",
	"claude-2.1":               "anthropic.claude-v2:1",
	"claude-3-sonnet-20240229": "anthropic.claude-3-sonnet-20240229-v1:0",
	"claude-3-opus-20240229":   "anthropic.claude-3-opus-20240229-v1:0",
	"claude-3-haiku-20240307":  "anthropic.claude-3-haiku-20240307-v1:0",
}

func awsModelID(requestModel string) (string, error) {
	if awsModelID, ok := awsModelIDMap[requestModel]; ok {
		return awsModelID, nil
	}

	return "", errors.Errorf("model %s not found", requestModel)
}

func Handler(c *gin.Context, awsCli *bedrockruntime.Client, modelName string) (*relaymodel.ErrorWithStatusCode, *relaymodel.Usage) {
	awsModelId, err := awsModelID(c.GetString(ctxkey.RequestModel))
	if err != nil {
		return wrapErr(errors.Wrap(err, "awsModelID")), nil
	}

	awsReq := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(awsModelId),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	}

	claudeReq_, ok := c.Get(ctxkey.ConvertedRequest)
	if !ok {
		return wrapErr(errors.New("request not found")), nil
	}
	claudeReq := claudeReq_.(*anthropic.Request)
	awsClaudeReq := &Request{
		AnthropicVersion: "bedrock-2023-05-31",
	}
	if err = copier.Copy(awsClaudeReq, claudeReq); err != nil {
		return wrapErr(errors.Wrap(err, "copy request")), nil
	}

	awsReq.Body, err = json.Marshal(awsClaudeReq)
	if err != nil {
		return wrapErr(errors.Wrap(err, "marshal request")), nil
	}

	awsResp, err := awsCli.InvokeModel(c.Request.Context(), awsReq)
	if err != nil {
		return wrapErr(errors.Wrap(err, "InvokeModel")), nil
	}

	claudeResponse := new(anthropic.Response)
	err = json.Unmarshal(awsResp.Body, claudeResponse)
	if err != nil {
		return wrapErr(errors.Wrap(err, "unmarshal response")), nil
	}

	openaiResp := anthropic.ResponseClaude2OpenAI(claudeResponse)
	openaiResp.Model = modelName
	usage := relaymodel.Usage{
		PromptTokens:     claudeResponse.Usage.InputTokens,
		CompletionTokens: claudeResponse.Usage.OutputTokens,
		TotalTokens:      claudeResponse.Usage.InputTokens + claudeResponse.Usage.OutputTokens,
	}
	openaiResp.Usage = usage

	c.JSON(http.StatusOK, openaiResp)
	return nil, &usage
}

func StreamHandler(c *gin.Context, awsCli *bedrockruntime.Client) (*relaymodel.ErrorWithStatusCode, *relaymodel.Usage) {
	createdTime := helper.GetTimestamp()
	awsModelId, err := awsModelID(c.GetString(ctxkey.RequestModel))
	if err != nil {
		return wrapErr(errors.Wrap(err, "awsModelID")), nil
	}

	awsReq := &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(awsModelId),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	}

	claudeReq_, ok := c.Get(ctxkey.ConvertedRequest)
	if !ok {
		return wrapErr(errors.New("request not found")), nil
	}
	claudeReq := claudeReq_.(*anthropic.Request)

	awsClaudeReq := &Request{
		AnthropicVersion: "bedrock-2023-05-31",
	}
	if err = copier.Copy(awsClaudeReq, claudeReq); err != nil {
		return wrapErr(errors.Wrap(err, "copy request")), nil
	}
	awsReq.Body, err = json.Marshal(awsClaudeReq)
	if err != nil {
		return wrapErr(errors.Wrap(err, "marshal request")), nil
	}

	awsResp, err := awsCli.InvokeModelWithResponseStream(c.Request.Context(), awsReq)
	if err != nil {
		return wrapErr(errors.Wrap(err, "InvokeModelWithResponseStream")), nil
	}
	stream := awsResp.GetStream()
	defer stream.Close()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	var usage relaymodel.Usage
	var id string
	c.Stream(func(w io.Writer) bool {
		event, ok := <-stream.Events()
		if !ok {
			c.Render(-1, common.CustomEvent{Data: "data: [DONE]"})
			return false
		}

		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			claudeResp := new(anthropic.StreamResponse)
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(claudeResp)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				return false
			}

			response, meta := anthropic.StreamResponseClaude2OpenAI(claudeResp)
			if meta != nil {
				usage.PromptTokens += meta.Usage.InputTokens
				usage.CompletionTokens += meta.Usage.OutputTokens
				id = fmt.Sprintf("chatcmpl-%s", meta.Id)
				return true
			}
			if response == nil {
				return true
			}
			response.Id = id
			response.Model = c.GetString(ctxkey.OriginalModel)
			response.Created = createdTime
			jsonStr, err := json.Marshal(response)
			if err != nil {
				logger.SysError("error marshalling stream response: " + err.Error())
				return true
			}
			c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonStr)})
			return true
		case *types.UnknownUnionMember:
			fmt.Println("unknown tag:", v.Tag)
			return false
		default:
			fmt.Println("union is nil or unknown type")
			return false
		}
	})

	return nil, &usage
}
