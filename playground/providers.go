package playground

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ---------- Tavily ----------

type tavilyProvider struct{}

func (tavilyProvider) ID() string        { return "tavily" }
func (tavilyProvider) Name() string      { return "Tavily" }
func (tavilyProvider) RequiresKey() bool { return true }
func (tavilyProvider) Search(ctx context.Context, query string, maxResults int, cfg *PlaygroundConfig) ([]SearchResult, error) {
	var resp struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}
	err := doJSONRequest(ctx, http.MethodPost, providerEndpoint(cfg, "tavily", "https://api.tavily.com/search"), nil, map[string]any{
		"api_key":      providerKey(cfg, "tavily"),
		"query":        query,
		"max_results":  maxResults,
		"search_depth": "basic",
	}, &resp)
	if err != nil {
		return nil, err
	}
	out := make([]SearchResult, 0, len(resp.Results))
	for _, r := range resp.Results {
		out = append(out, SearchResult{Title: r.Title, URL: r.URL, Snippet: r.Content})
	}
	return out, nil
}

// ---------- Serper（Google） ----------

type serperProvider struct{}

func (serperProvider) ID() string        { return "serper" }
func (serperProvider) Name() string      { return "Serper（Google）" }
func (serperProvider) RequiresKey() bool { return true }
func (serperProvider) Search(ctx context.Context, query string, maxResults int, cfg *PlaygroundConfig) ([]SearchResult, error) {
	var resp struct {
		Organic []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"organic"`
	}
	err := doJSONRequest(ctx, http.MethodPost, providerEndpoint(cfg, "serper", "https://google.serper.dev/search"),
		map[string]string{"X-API-Key": providerKey(cfg, "serper")},
		map[string]any{"q": query, "num": maxResults}, &resp)
	if err != nil {
		return nil, err
	}
	out := make([]SearchResult, 0, len(resp.Organic))
	for _, r := range resp.Organic {
		out = append(out, SearchResult{Title: r.Title, URL: r.Link, Snippet: r.Snippet})
	}
	return out, nil
}

// ---------- Brave Search ----------

type braveProvider struct{}

func (braveProvider) ID() string        { return "brave" }
func (braveProvider) Name() string      { return "Brave Search" }
func (braveProvider) RequiresKey() bool { return true }
func (braveProvider) Search(ctx context.Context, query string, maxResults int, cfg *PlaygroundConfig) ([]SearchResult, error) {
	u := providerEndpoint(cfg, "brave", "https://api.search.brave.com/res/v1/web/search")
	params := url.Values{}
	params.Set("q", query)
	params.Set("count", strconv.Itoa(maxResults))
	var resp struct {
		Web struct {
			Results []struct {
				Title       string `json:"title"`
				URL         string `json:"url"`
				Description string `json:"description"`
			} `json:"results"`
		} `json:"web"`
	}
	err := doJSONRequest(ctx, http.MethodGet, u+"?"+params.Encode(),
		map[string]string{"Accept": "application/json", "X-Subscription-Token": providerKey(cfg, "brave")}, nil, &resp)
	if err != nil {
		return nil, err
	}
	out := make([]SearchResult, 0, len(resp.Web.Results))
	for _, r := range resp.Web.Results {
		out = append(out, SearchResult{Title: r.Title, URL: r.URL, Snippet: r.Description})
	}
	return out, nil
}

// ---------- Exa ----------

type exaProvider struct{}

func (exaProvider) ID() string        { return "exa" }
func (exaProvider) Name() string      { return "Exa" }
func (exaProvider) RequiresKey() bool { return true }
func (exaProvider) Search(ctx context.Context, query string, maxResults int, cfg *PlaygroundConfig) ([]SearchResult, error) {
	var resp struct {
		Results []struct {
			Title string `json:"title"`
			URL   string `json:"url"`
			Text  string `json:"text"`
		} `json:"results"`
	}
	err := doJSONRequest(ctx, http.MethodPost, providerEndpoint(cfg, "exa", "https://api.exa.ai/search"),
		map[string]string{"x-api-key": providerKey(cfg, "exa")},
		map[string]any{
			"query":      query,
			"numResults": maxResults,
			"contents":   map[string]any{"text": map[string]any{"maxCharacters": 400}},
		}, &resp)
	if err != nil {
		return nil, err
	}
	out := make([]SearchResult, 0, len(resp.Results))
	for _, r := range resp.Results {
		out = append(out, SearchResult{Title: r.Title, URL: r.URL, Snippet: r.Text})
	}
	return out, nil
}

// ---------- Bing Web Search（Azure） ----------

type bingProvider struct{}

func (bingProvider) ID() string        { return "bing" }
func (bingProvider) Name() string      { return "Bing Web Search" }
func (bingProvider) RequiresKey() bool { return true }
func (bingProvider) Search(ctx context.Context, query string, maxResults int, cfg *PlaygroundConfig) ([]SearchResult, error) {
	u := providerEndpoint(cfg, "bing", "https://api.bing.microsoft.com/v7.0/search")
	params := url.Values{}
	params.Set("q", query)
	params.Set("count", strconv.Itoa(maxResults))
	params.Set("responseFilter", "Webpages")
	var resp struct {
		WebPages struct {
			Value []struct {
				Name    string `json:"name"`
				URL     string `json:"url"`
				Snippet string `json:"snippet"`
			} `json:"value"`
		} `json:"webPages"`
	}
	err := doJSONRequest(ctx, http.MethodGet, u+"?"+params.Encode(),
		map[string]string{"Ocp-Apim-Subscription-Key": providerKey(cfg, "bing")}, nil, &resp)
	if err != nil {
		return nil, err
	}
	out := make([]SearchResult, 0, len(resp.WebPages.Value))
	for _, r := range resp.WebPages.Value {
		out = append(out, SearchResult{Title: r.Name, URL: r.URL, Snippet: r.Snippet})
	}
	return out, nil
}

// ---------- 博查（Bocha） ----------

type bochaProvider struct{}

func (bochaProvider) ID() string        { return "bocha" }
func (bochaProvider) Name() string      { return "博查搜索" }
func (bochaProvider) RequiresKey() bool { return true }
func (bochaProvider) Search(ctx context.Context, query string, maxResults int, cfg *PlaygroundConfig) ([]SearchResult, error) {
	var resp struct {
		Code int `json:"code"`
		Data struct {
			WebPages struct {
				Value []struct {
					Name    string `json:"name"`
					URL     string `json:"url"`
					Snippet string `json:"snippet"`
					Summary string `json:"summary"`
				} `json:"value"`
			} `json:"webPages"`
		} `json:"data"`
		Message string `json:"message"`
	}
	err := doJSONRequest(ctx, http.MethodPost, providerEndpoint(cfg, "bocha", "https://api.bochaai.com/v1/web-search"),
		map[string]string{"Authorization": "Bearer " + providerKey(cfg, "bocha")},
		map[string]any{"query": query, "count": maxResults, "page": 1, "freshness": "noLimit"}, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 && resp.Code != 200 {
		msg := resp.Message
		if msg == "" {
			msg = fmt.Sprintf("code=%d", resp.Code)
		}
		return nil, fmt.Errorf("博查搜索返回错误：%s", msg)
	}
	out := make([]SearchResult, 0, len(resp.Data.WebPages.Value))
	for _, r := range resp.Data.WebPages.Value {
		snippet := r.Snippet
		if snippet == "" {
			snippet = r.Summary
		}
		out = append(out, SearchResult{Title: r.Name, URL: r.URL, Snippet: snippet})
	}
	return out, nil
}

// ---------- SearXNG（自建实例） ----------

type searxngProvider struct{}

func (searxngProvider) ID() string        { return "searxng" }
func (searxngProvider) Name() string      { return "SearXNG（自建）" }
func (searxngProvider) RequiresKey() bool { return true }
func (searxngProvider) Search(ctx context.Context, query string, maxResults int, cfg *PlaygroundConfig) ([]SearchResult, error) {
	base := strings.TrimSuffix(providerKey(cfg, "searxng"), "/")
	if base == "" {
		return nil, fmt.Errorf("请先填写 SearXNG 实例地址")
	}
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("pageno", "1")
	var resp struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}
	err := doJSONRequest(ctx, http.MethodGet, base+"/search?"+params.Encode(),
		map[string]string{"Accept": "application/json"}, nil, &resp)
	if err != nil {
		return nil, err
	}
	out := make([]SearchResult, 0, len(resp.Results))
	for _, r := range resp.Results {
		out = append(out, SearchResult{Title: r.Title, URL: r.URL, Snippet: r.Content})
	}
	return out, nil
}

// ---------- Perplexity（Sonar 在线模型） ----------

type perplexityProvider struct{}

func (perplexityProvider) ID() string        { return "perplexity" }
func (perplexityProvider) Name() string      { return "Perplexity（Sonar）" }
func (perplexityProvider) RequiresKey() bool { return true }
func (perplexityProvider) Search(ctx context.Context, query string, maxResults int, cfg *PlaygroundConfig) ([]SearchResult, error) {
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Citations []string `json:"citations"`
	}
	err := doJSONRequest(ctx, http.MethodPost, providerEndpoint(cfg, "perplexity", "https://api.perplexity.ai/chat/completions"),
		map[string]string{"Authorization": "Bearer " + providerKey(cfg, "perplexity")},
		map[string]any{
			"model": "sonar",
			"messages": []map[string]string{
				{"role": "system", "content": "你是一个搜索助手，请基于实时检索结果简洁回答用户问题。"},
				{"role": "user", "content": query},
			},
		}, &resp)
	if err != nil {
		return nil, err
	}
	out := make([]SearchResult, 0, len(resp.Citations))
	for i, cite := range resp.Citations {
		if len(out) >= maxResults {
			break
		}
		title := cite
		if u, err := url.Parse(cite); err == nil && u.Host != "" {
			title = u.Host
		}
		snippet := ""
		if i == 0 && len(resp.Choices) > 0 {
			snippet = resp.Choices[0].Message.Content
		}
		out = append(out, SearchResult{Title: title, URL: cite, Snippet: snippet})
	}
	return out, nil
}

func init() {
	registerProvider(tavilyProvider{})
	registerProvider(serperProvider{})
	registerProvider(braveProvider{})
	registerProvider(exaProvider{})
	registerProvider(bingProvider{})
	registerProvider(bochaProvider{})
	registerProvider(searxngProvider{})
	registerProvider(perplexityProvider{})
}
