package ali_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common/test"
	_ "one-api/common/test/init"
	"one-api/providers"
	providers_base "one-api/providers/base"
	"one-api/types"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func getChatProvider(url string, context *gin.Context) providers_base.ChatInterface {
	channel := getAliChannel(url)
	provider := providers.GetProvider(&channel, context)
	chatProvider, _ := provider.(providers_base.ChatInterface)

	return chatProvider
}

func TestChatCompletions(t *testing.T) {
	url, server, teardown := setupAliTestServer()
	context, _ := test.GetContext("POST", "/v1/chat/completions", test.RequestJSONConfig(), nil)
	defer teardown()
	server.RegisterHandler("/api/v1/services/aigc/text-generation/generation", handleChatCompletionEndpoint)

	chatRequest := test.GetChatCompletionRequest("default", "qwen-turbo", "false")

	chatProvider := getChatProvider(url, context)
	usage := &types.Usage{}
	chatProvider.SetUsage(usage)
	response, errWithCode := chatProvider.CreateChatCompletion(chatRequest)

	assert.Nil(t, errWithCode)
	assert.IsType(t, &types.Usage{}, usage)
	assert.Equal(t, 33, usage.TotalTokens)
	assert.Equal(t, 14, usage.PromptTokens)
	assert.Equal(t, 19, usage.CompletionTokens)

	// 转换成JSON字符串
	responseBody, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
		assert.Fail(t, "json marshal error")
	}
	fmt.Println(string(responseBody))

	test.CheckChat(t, response, "qwen-turbo", usage)
}

func TestChatCompletionsError(t *testing.T) {
	url, server, teardown := setupAliTestServer()
	context, _ := test.GetContext("POST", "/v1/chat/completions", test.RequestJSONConfig(), nil)
	defer teardown()
	server.RegisterHandler("/api/v1/services/aigc/text-generation/generation", handleChatCompletionErrorEndpoint)

	chatRequest := test.GetChatCompletionRequest("default", "qwen-turbo", "false")

	chatProvider := getChatProvider(url, context)
	_, err := chatProvider.CreateChatCompletion(chatRequest)
	usage := chatProvider.GetUsage()

	assert.NotNil(t, err)
	assert.Nil(t, usage)
	assert.Equal(t, "InvalidParameter", err.Code)
}

// func TestChatCompletionsStream(t *testing.T) {
// 	url, server, teardown := setupAliTestServer()
// 	context, w := test.GetContext("POST", "/v1/chat/completions", test.RequestJSONConfig(), nil)
// 	defer teardown()
// 	server.RegisterHandler("/api/v1/services/aigc/text-generation/generation", handleChatCompletionStreamEndpoint)

// 	channel := getAliChannel(url)
// 	provider := providers.GetProvider(&channel, context)
// 	chatProvider, _ := provider.(providers_base.ChatInterface)
// 	chatRequest := test.GetChatCompletionRequest("default", "qwen-turbo", "true")

// 	usage := &types.Usage{}
// 	chatProvider.SetUsage(usage)
// 	response, errWithCode := chatProvider.CreateChatCompletionStream(chatRequest)
// 	assert.Nil(t, errWithCode)

// 	assert.IsType(t, &types.Usage{}, usage)
// 	assert.Equal(t, 16, usage.TotalTokens)
// 	assert.Equal(t, 8, usage.PromptTokens)
// 	assert.Equal(t, 8, usage.CompletionTokens)

// 	streamResponseCheck(t, w.Body.String())
// }

// func TestChatCompletionsStreamError(t *testing.T) {
// 	url, server, teardown := setupAliTestServer()
// 	context, w := test.GetContext("POST", "/v1/chat/completions", test.RequestJSONConfig(), nil)
// 	defer teardown()
// 	server.RegisterHandler("/api/v1/services/aigc/text-generation/generation", handleChatCompletionStreamErrorEndpoint)

// 	channel := getAliChannel(url)
// 	provider := providers.GetProvider(&channel, context)
// 	chatProvider, _ := provider.(providers_base.ChatInterface)
// 	chatRequest := test.GetChatCompletionRequest("default", "qwen-turbo", "true")

// 	usage, err := chatProvider.ChatAction(chatRequest, 0)

// 	// 打印 context 写入的内容
// 	fmt.Println(w.Body.String())

// 	assert.NotNil(t, err)
// 	assert.Nil(t, usage)
// }

// func TestChatImageCompletions(t *testing.T) {
// 	url, server, teardown := setupAliTestServer()
// 	context, _ := test.GetContext("POST", "/v1/chat/completions", test.RequestJSONConfig(), nil)
// 	defer teardown()
// 	server.RegisterHandler("/api/v1/services/aigc/multimodal-generation/generation", handleChatImageCompletionEndpoint)

