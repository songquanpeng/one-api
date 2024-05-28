package gemini

import (
	"encoding/json"
	"net/http"
	"one-api/common"
	"one-api/common/image"
	"one-api/common/utils"
	"one-api/types"
)

type GeminiChatRequest struct {
	Contents         []GeminiChatContent        `json:"contents"`
	SafetySettings   []GeminiChatSafetySettings `json:"safety_settings,omitempty"`
	GenerationConfig GeminiChatGenerationConfig `json:"generation_config,omitempty"`
	Tools            []GeminiChatTools          `json:"tools,omitempty"`
}

type GeminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type GeminiPart struct {
	FunctionCall     *GeminiFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *GeminiFunctionResponse `json:"functionResponse,omitempty"`
	Text             string                  `json:"text,omitempty"`
	InlineData       *GeminiInlineData       `json:"inlineData,omitempty"`
}

type GeminiFunctionCall struct {
	Name string                 `json:"name,omitempty"`
	Args map[string]interface{} `json:"args,omitempty"`
}

func (candidate *GeminiChatCandidate) ToOpenAIStreamChoice(request *types.ChatCompletionRequest) types.ChatCompletionStreamChoice {
	choice := types.ChatCompletionStreamChoice{
		Index: int(candidate.Index),
		Delta: types.ChatCompletionStreamChoiceDelta{
			Role: types.ChatMessageRoleAssistant,
		},
		FinishReason: types.FinishReasonStop,
	}

	content := ""
	isTools := false

	for _, part := range candidate.Content.Parts {
		if part.FunctionCall != nil {
			if choice.Delta.ToolCalls == nil {
				choice.Delta.ToolCalls = make([]*types.ChatCompletionToolCalls, 0)
			}
			isTools = true
			choice.Delta.ToolCalls = append(choice.Delta.ToolCalls, part.FunctionCall.ToOpenAITool())
		} else {
			content += part.Text
		}
	}

	choice.Delta.Content = content

	if isTools {
		choice.FinishReason = types.FinishReasonToolCalls
	}
	choice.CheckChoice(request)

	return choice
}

func (candidate *GeminiChatCandidate) ToOpenAIChoice(request *types.ChatCompletionRequest) types.ChatCompletionChoice {
	choice := types.ChatCompletionChoice{
		Index: int(candidate.Index),
		Message: types.ChatCompletionMessage{
			Role: "assistant",
		},
		FinishReason: types.FinishReasonStop,
	}

	if len(candidate.Content.Parts) == 0 {
		choice.Message.Content = ""
		return choice
	}

	content := ""
	useTools := false

	for _, part := range candidate.Content.Parts {
		if part.FunctionCall != nil {
			if choice.Message.ToolCalls == nil {
				choice.Message.ToolCalls = make([]*types.ChatCompletionToolCalls, 0)
			}
			useTools = true
			choice.Message.ToolCalls = append(choice.Message.ToolCalls, part.FunctionCall.ToOpenAITool())
		} else {
			content += part.Text
		}
	}

	choice.Message.Content = content

	if useTools {
		choice.FinishReason = types.FinishReasonToolCalls
	}

	choice.CheckChoice(request)

	return choice
}

type GeminiFunctionResponse struct {
	Name     string                        `json:"name,omitempty"`
	Response GeminiFunctionResponseContent `json:"response,omitempty"`
}

type GeminiFunctionResponseContent struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
}

func (g *GeminiFunctionCall) ToOpenAITool() *types.ChatCompletionToolCalls {
	args, _ := json.Marshal(g.Args)

	return &types.ChatCompletionToolCalls{
		Id:    "call_" + utils.GetRandomString(24),
		Type:  types.ChatMessageRoleFunction,
		Index: 0,
		Function: &types.ChatCompletionToolCallsFunction{
			Name:      g.Name,
			Arguments: string(args),
		},
	}
}

type GeminiChatContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiChatSafetySettings struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type GeminiChatTools struct {
	FunctionDeclarations []types.ChatCompletionFunction `json:"functionDeclarations,omitempty"`
}

type GeminiChatGenerationConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            float64  `json:"topK,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	CandidateCount  int      `json:"candidateCount,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type GeminiErrorResponse struct {
	Error GeminiError `json:"error,omitempty"`
}

