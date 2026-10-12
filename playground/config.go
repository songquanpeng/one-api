// Package playground 提供「模型游乐场」能力：
// 管理员可开关的临时模型体验空间，支持临时会话、图片上传、
// 模型/分组选择与联网搜索（多家搜索服务商）。
// 注意：游乐场调用不经过计费链路，不消耗额度。
package playground

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/songquanpeng/one-api/common/config"
)

// PlaygroundConfig 游乐场配置（整体以 JSON 形式存放在 PlaygroundConfig 选项中）
type PlaygroundConfig struct {
	// Enabled 是否开启游乐场
	Enabled bool `json:"enabled"`
	// SearchProvider 搜索服务商 ID，空表示不启用联网搜索
	SearchProvider string `json:"search_provider"`
	// SearchKeys 各搜索服务商的密钥 / 实例地址（SearXNG 存实例地址）
	SearchKeys map[string]string `json:"search_keys"`
	// SearchEndpoints 可选：覆盖各服务商的默认接口地址
	SearchEndpoints map[string]string `json:"search_endpoints"`
	// SearchMaxResults 每次搜索返回的结果条数，默认 5
	SearchMaxResults int `json:"search_max_results"`
	// SessionTTLMinutes 临时会话保留时间（分钟），默认 120
	SessionTTLMinutes int `json:"session_ttl_minutes"`
	// MaxImageSizeMB 单张图片大小限制（MB），默认 5
	MaxImageSizeMB int `json:"max_image_size_mb"`
	// Groups 游乐场中可选的分组，为空时默认 ["default"]
	Groups []string `json:"groups"`
}

// DefaultConfig 返回默认配置（未开启）
func DefaultConfig() *PlaygroundConfig {
	return &PlaygroundConfig{
		Enabled:           false,
		SearchProvider:    "",
		SearchKeys:        map[string]string{},
		SearchEndpoints:   map[string]string{},
		SearchMaxResults:  5,
		SessionTTLMinutes: 120,
		MaxImageSizeMB:    5,
		Groups:            []string{"default"},
	}
}

// DefaultConfigJSON 默认配置的 JSON 形式
func DefaultConfigJSON() string {
	return `{"enabled":false,"search_provider":"","search_keys":{},"search_endpoints":{},"search_max_results":5,"session_ttl_minutes":120,"max_image_size_mb":5,"groups":["default"]}`
}

var (
	configCacheMu sync.RWMutex
	configCache   *PlaygroundConfig
)

// GetConfig 读取游乐场配置（带内存缓存，选项变更时缓存失效）
func GetConfig() *PlaygroundConfig {
	configCacheMu.RLock()
	cached := configCache
	configCacheMu.RUnlock()
	if cached != nil {
		return cached
	}

	cfg := DefaultConfig()
	raw := ""
	config.OptionMapRWMutex.RLock()
	raw = config.OptionMap["PlaygroundConfig"]
	config.OptionMapRWMutex.RUnlock()
	if raw != "" {
		parsed := &PlaygroundConfig{}
		if err := json.Unmarshal([]byte(raw), parsed); err == nil {
			cfg = parsed
		}
	}
	cfg.normalize()

	configCacheMu.Lock()
	configCache = cfg
	configCacheMu.Unlock()
	return cfg
}

// InvalidateConfigCache 使配置缓存失效（选项更新后调用）
func InvalidateConfigCache() {
	configCacheMu.Lock()
	configCache = nil
	configCacheMu.Unlock()
}

func (cfg *PlaygroundConfig) normalize() {
	if cfg.SearchKeys == nil {
		cfg.SearchKeys = map[string]string{}
	}
	if cfg.SearchEndpoints == nil {
		cfg.SearchEndpoints = map[string]string{}
	}
	if cfg.SearchMaxResults <= 0 {
		cfg.SearchMaxResults = 5
	}
	if cfg.SearchMaxResults > 20 {
		cfg.SearchMaxResults = 20
	}
	if cfg.SessionTTLMinutes <= 0 {
		cfg.SessionTTLMinutes = 120
	}
	if cfg.MaxImageSizeMB <= 0 {
		cfg.MaxImageSizeMB = 5
	}
	if cfg.MaxImageSizeMB > 50 {
		cfg.MaxImageSizeMB = 50
	}
	if len(cfg.Groups) == 0 {
		cfg.Groups = []string{"default"}
	}
}

// ValidateConfigJSON 校验管理员提交的配置 JSON
func ValidateConfigJSON(raw string) error {
	cfg := &PlaygroundConfig{}
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return err
	}
	cfg.normalize()
	if cfg.SearchProvider != "" {
		provider, ok := GetSearchProvider(cfg.SearchProvider)
		if !ok {
			return &ConfigError{"未知的搜索服务商：" + cfg.SearchProvider}
		}
		if provider.RequiresKey() && strings.TrimSpace(cfg.SearchKeys[provider.ID()]) == "" {
			return &ConfigError{"启用「" + provider.Name() + "」需要先填写对应的 API Key / 实例地址"}
		}
	}
	return nil
}

// ConfigError 配置错误
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
