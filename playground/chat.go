package playground

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/songquanpeng/one-api/model"
	openaiadaptor "github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/apitype"
	"github.com/songquanpeng/one-api/relay/channeltype"
)

// SearchOptions 聊天请求中的联网搜索选项
type SearchOptions struct {
	Enabled bool   `json:"enabled"`
	Query   string `json:"query"` // 可选，缺省取最后一条用户消息
}

// ChatRequest 游乐场聊天请求
type ChatRequest struct {
	SessionID string        `json:"session_id"`
	Model     string        `json:"model"`
	Group     string        `json:"group"`
	Messages  []Message     `json:"messages"`
	Images    []string      `json:"images"` // data URL / http(s) URL / 已上传文件名
	Search    SearchOptions `json:"search"`
	Stream    bool          `json:"stream"`
}

// ChatResult 非流式聊天结果
type ChatResult struct {
	Reply         string         `json:"reply"`
	SearchResults []SearchResult `json:"search_results,omitempty"`
	Usage         map[string]any `json:"usage,omitempty"`
	ChannelID     int            `json:"channel_id"`
	ChannelName   string         `json:"channel_name"`
	ActualModel   string         `json:"actual_model"`
}

// NormalizeMessages 把客户端消息规整为内部格式（content 支持字符串或片段数组）
func NormalizeMessages(raw []Message) []Message {
	out := make([]Message, 0, len(raw))
	for _, m := range raw {
		msg := Message{Role: m.Role}
		if len(m.Content) == 0 {
			continue
		}
		for _, part := range m.Content {
			switch part.Type {
			case "text":
				if strings.TrimSpace(part.Text) != "" {
					msg.Content = append(msg.Content, ContentPart{Type: "text", Text: part.Text})
				}
			case "image_url":
				if part.ImageURL != nil && strings.TrimSpace(part.ImageURL.URL) != "" {
					msg.Content = append(msg.Content, ContentPart{Type: "image_url", ImageURL: &ImageURL{URL: part.ImageURL.URL}})
				}
			}
		}
		if len(msg.Content) > 0 {
			out = append(out, msg)
		}
	}
	return out
}

// ResolveImages 把图片引用统一解析为可直接使用的 URL：
// data URL 与 http(s) URL 原样返回，其余视为已上传文件名并转为 base64 data URL
func ResolveImages(images []string) ([]string, error) {
	out := make([]string, 0, len(images))
	for _, img := range images {
		img = strings.TrimSpace(img)
		if img == "" {
			continue
		}
		if strings.HasPrefix(img, "data:") || strings.HasPrefix(img, "http://") || strings.HasPrefix(img, "https://") {
			out = append(out, img)
			continue
		}
		dataURL, err := ReadUploadFileAsDataURL(img)
		if err != nil {
			return nil, fmt.Errorf("图片 %s 读取失败：%w", img, err)
		}
		out = append(out, dataURL)
	}
	return out, nil
}

// AppendImagesToLastUserMessage 把图片附加到最后一条 user 消息上
func AppendImagesToLastUserMessage(messages []Message, images []string) []Message {
	if len(images) == 0 {
		return messages
	}
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			for _, img := range images {
				messages[i].Content = append(messages[i].Content, ContentPart{
					Type:     "image_url",
					ImageURL: &ImageURL{URL: img},
				})
			}
			return messages
		}
	}
	return messages
}

// ExtractLastUserText 提取最后一条用户消息的文本（用作默认搜索词）
func ExtractLastUserText(messages []Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role != "user" {
			continue
		}
		var sb strings.Builder
		for _, part := range messages[i].Content {
			if part.Type == "text" {
				sb.WriteString(part.Text)
			}
		}
		if text := strings.TrimSpace(sb.String()); text != "" {
			return text
		}
	}
	return ""
}

// SelectPlaygroundChannel 按模型与分组挑选一个可用的 OpenAI 兼容渠道
func SelectPlaygroundChannel(group, modelName string) (*model.Channel, error) {
	channel, err := model.CacheGetRandomSatisfiedChannel(group, modelName, false)
	if err != nil {
		channel, err = model.CacheGetRandomSatisfiedChannel(group, modelName, true)
		if err != nil {
			return nil, fmt.Errorf("当前分组下没有可用渠道支持模型 %s", modelName)
		}
	}
	if channel.Status != model.ChannelStatusEnabled {
		return nil, fmt.Errorf("渠道 %s 当前不可用", channel.Name)
	}
	// 游乐场直连上游，仅支持 OpenAI 兼容请求格式的渠道（Azure 的 URL 形态特殊，除外）
	if channeltype.ToAPIType(channel.Type) != apitype.OpenAI || channel.Type == channeltype.Azure {
		return nil, fmt.Errorf("渠道 %s（类型 %d）暂不支持游乐场，仅支持 OpenAI 兼容渠道", channel.Name, channel.Type)
	}
	return channel, nil
}

