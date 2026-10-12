package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/playground"
)

// playgroundModelsInGroups 返回指定分组下可用的模型集合（并集）
func playgroundModelsInGroups(c *gin.Context, groups []string) map[string]bool {
	set := map[string]bool{}
	for _, g := range groups {
		models, err := model.GetGroupModels(c.Request.Context(), g)
		if err != nil {
			continue
		}
		for _, m := range models {
			set[m] = true
		}
	}
	return set
}

// GetPlaygroundConfig 返回游乐场页面所需的配置（不含任何密钥）
func GetPlaygroundConfig(c *gin.Context) {
	cfg := playground.GetConfig()
	if !cfg.Enabled {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "",
			"data":    gin.H{"enabled": false},
		})
		return
	}
	modelSet := playgroundModelsInGroups(c, cfg.Groups)
	models := make([]string, 0, len(modelSet))
	for m := range modelSet {
		models = append(models, m)
	}
	sort.Strings(models)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"enabled":          true,
			"models":           models,
			"groups":           cfg.Groups,
			"search_provider":  cfg.SearchProvider,
			"search_providers": playground.ProviderInfos(cfg),
			"limits": gin.H{
				"max_image_size_mb":   cfg.MaxImageSizeMB,
				"session_ttl_minutes": cfg.SessionTTLMinutes,
				"search_max_results":  cfg.SearchMaxResults,
			},
		},
	})
}