type GeminiChatResponse struct {
	Candidates     []GeminiChatCandidate    `json:"candidates"`
	PromptFeedback GeminiChatPromptFeedback `json:"promptFeedback"`
	UsageMetadata  *GeminiUsageMetadata     `json:"usageMetadata,omitempty"`
	Model          string                   `json:"model,omitempty"`
	GeminiErrorResponse
}

type GeminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

type GeminiChatCandidate struct {
	Content       GeminiChatContent        `json:"content"`
	FinishReason  string                   `json:"finishReason"`
	Index         int64                    `json:"index"`
	SafetyRatings []GeminiChatSafetyRating `json:"safetyRatings"`
}

type GeminiChatSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

type GeminiChatPromptFeedback struct {
	SafetyRatings []GeminiChatSafetyRating `json:"safetyRatings"`
}

func (g *GeminiChatResponse) GetResponseText() string {
	if g == nil {
		return ""
	}
	if len(g.Candidates) > 0 && len(g.Candidates[0].Content.Parts) > 0 {
		return g.Candidates[0].Content.Parts[0].Text
	}
	return ""
}

func OpenAIToGeminiChatContent(openaiContents []types.ChatCompletionMessage) ([]GeminiChatContent, *types.OpenAIErrorWithStatusCode) {
	contents := make([]GeminiChatContent, 0)
	useToolName := ""
	for _, openaiContent := range openaiContents {
		content := GeminiChatContent{
			Role:  ConvertRole(openaiContent.Role),
			Parts: make([]GeminiPart, 0),
		}
		content.Role = ConvertRole(openaiContent.Role)
		if openaiContent.ToolCalls != nil || openaiContent.FunctionCall != nil {
			if openaiContent.ToolCalls != nil {
				useToolName = openaiContent.ToolCalls[0].Function.Name
			} else {
				useToolName = openaiContent.FunctionCall.Name
			}
			content = GeminiChatContent{
				Role: "model",
				Parts: []GeminiPart{
					{
						FunctionCall: &GeminiFunctionCall{
							Name: useToolName,
							Args: map[string]interface{}{},
						},
					},
				},
			}
		} else if openaiContent.Role == types.ChatMessageRoleFunction || openaiContent.Role == types.ChatMessageRoleTool {
			if openaiContent.Name == nil {
				openaiContent.Name = &useToolName
			}
			content = GeminiChatContent{
				Role: "function",
				Parts: []GeminiPart{
					{
						FunctionResponse: &GeminiFunctionResponse{
							Name: *openaiContent.Name,
							Response: GeminiFunctionResponseContent{
								Name:    *openaiContent.Name,
								Content: openaiContent.StringContent(),
							},
						},
					},
				},
			}
		} else {
			openaiMessagePart := openaiContent.ParseContent()
			imageNum := 0
			for _, openaiPart := range openaiMessagePart {
				if openaiPart.Type == types.ContentTypeText {
					content.Parts = append(content.Parts, GeminiPart{
						Text: openaiPart.Text,
					})
				} else if openaiPart.Type == types.ContentTypeImageURL {
					imageNum += 1
					if imageNum > GeminiVisionMaxImageNum {
						continue
					}
					mimeType, data, err := image.GetImageFromUrl(openaiPart.ImageURL.URL)
					if err != nil {
						return nil, common.ErrorWrapper(err, "image_url_invalid", http.StatusBadRequest)
					}
					content.Parts = append(content.Parts, GeminiPart{
						InlineData: &GeminiInlineData{
							MimeType: mimeType,
							Data:     data,
						},
					})
				}
			}
		}
		contents = append(contents, content)
		if openaiContent.Role == types.ChatMessageRoleSystem {
			contents = append(contents, GeminiChatContent{
				Role: "model",
				Parts: []GeminiPart{
					{
						Text: "Okay",
					},
				},
			})
		}

	}

	return contents, nil
}

func ConvertRole(roleName string) string {
	switch roleName {
	case types.ChatMessageRoleFunction, types.ChatMessageRoleTool:
		return types.ChatMessageRoleFunction
	case types.ChatMessageRoleAssistant:
		return "model"
	default:
		return types.ChatMessageRoleUser
	}
}

type ModelListResponse struct {
	Models []ModelDetails `json:"models"`
}

type ModelDetails struct {
	Name                       string   `json:"name"`
	SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
}