// BuildUpstreamRequest 构造发往上游的请求体
func BuildUpstreamRequest(channel *model.Channel, modelName string, messages []Message, stream bool) map[string]any {
	actualModel := modelName
	if mapping := channel.GetModelMapping(); mapping != nil {
		if mapped, ok := mapping[modelName]; ok && mapped != "" {
			actualModel = mapped
		}
	}
	return map[string]any{
		"model":    actualModel,
		"messages": messages,
		"stream":   stream,
	}
}

// UpstreamRequestURL 按渠道类型拼装上游 /v1/chat/completions 地址（与 relay 逻辑一致）
func UpstreamRequestURL(channel *model.Channel) string {
	baseURL := channel.GetBaseURL()
	if baseURL == "" {
		baseURL = channeltype.ChannelBaseURLs[channel.Type]
	}
	return openaiadaptor.GetFullRequestURL(baseURL, "/v1/chat/completions", channel.Type)
}

// upstreamHTTPClient 流式请求不设整体超时
var upstreamHTTPClient = &http.Client{Timeout: 120 * time.Second}

// CallUpstream 调用上游 OpenAI 兼容接口
func CallUpstream(ctx context.Context, channel *model.Channel, body []byte, stream bool) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, UpstreamRequestURL(channel), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+channel.Key)
	req.Header.Set("Accept", "application/json")
	client := upstreamHTTPClient
	if stream {
		client = &http.Client{}
	}
	return client.Do(req)
}

// ParseNonStreamResponse 解析非流式响应
func ParseNonStreamResponse(body []byte) (string, map[string]any, error) {
	var resp struct {
		Choices []struct {
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage map[string]any `json:"usage"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", nil, fmt.Errorf("解析上游响应失败：%w", err)
	}
	if len(resp.Choices) == 0 {
		return "", nil, fmt.Errorf("上游响应中没有 choices")
	}
	content, err := ExtractContent(resp.Choices[0].Message.Content)
	if err != nil {
		return "", nil, err
	}
	return content, resp.Usage, nil
}

// ExtractContent 从原始 JSON content 中提取文本（兼容字符串与片段数组两种格式）
func ExtractContent(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "", nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err == nil {
		var sb strings.Builder
		for _, p := range parts {
			if p.Type == "text" || p.Type == "" {
				sb.WriteString(p.Text)
			}
		}
		return sb.String(), nil
	}
	return "", fmt.Errorf("无法解析上游响应内容")
}

// StreamChunk 流式响应中的一个增量片段
type StreamChunk struct {
	Type     string         `json:"type"` // "content" | "search_results"
	Content  string         `json:"content,omitempty"`
	Results  []SearchResult `json:"results,omitempty"`
	Error    string         `json:"error,omitempty"`
	Finished bool           `json:"finished,omitempty"`
}

// SSEEvent 序列化为 SSE data 行
func (c StreamChunk) SSEEvent() string {
	data, _ := json.Marshal(c)
	return "data: " + string(data) + "\n\n"
}

// StreamRelay 把上游 SSE 流转发给客户端，同时累积完整回复文本。
// onChunk 在每次拿到增量内容时回调（用于会话持久化）。
func StreamRelay(w io.Writer, flusher http.Flusher, body io.Reader, onChunk func(delta string)) (string, error) {
	reader := bufio.NewReader(body)
	var full strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			if _, werr := io.WriteString(w, line); werr != nil {
				return full.String(), werr
			}
			flusher.Flush()
			trimmed := strings.TrimRight(line, "\r\n")
			if strings.HasPrefix(trimmed, "data:") {
				payload := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
				if payload != "" && payload != "[DONE]" {
					delta := extractDeltaContent(payload)
					if delta != "" {
						full.WriteString(delta)
						if onChunk != nil {
							onChunk(delta)
						}
					}
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				return full.String(), nil
			}
			return full.String(), err
		}
	}
}

// extractDeltaContent 从 SSE data 行中提取增量文本
func extractDeltaContent(payload string) string {
	var chunk struct {
		Choices []struct {
			Delta struct {
				Content json.RawMessage `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
		return ""
	}
	if len(chunk.Choices) == 0 {
		return ""
	}
	text, err := ExtractContent(chunk.Choices[0].Delta.Content)
	if err != nil {
		return ""
	}
	return text
}

// ErrorBody 统一错误响应体
func ErrorBody(status int, message string) (int, map[string]any) {
	return status, map[string]any{
		"success": false,
		"message": message,
	}
}