// 	channel := getAliChannel(url)
// 	provider := providers.GetProvider(&channel, context)
// 	chatProvider, _ := provider.(providers_base.ChatInterface)
// 	chatRequest := test.GetChatCompletionRequest("image", "qwen-vl-plus", "false")

// 	usage, err := chatProvider.ChatAction(chatRequest, 0)

// 	assert.Nil(t, err)
// 	assert.IsType(t, &types.Usage{}, usage)
// 	assert.Equal(t, 1306, usage.TotalTokens)
// 	assert.Equal(t, 1279, usage.PromptTokens)
// 	assert.Equal(t, 27, usage.CompletionTokens)
// }

// func TestChatImageCompletionsStream(t *testing.T) {
// 	url, server, teardown := setupAliTestServer()
// 	context, w := test.GetContext("POST", "/v1/chat/completions", test.RequestJSONConfig(), nil)
// 	defer teardown()
// 	server.RegisterHandler("/api/v1/services/aigc/multimodal-generation/generation", handleChatImageCompletionStreamEndpoint)

// 	channel := getAliChannel(url)
// 	provider := providers.GetProvider(&channel, context)
// 	chatProvider, _ := provider.(providers_base.ChatInterface)
// 	chatRequest := test.GetChatCompletionRequest("image", "qwen-vl-plus", "true")

// 	usage, err := chatProvider.ChatAction(chatRequest, 0)

// 	fmt.Println(w.Body.String())

// 	assert.Nil(t, err)
// 	assert.IsType(t, &types.Usage{}, usage)
// 	assert.Equal(t, 1342, usage.TotalTokens)
// 	assert.Equal(t, 1279, usage.PromptTokens)
// 	assert.Equal(t, 63, usage.CompletionTokens)
// 	streamResponseCheck(t, w.Body.String())
// }

func handleChatCompletionEndpoint(w http.ResponseWriter, r *http.Request) {
	// completions only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	response := `{"output":{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":"您好！我可以帮您查询最近的公园，请问您现在所在的位置是哪里呢？"}}]},"usage":{"total_tokens":33,"output_tokens":19,"input_tokens":14},"request_id":"2479f818-9717-9b0b-9769-0d26e873a3f6"}`

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, response)
}

func handleChatCompletionErrorEndpoint(w http.ResponseWriter, r *http.Request) {
	// completions only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	response := `{"code":"InvalidParameter","message":"Role must be user or assistant and Content length must be greater than 0","request_id":"4883ee8d-f095-94ff-a94a-5ce0a94bc81f"}`

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, response)
}