// CreatePlaygroundSession 创建临时会话
func CreatePlaygroundSession(c *gin.Context) {
	userId := c.GetInt("id")
	cfg := playground.GetConfig()
	var body struct {
		Model string `json:"model"`
		Group string `json:"group"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的请求参数"})
		return
	}
	body.Model = strings.TrimSpace(body.Model)
	body.Group = strings.TrimSpace(body.Group)
	if body.Group == "" {
		body.Group = "default"
	}
	if !playgroundContains(cfg.Groups, body.Group) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "非法的模型分组：" + body.Group})
		return
	}
	if body.Model == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "请选择模型"})
		return
	}
	// 校验该分组下确实存在此模型
	models, err := model.GetGroupModels(c.Request.Context(), body.Group)
	if err != nil || !playgroundContains(models, body.Model) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "模型 " + body.Model + " 在分组 " + body.Group + " 下不可用"})
		return
	}
	session := playground.CreateSession(userId, body.Model, body.Group, cfg.SessionTTLMinutes)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"id":         session.ID,
			"model":      session.Model,
			"group":      session.Group,
			"expires_at": session.ExpiresAt.Unix(),
		},
	})
}

// GetPlaygroundSession 读取临时会话
func GetPlaygroundSession(c *gin.Context) {
	userId := c.GetInt("id")
	session, err := playground.GetSession(c.Param("id"), userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"id":         session.ID,
			"model":      session.Model,
			"group":      session.Group,
			"messages":   session.Messages,
			"expires_at": session.ExpiresAt.Unix(),
		},
	})
}

// DeletePlaygroundSession 销毁临时会话
func DeletePlaygroundSession(c *gin.Context) {
	userId := c.GetInt("id")
	if err := playground.DeleteSession(c.Param("id"), userId); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

// UploadPlaygroundImage 上传图片（临时存储，供游乐场引用）
func UploadPlaygroundImage(c *gin.Context) {
	cfg := playground.GetConfig()
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "请选择要上传的图片文件"})
		return
	}
	name, err := playground.SaveUploadedImage(file, cfg.MaxImageSizeMB)
	if err != nil {
		msg := err.Error()
		if errors.Is(err, playground.ErrImageTooLarge) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "图片大小超出限制（最大 " + itoa(cfg.MaxImageSizeMB) + "MB）",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": false, "message": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"name": name,
			"url":  "/api/playground/uploads/" + name,
			"size": file.Size,
		},
	})
}

// ServePlaygroundUpload 提供临时图片的读取（带鉴权与防路径穿越）
func ServePlaygroundUpload(c *gin.Context) {
	name := c.Param("name")
	data, err := playground.ReadUploadFile(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "图片不存在或已过期"})
		return
	}
	c.Header("Content-Type", playground.GetUploadContentType(name))
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, playground.GetUploadContentType(name), data)
}

// PlaygroundChat 游乐场聊天（支持临时会话、图片、联网搜索、流式）
func PlaygroundChat(c *gin.Context) {
	userId := c.GetInt("id")
	cfg := playground.GetConfig()
	var req playground.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的请求参数"})
		return
	}
	req.Model = strings.TrimSpace(req.Model)
	req.Group = strings.TrimSpace(req.Group)
	if req.Group == "" {
		req.Group = "default"
	}
	if req.Model == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "请选择模型"})
		return
	}
	if !playgroundContains(cfg.Groups, req.Group) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "非法的模型分组：" + req.Group})
		return
	}
	messages := playground.NormalizeMessages(req.Messages)
	if len(messages) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "消息不能为空"})
		return
	}

	// 会话解析：客户端始终发送完整消息列表；服务端仅在成功应答后落盘到临时会话
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID != "" {
		if _, err := playground.GetSession(sessionID, userId); err != nil {
			sessionID = "" // 已过期或不存在，自动创建新会话承接本次请求
		}
	}
	if sessionID == "" {
		sessionID = playground.CreateSession(userId, req.Model, req.Group, cfg.SessionTTLMinutes).ID
	}
	_ = playground.UpdateSessionModel(sessionID, userId, req.Model, req.Group, cfg.SessionTTLMinutes)

	// 图片解析并附加到最后一条用户消息
	images, err := playground.ResolveImages(req.Images)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	messages = playground.AppendImagesToLastUserMessage(messages, images)

	// 联网搜索
	var searchResults []playground.SearchResult
	var searchError string
	if req.Search.Enabled && cfg.SearchProvider != "" {
		query := strings.TrimSpace(req.Search.Query)
		if query == "" {
			query = playground.ExtractLastUserText(messages)
		}
		if query != "" {
			results, err := playground.Search(c.Request.Context(), query, cfg)
			if err != nil {
				searchError = err.Error()
			} else {
				searchResults = results
				if ctx := playground.BuildSearchContext(results, playgroundProviderName(cfg.SearchProvider)); ctx != "" {
					messages = append([]playground.Message{{
						Role:    "system",
						Content: playground.TextContent(ctx),
					}}, messages...)
				}
			}
		}
	}

	// 选择渠道并调用上游
	channel, err := playground.SelectPlaygroundChannel(req.Group, req.Model)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	body := mustMarshalJSON(playground.BuildUpstreamRequest(channel, req.Model, messages, req.Stream))

	if !req.Stream {
		resp, err := playground.CallUpstream(c.Request.Context(), channel, body, false)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "调用上游失败：" + err.Error()})
			return
		}
		respBody := readLimited(resp.Body, 1<<20)
		_ = resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "上游返回错误（HTTP " + itoa(resp.StatusCode) + "）：" + truncateString(string(respBody), 300),
			})
			return
		}
		reply, usage, err := playground.ParseNonStreamResponse(respBody)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			return
		}
		if reply != "" {
			_ = playground.AppendMessages(sessionID, userId, cfg.SessionTTLMinutes,
				playground.Message{Role: "user", Content: lastUserContent(messages)},
				playground.Message{Role: "assistant", Content: playground.TextContent(reply)},
			)
		}
		c.JSON(http.StatusOK, gin.H{
			"success":         true,
			"message":         "",
			"session_id":      sessionID,
			"session_expired": req.SessionID != "" && req.SessionID != sessionID,
			"reply":           reply,
			"search_results":  searchResults,
			"search_error":    searchError,
			"usage":           usage,
			"channel_id":      channel.Id,
			"channel_name":    channel.Name,
		})
		return
	}

	// 流式响应（SSE 透传）
	resp, err := playground.CallUpstream(c.Request.Context(), channel, body, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "调用上游失败：" + err.Error()})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody := readLimited(resp.Body, 1<<20)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "上游返回错误（HTTP " + itoa(resp.StatusCode) + "）：" + truncateString(string(respBody), 300),
		})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "当前环境不支持流式输出"})
		return
	}

	writeSSE := func(chunk playground.StreamChunk) {
		_, _ = c.Writer.WriteString(chunk.SSEEvent())
		flusher.Flush()
	}
	writeSSE(playground.StreamChunk{Type: "meta", Content: sessionID})
	if searchError != "" {
		writeSSE(playground.StreamChunk{Type: "search_error", Error: searchError})
	}
	if len(searchResults) > 0 {
		writeSSE(playground.StreamChunk{Type: "search_results", Results: searchResults})
	}

	fullText, streamErr := playground.StreamRelay(c.Writer, flusher, resp.Body, nil)
	if fullText != "" {
		_ = playground.AppendMessages(sessionID, userId, cfg.SessionTTLMinutes,
			playground.Message{Role: "user", Content: lastUserContent(messages)},
			playground.Message{Role: "assistant", Content: playground.TextContent(fullText)},
		)
	}
	if streamErr != nil {
		writeSSE(playground.StreamChunk{Type: "error", Error: streamErr.Error()})
	}
	writeSSE(playground.StreamChunk{Type: "done", Finished: true, Content: sessionID})
}

// lastUserContent 取最后一条 user 消息的内容（用于会话持久化）
func lastUserContent(messages []playground.Message) []playground.ContentPart {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return messages[i].Content
		}
	}
	return nil
}

// playgroundProviderName 取搜索服务商展示名
func playgroundProviderName(id string) string {
	if p, ok := playground.GetSearchProvider(id); ok {
		return p.Name()
	}
	return id
}

func playgroundContains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

// itoa 整数转字符串
func itoa(n int) string {
	return strconv.Itoa(n)
}

// mustMarshalJSON JSON 序列化（游乐场请求体均为可序列化结构）
func mustMarshalJSON(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return data
}

// readLimited 读取响应体并限制大小
func readLimited(r io.Reader, limit int64) []byte {
	data, _ := io.ReadAll(io.LimitReader(r, limit))
	return data
}

// truncateString 截断字符串
func truncateString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
