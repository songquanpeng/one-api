package playground

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"sync"
	"time"
)

// ContentPart 消息内容片段（OpenAI 兼容格式）
type ContentPart struct {
	Type     string    `json:"type"` // "text" | "image_url"
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL 图片引用
type ImageURL struct {
	URL string `json:"url"`
}

// Message 一条会话消息
type Message struct {
	Role    string        `json:"role"`
	Content []ContentPart `json:"content"`
}

// TextContent 快捷构造纯文本消息
func TextContent(text string) []ContentPart {
	return []ContentPart{{Type: "text", Text: text}}
}

// Session 临时会话（仅存内存，不落库）
type Session struct {
	ID        string
	UserID    int
	Model     string
	Group     string
	Messages  []Message
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time
}

// ErrSessionNotFound 会话不存在或已过期
var ErrSessionNotFound = errors.New("会话不存在或已过期")

const (
	maxSessions     = 1000 // 最多同时保留的会话数
	cleanupInterval = time.Minute
	sessionIDLength = 16
)

var (
	sessionMu   sync.RWMutex
	sessions    = make(map[string]*Session)
	cleanupOnce sync.Once
)

func startCleanupLoop() {
	cleanupOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(cleanupInterval)
			defer ticker.Stop()
			for range ticker.C {
				CleanupExpiredSessions()
			}
		}()
	})
}

func newSessionID() string {
	b := make([]byte, sessionIDLength)
	if _, err := rand.Read(b); err != nil {
		// 极端情况回退到时间戳
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405.000000000")))
	}
	return hex.EncodeToString(b)
}

// CreateSession 创建临时会话
func CreateSession(userID int, model, group string, ttlMinutes int) *Session {
	startCleanupLoop()
	now := time.Now()
	s := &Session{
		ID:        newSessionID(),
		UserID:    userID,
		Model:     model,
		Group:     group,
		Messages:  []Message{},
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(time.Duration(ttlMinutes) * time.Minute),
	}
	sessionMu.Lock()
	defer sessionMu.Unlock()
	sessions[s.ID] = s
	evictIfNeeded()
	return s
}

// GetSession 按 ID 获取会话（校验归属与过期）
func GetSession(id string, userID int) (*Session, error) {
	sessionMu.RLock()
	s, ok := sessions[id]
	sessionMu.RUnlock()
	if !ok || s.UserID != userID {
		return nil, ErrSessionNotFound
	}
	if time.Now().After(s.ExpiresAt) {
		DeleteSession(id, userID)
		return nil, ErrSessionNotFound
	}
	return s, nil
}

// AppendMessages 追加消息并刷新过期时间
func AppendMessages(id string, userID int, ttlMinutes int, msgs ...Message) error {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	s, ok := sessions[id]
	if !ok || s.UserID != userID {
		return ErrSessionNotFound
	}
	if time.Now().After(s.ExpiresAt) {
		delete(sessions, id)
		return ErrSessionNotFound
	}
	s.Messages = append(s.Messages, msgs...)
	s.UpdatedAt = time.Now()
	s.ExpiresAt = time.Now().Add(time.Duration(ttlMinutes) * time.Minute)
	return nil
}

// UpdateSessionModel 更新会话绑定的模型与分组
func UpdateSessionModel(id string, userID int, model, group string, ttlMinutes int) error {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	s, ok := sessions[id]
	if !ok || s.UserID != userID {
		return ErrSessionNotFound
	}
	if time.Now().After(s.ExpiresAt) {
		delete(sessions, id)
		return ErrSessionNotFound
	}
	if model != "" {
		s.Model = model
	}
	if group != "" {
		s.Group = group
	}
	s.ExpiresAt = time.Now().Add(time.Duration(ttlMinutes) * time.Minute)
	return nil
}

// DeleteSession 删除会话
func DeleteSession(id string, userID int) error {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	s, ok := sessions[id]
	if !ok || s.UserID != userID {
		return ErrSessionNotFound
	}
	delete(sessions, id)
	return nil
}

// CleanupExpiredSessions 清理过期会话
func CleanupExpiredSessions() {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	now := time.Now()
	for id, s := range sessions {
		if now.After(s.ExpiresAt) {
			delete(sessions, id)
		}
	}
}

// evictIfNeeded 超出容量时优先淘汰最早过期的会话（调用方需持有写锁）
func evictIfNeeded() {
	if len(sessions) <= maxSessions {
		return
	}
	type entry struct {
		id        string
		expiresAt time.Time
	}
	all := make([]entry, 0, len(sessions))
	for id, s := range sessions {
		all = append(all, entry{id, s.ExpiresAt})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].expiresAt.Before(all[j].expiresAt) })
	for _, e := range all[:len(all)-maxSessions] {
		delete(sessions, e.id)
	}
}
