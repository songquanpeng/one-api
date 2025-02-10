package gemini

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/songquanpeng/one-api/common/render"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/image"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/model"

	"github.com/gin-gonic/gin"
)

// https://ai.google.dev/docs/gemini_api_overview?hl=zh-cn

const (
	VisionMaxImageNum = 16
)

var mimeTypeMap = map[string]string{
	"json_object": "application/json",
	"text":        "text/plain",
}

var toolChoiceTypeMap = map[string]string{
	"none":     "NONE",
	"auto":     "AUTO",
	"required": "ANY",
}

// Setting safety to the lowest possible values since Gemini is already powerless enough
func ConvertRequest(textRequest model.GeneralOpenAIRequest) *ChatRequest {
	geminiRequest := ChatRequest{
		Contents: make([]ChatContent, 0, len(textRequest.Messages)),
		SafetySettings: []ChatSafetySettings{
			{
				Category:  "HARM_CATEGORY_HARASSMENT",
				Threshold: config.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_HATE_SPEECH",
				Threshold: config.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
				Threshold: config.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
				Threshold: config.GeminiSafetySetting,
			},
			{
				Category:  "HARM_CATEGORY_CIVIC_INTEGRITY",
				Threshold: config.GeminiSafetySetting,
			},
		},
		GenerationConfig: ChatGenerationConfig{
			Temperature:     textRequest.Temperature,
			TopP:            textRequest.TopP,
			MaxOutputTokens: textRequest.MaxTokens,
		},
	}
	if textRequest.ResponseFormat != nil {
		if mimeType, ok := mimeTypeMap[textRequest.ResponseFormat.Type]; ok {
			geminiRequest.GenerationConfig.ResponseMimeType = mimeType
		}
		if textRequest.ResponseFormat.JsonSchema != nil {
			geminiRequest.GenerationConfig.ResponseSchema = textRequest.ResponseFormat.JsonSchema.Schema
			geminiRequest.GenerationConfig.ResponseMimeType = mimeTypeMap["json_object"]
		}
	}
	if textRequest.Tools != nil {
		functions := make([]model.Function, 0, len(textRequest.Tools))
		for _, tool := range textRequest.Tools {
			functions = append(functions, tool.Function)
		}
		geminiRequest.Tools = []ChatTools{
			{
				FunctionDeclarations: functions,
			},
		}
	} else if textRequest.Functions != nil {
		geminiRequest.Tools = []ChatTools{
			{
				FunctionDeclarations: textRequest.Functions,
			},
		}
	}
	if textRequest.ToolChoice != nil {
		geminiRequest.ToolConfig = &ToolConfig{
			FunctionCallingConfig: FunctionCallingConfig{
				Mode: "auto",
			},
		}
		switch mode := textRequest.ToolChoice.(type) {
		case string:
			geminiRequest.ToolConfig.FunctionCallingConfig.Mode = toolChoiceTypeMap[mode]
		case map[string]interface{}:
			geminiRequest.ToolConfig.FunctionCallingConfig.Mode = "ANY"
			if fn, ok := mode["function"].(map[string]interface{}); ok {
				if name, ok := fn["name"].(string); ok {
					geminiRequest.ToolConfig.FunctionCallingConfig.AllowedFunctionNames = []string{name}
				}
			}
		}
	}
	for _, message := range textRequest.Messages {
		content := ChatContent{
			Role: message.Role,
			Parts: []Part{
				{
					Text: message.StringContent(),
				},
			},
		}
		openaiContent := message.ParseContent()
		var parts []Part
		imageNum := 0
		for _, part := range openaiContent {
			if part.Type == model.ContentTypeText {
				parts = append(parts, Part{
					Text: part.Text,
				})
			} else if part.Type == model.ContentTypeImageURL {
				imageNum += 1
				if imageNum > VisionMaxImageNum {
					continue
				}
				mimeType, data, _ := image.GetImageFromUrl(part.ImageURL.Url)
				parts = append(parts, Part{
					InlineData: &InlineData{
						MimeType: mimeType,
						Data:     data,
					},
				})
			}
		}
		content.Parts = parts

		// there's no assistant role in gemini and API shall vomit if Role is not user or model
		if content.Role == "assistant" {
			content.Role = "model"
		}
		// Converting system prompt to SystemInstructions
		if content.Role == "system" {
			geminiRequest.SystemInstruction = &content
			continue
		}
		geminiRequest.Contents = append(geminiRequest.Contents, content)
	}

	return &geminiRequest
}

func ConvertEmbeddingRequest(request model.GeneralOpenAIRequest) *BatchEmbeddingRequest {
	inputs := request.ParseInput()
	requests := make([]EmbeddingRequest, len(inputs))
	model := fmt.Sprintf("models/%s", request.Model)

	for i, input := range inputs {
		requests[i] = EmbeddingRequest{
			Model: model,
			Content: ChatContent{
				Parts: []Part{
					{
						Text: input,
					},
				},
			},
		}
	}

	return &BatchEmbeddingRequest{
		Requests: requests,
	}
}

type ChatResponse struct {
	Candidates     []ChatCandidate    `json:"candidates"`
	PromptFeedback ChatPromptFeedback `json:"promptFeedback"`
}

func (g *ChatResponse) GetResponseText() string {
	if g == nil {
		return ""
	}
	var builder strings.Builder
	for _, candidate := range g.Candidates {
		for idx, part := range candidate.Content.Parts {
			if idx > 0 {
				builder.WriteString("\n")
			}
			builder.WriteString(part.Text)
		}
	}
	return builder.String()
}