func handleChatCompletionStreamEndpoint(w http.ResponseWriter, r *http.Request) {
	// completions only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	// 检测头部是否有X-DashScope-SSE: enable
	if r.Header.Get("X-DashScope-SSE") != "enable" {
		http.Error(w, "Header X-DashScope-SSE not found", http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "text/event-stream")

	// Send test responses
	dataBytes := []byte{}
	dataBytes = append(dataBytes, []byte("id:1\n")...)
	dataBytes = append(dataBytes, []byte("event:result\n")...)
	dataBytes = append(dataBytes, []byte(":HTTP_STATUS/200\n")...)
	//nolint:lll
	data := `{"output":{"choices":[{"message":{"content":"你好！","role":"assistant"},"finish_reason":"null"}]},"usage":{"total_tokens":10,"input_tokens":8,"output_tokens":2},"request_id":"215a2614-5486-936c-8d42-3b472d6fbd1c"}`
	dataBytes = append(dataBytes, []byte("data:"+data+"\n\n")...)

	dataBytes = append(dataBytes, []byte("id:2\n")...)
	dataBytes = append(dataBytes, []byte("event:result\n")...)
	dataBytes = append(dataBytes, []byte(":HTTP_STATUS/200\n")...)
	//nolint:lll
	data = `{"output":{"choices":[{"message":{"content":"有什么我可以帮助你的吗？","role":"assistant"},"finish_reason":"null"}]},"usage":{"total_tokens":16,"input_tokens":8,"output_tokens":8},"request_id":"215a2614-5486-936c-8d42-3b472d6fbd1c"}`
	dataBytes = append(dataBytes, []byte("data:"+data+"\n\n")...)

	dataBytes = append(dataBytes, []byte("id:3\n")...)
	dataBytes = append(dataBytes, []byte("event:result\n")...)
	dataBytes = append(dataBytes, []byte(":HTTP_STATUS/200\n")...)
	//nolint:lll
	data = `{"output":{"choices":[{"message":{"content":"","role":"assistant"},"finish_reason":"stop"}]},"usage":{"total_tokens":16,"input_tokens":8,"output_tokens":8},"request_id":"215a2614-5486-936c-8d42-3b472d6fbd1c"}`
	dataBytes = append(dataBytes, []byte("data:"+data+"\n\n")...)

	_, err := w.Write(dataBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleChatCompletionStreamErrorEndpoint(w http.ResponseWriter, r *http.Request) {
	// completions only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	// 检测头部是否有X-DashScope-SSE: enable
	if r.Header.Get("X-DashScope-SSE") != "enable" {
		http.Error(w, "Header X-DashScope-SSE not found", http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "text/event-stream")

	// Send test responses
	dataBytes := []byte{}
	dataBytes = append(dataBytes, []byte("id:1\n")...)
	dataBytes = append(dataBytes, []byte("event:error\n")...)
	dataBytes = append(dataBytes, []byte(":HTTP_STATUS/400\n")...)
	//nolint:lll
	data := `{"code":"InvalidParameter","message":"Role must be user or assistant and Content length must be greater than 0","request_id":"6b932ba9-41bd-9ad3-b430-24bc1e125880"}`
	dataBytes = append(dataBytes, []byte("data:"+data+"\n\n")...)

	_, err := w.Write(dataBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleChatImageCompletionEndpoint(w http.ResponseWriter, r *http.Request) {
	// completions only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	response := `{"output":{"finish_reason":"stop","choices":[{"message":{"role":"assistant","content":[{"text":"这张照片展示的是一个海滩的场景，但是并没有明确指出具体的位置。可以看到海浪和日落背景下的沙滩景色。"}]}}]},"usage":{"output_tokens":27,"input_tokens":1279,"image_tokens":1247},"request_id":"a360d53b-b993-927f-9a68-bef6b2b2042e"}`

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, response)
}

func handleChatImageCompletionStreamEndpoint(w http.ResponseWriter, r *http.Request) {
	// completions only accepts POST requests
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	// 检测头部是否有X-DashScope-SSE: enable
	if r.Header.Get("X-DashScope-SSE") != "enable" {
		http.Error(w, "Header X-DashScope-SSE not found", http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "text/event-stream")

	// Send test responses
	dataBytes := []byte{}
	dataBytes = append(dataBytes, []byte("id:1\n")...)
	dataBytes = append(dataBytes, []byte("event:result\n")...)
	dataBytes = append(dataBytes, []byte(":HTTP_STATUS/200\n")...)
	//nolint:lll
	data := `{"output":{"choices":[{"message":{"content":[{"text":"这张"}],"role":"assistant"}}],"finish_reason":"null"},"usage":{"input_tokens":1279,"output_tokens":1,"image_tokens":1247},"request_id":"37bead8b-d87a-98f8-9193-b9e2da9d2451"}`
	dataBytes = append(dataBytes, []byte("data:"+data+"\n\n")...)

	dataBytes = append(dataBytes, []byte("id:2\n")...)
	dataBytes = append(dataBytes, []byte("event:result\n")...)
	dataBytes = append(dataBytes, []byte(":HTTP_STATUS/200\n")...)
	//nolint:lll
	data = `{"output":{"choices":[{"message":{"content":[{"text":"这张照片"}],"role":"assistant"}}],"finish_reason":"null"},"usage":{"input_tokens":1279,"output_tokens":2,"image_tokens":1247},"request_id":"37bead8b-d87a-98f8-9193-b9e2da9d2451"}`
	dataBytes = append(dataBytes, []byte("data:"+data+"\n\n")...)

	dataBytes = append(dataBytes, []byte("id:3\n")...)
	dataBytes = append(dataBytes, []byte("event:result\n")...)
	dataBytes = append(dataBytes, []byte(":HTTP_STATUS/200\n")...)
	//nolint:lll
	data = `{"output":{"choices":[{"message":{"content":[{"text":"这张照片展示的是一个海滩的场景，具体来说是在日落时分。由于没有明显的地标或建筑物等特征可以辨认出具体的地点信息，所以无法确定这是哪个地方的海滩。但是根据图像中的元素和环境特点，我们可以推测这可能是一个位于沿海地区的沙滩海岸线。"}],"role":"assistant"}}],"finish_reason":"stop"},"usage":{"input_tokens":1279,"output_tokens":63,"image_tokens":1247},"request_id":"37bead8b-d87a-98f8-9193-b9e2da9d2451"}`
	dataBytes = append(dataBytes, []byte("data:"+data+"\n\n")...)

	_, err := w.Write(dataBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func streamResponseCheck(t *testing.T, response string) {
	// 以换行符分割response
	lines := strings.Split(response, "\n\n")
	// 如果最后一行为空，则删除最后一行
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// 循环遍历每一行
	for _, line := range lines {
		if line == "" {
			continue
		}
		// assert判断 是否以data: 开头
		assert.True(t, strings.HasPrefix(line, "data: "))
	}

	// 检测最后一行是否以data: [DONE] 结尾
	assert.True(t, strings.HasSuffix(lines[len(lines)-1], "data: [DONE]"))
	// 检测倒数第二行是否存在 `"finish_reason":"stop"`
	assert.True(t, strings.Contains(lines[len(lines)-2], `"finish_reason":"stop"`))
}
