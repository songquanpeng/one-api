package playground

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SearchResult 归一化后的搜索结果
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// SearchProvider 联网搜索服务商接口
type SearchProvider interface {
	// ID 服务商标识（配置中使用）
	ID() string
	// Name 展示名称
	Name() string
	// RequiresKey 是否必须配置 Key / 实例地址
	RequiresKey() bool
	// Search 执行搜索，cfg 中已填充该服务商的 Key / Endpoint
	Search(ctx context.Context, query string, maxResults int, cfg *PlaygroundConfig) ([]SearchResult, error)
}

var searchProviders = map[string]SearchProvider{}

func registerProvider(p SearchProvider) {
	searchProviders[p.ID()] = p
}

// GetSearchProvider 按 ID 获取服务商
func GetSearchProvider(id string) (SearchProvider, bool) {
	p, ok := searchProviders[id]
	return p, ok
}

// ProviderInfo 服务商信息（供前端展示）
type ProviderInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Configured bool   `json:"configured"`
}

// ProviderInfos 返回全部服务商及配置状态
func ProviderInfos(cfg *PlaygroundConfig) []ProviderInfo {
	infos := make([]ProviderInfo, 0, len(searchProviders))
	for _, p := range searchProviders {
		configured := true
		if p.RequiresKey() {
			configured = strings.TrimSpace(cfg.SearchKeys[p.ID()]) != ""
		}
		infos = append(infos, ProviderInfo{ID: p.ID(), Name: p.Name(), Configured: configured})
	}
	return infos
}

// Search 按配置执行联网搜索，结果已归一化并截断摘要长度
func Search(ctx context.Context, query string, cfg *PlaygroundConfig) ([]SearchResult, error) {
	provider, ok := GetSearchProvider(cfg.SearchProvider)
	if !ok {
		return nil, fmt.Errorf("未知的搜索服务商：%s", cfg.SearchProvider)
	}
	results, err := provider.Search(ctx, query, cfg.SearchMaxResults, cfg)
	if err != nil {
		return nil, err
	}
	// 归一化：截断过长摘要，过滤空结果
	out := make([]SearchResult, 0, len(results))
	for _, r := range results {
		r.Title = strings.TrimSpace(r.Title)
		r.URL = strings.TrimSpace(r.URL)
		r.Snippet = strings.TrimSpace(r.Snippet)
		if len(r.Snippet) > 500 {
			r.Snippet = r.Snippet[:500] + "…"
		}
		if r.URL == "" && r.Title == "" {
			continue
		}
		out = append(out, r)
		if len(out) >= cfg.SearchMaxResults {
			break
		}
	}
	return out, nil
}

// BuildSearchContext 把搜索结果拼装为注入模型的系统上下文
func BuildSearchContext(results []SearchResult, providerName string) string {
	if len(results) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("以下来自%s的实时联网搜索结果，请优先基于这些内容回答，并在引用信息时给出来源链接：", providerName))
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("\n[%d] %s\n链接: %s\n摘要: %s", i+1, r.Title, r.URL, r.Snippet))
	}
	return sb.String()
}

// providerEndpoint 取服务商接口地址（配置覆盖优先）
func providerEndpoint(cfg *PlaygroundConfig, id string, def string) string {
	if cfg.SearchEndpoints != nil {
		if ep := strings.TrimSpace(cfg.SearchEndpoints[id]); ep != "" {
			return ep
		}
	}
	return def
}

// providerKey 取服务商密钥
func providerKey(cfg *PlaygroundConfig, id string) string {
	if cfg.SearchKeys == nil {
		return ""
	}
	return strings.TrimSpace(cfg.SearchKeys[id])
}

// doJSONRequest 通用的 JSON 请求封装
func doJSONRequest(ctx context.Context, method, url string, headers map[string]string, body any, out any) error {
	var reqBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("搜索服务返回错误（HTTP %d）：%s", resp.StatusCode, truncate(string(respBody), 200))
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("解析搜索服务响应失败：%w", err)
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