type ChatCandidate struct {
	Content       ChatContent        `json:"content"`
	FinishReason  string             `json:"finishReason"`
	Index         int64              `json:"index"`
	SafetyRatings []ChatSafetyRating `json:"safetyRatings"`
}

type ChatSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

type ChatPromptFeedback struct {
	SafetyRatings []ChatSafetyRating `json:"safetyRatings"`
}

func getToolCalls(candidate *ChatCandidate) []model.Tool {
	var toolCalls []model.Tool

	item := candidate.Content.Parts[0]
	if item.FunctionCall == nil {
		return toolCalls
	}
	argsBytes, err := json.Marshal(item.FunctionCall.Arguments)
	if err != nil {
		logger.FatalLog("getToolCalls failed: " + err.Error())
		return toolCalls
	}
	toolCall := model.Tool{
		Id:   fmt.Sprintf("call_%s", random.GetUUID()),
		Type: "function",
		Function: model.Function{
			Arguments: string(argsBytes),
			Name:      item.FunctionCall.FunctionName,
		},
	}
	toolCalls = append(toolCalls, toolCall)
	return toolCalls
}

func responseGeminiChat2OpenAI(response *ChatResponse) *openai.TextResponse {
	fullTextResponse := openai.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", random.GetUUID()),
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: make([]openai.TextResponseChoice, 0, len(response.Candidates)),
	}
	for i, candidate := range response.Candidates {
		choice := openai.TextResponseChoice{
			Index: i,
			Message: model.Message{
				Role: "assistant",
			},
			FinishReason: constant.StopFinishReason,
		}
		if len(candidate.Content.Parts) > 0 {
			if candidate.Content.Parts[0].FunctionCall != nil {
				choice.Message.ToolCalls = getToolCalls(&candidate)
			} else {
				var builder strings.Builder
				for idx, part := range candidate.Content.Parts {
					if idx > 0 {
						builder.WriteString("\n")
					}
					builder.WriteString(part.Text)
				}
				choice.Message.Content = builder.String()
			}
		} else {
			choice.Message.Content = ""
			choice.FinishReason = candidate.FinishReason
		}
		fullTextResponse.Choices = append(fullTextResponse.Choices, choice)
	}
	return &fullTextResponse
}

func streamResponseGeminiChat2OpenAI(geminiResponse *ChatResponse) *openai.ChatCompletionsStreamResponse {
	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta.Content = geminiResponse.GetResponseText()
	//choice.FinishReason = &constant.StopFinishReason
	var response openai.ChatCompletionsStreamResponse
	response.Id = fmt.Sprintf("chatcmpl-%s", random.GetUUID())
	response.Created = helper.GetTimestamp()
	response.Object = "chat.completion.chunk"
	response.Model = "gemini"
	response.Choices = []openai.ChatCompletionsStreamResponseChoice{choice}
	return &response
}

func embeddingResponseGemini2OpenAI(response *EmbeddingResponse) *openai.EmbeddingResponse {
	openAIEmbeddingResponse := openai.EmbeddingResponse{
		Object: "list",
		Data:   make([]openai.EmbeddingResponseItem, 0, len(response.Embeddings)),
		Model:  "gemini-embedding",
		Usage:  model.Usage{TotalTokens: 0},
	}
	for _, item := range response.Embeddings {
		openAIEmbeddingResponse.Data = append(openAIEmbeddingResponse.Data, openai.EmbeddingResponseItem{
			Object:    `embedding`,
			Index:     0,
			Embedding: item.Values,
		})
	}
	return &openAIEmbeddingResponse
}

func StreamHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, string) {
	responseText := ""
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	common.SetEventStreamHeaders(c)

	for scanner.Scan() {
		data := scanner.Text()
		data = strings.TrimSpace(data)
		if !strings.HasPrefix(data, "data: ") {
			continue
		}
		data = strings.TrimPrefix(data, "data: ")
		data = strings.TrimSuffix(data, "\"")

		var geminiResponse ChatResponse
		err := json.Unmarshal([]byte(data), &geminiResponse)
		if err != nil {
			logger.SysError("error unmarshalling stream response: " + err.Error())
			continue
		}

		response := streamResponseGeminiChat2OpenAI(&geminiResponse)
		if response == nil {
			continue
		}

		responseText += response.Choices[0].Delta.StringContent()

		err = render.ObjectData(c, response)
		if err != nil {
			logger.SysError(err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		logger.SysError("error reading stream: " + err.Error())
	}

	render.Done(c)

	err := resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), ""
	}

	return nil, responseText
}

func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var geminiResponse ChatResponse
	err = json.Unmarshal(responseBody, &geminiResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if len(geminiResponse.Candidates) == 0 {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: "No candidates returned",
				Type:    "server_error",
				Param:   "",
				Code:    500,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := responseGeminiChat2OpenAI(&geminiResponse)
	fullTextResponse.Model = modelName
	completionTokens := openai.CountTokenText(geminiResponse.GetResponseText(), modelName)
	usage := model.Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}
	fullTextResponse.Usage = usage
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &usage
}

func EmbeddingHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	var geminiEmbeddingResponse EmbeddingResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &geminiEmbeddingResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if geminiEmbeddingResponse.Error != nil {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: geminiEmbeddingResponse.Error.Message,
				Type:    "gemini_error",
				Param:   "",
				Code:    geminiEmbeddingResponse.Error.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}
	fullTextResponse := embeddingResponseGemini2OpenAI(&geminiEmbeddingResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}
