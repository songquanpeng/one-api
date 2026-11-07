package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
)

// newRateLimitedRouter returns a router protected by CriticalRateLimit, backed by the
// in-memory limiter (Redis disabled) so the test needs no external services.
// The limiter is a package-level singleton keyed by mark+ClientIP(), so each test gets
// its own client address (see nextClientAddr) instead of resetting shared state.
func newRateLimitedRouter(t *testing.T) *gin.Engine {
	t.Helper()
	originalMode := gin.Mode()
	originalRedis := common.RedisEnabled
	t.Cleanup(func() {
		gin.SetMode(originalMode)
		common.RedisEnabled = originalRedis
	})
	gin.SetMode(gin.TestMode)
	common.RedisEnabled = false

	router := gin.New()
	router.POST("/api/user/login", CriticalRateLimit(), func(c *gin.Context) {
		c.Status(http.StatusUnauthorized)
	})
	return router
}

var clientCounter atomic.Uint32

// nextClientAddr returns a client address never used before in this process, so
// limiter state from earlier tests (or earlier -count runs) cannot leak in.
func nextClientAddr() string {
	n := clientCounter.Add(1)
	return fmt.Sprintf("192.0.2.%d:%d", n%250+1, 1000+n)
}

func hitLogin(router *gin.Engine, clientAddr string) int {
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", nil)
	req.RemoteAddr = clientAddr
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder.Code
}

func withRateLimitConfig(t *testing.T, debugEnabled bool, maxRequestNum int) {
	t.Helper()
	originalDebug := config.DebugEnabled
	originalNum := config.CriticalRateLimitNum
	t.Cleanup(func() {
		config.DebugEnabled = originalDebug
		config.CriticalRateLimitNum = originalNum
	})
	config.DebugEnabled = debugEnabled
	config.CriticalRateLimitNum = maxRequestNum
}

// CriticalRateLimit guards the login endpoint, so it has to keep working when DEBUG=true:
// debug mode is a logging verbosity switch, not a security switch.
// Regression test for https://github.com/songquanpeng/one-api/issues/2466
func TestCriticalRateLimitRemainsActiveInDebugMode(t *testing.T) {
	withRateLimitConfig(t, true, 3)
	router := newRateLimitedRouter(t)
	clientAddr := nextClientAddr()

	for i := 0; i < config.CriticalRateLimitNum; i++ {
		if code := hitLogin(router, clientAddr); code != http.StatusUnauthorized {
			t.Fatalf("request %d within the limit of %d should reach the handler (401), got %d", i+1, config.CriticalRateLimitNum, code)
		}
	}
	if code := hitLogin(router, clientAddr); code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after %d requests with DEBUG=true, got %d", config.CriticalRateLimitNum, code)
	}
}

// Setting maxRequestNum to 0 stays an explicit opt-out, with or without DEBUG.
func TestRateLimitDisabledWhenMaxRequestNumIsZero(t *testing.T) {
	for _, debugEnabled := range []bool{true, false} {
		withRateLimitConfig(t, debugEnabled, 0)
		router := newRateLimitedRouter(t)
		clientAddr := nextClientAddr()
		for i := 0; i < 5; i++ {
			if code := hitLogin(router, clientAddr); code != http.StatusUnauthorized {
				t.Fatalf("rate limit should be off when maxRequestNum is 0 (DEBUG=%v), got %d on request %d", debugEnabled, code, i+1)
			}
		}
	}
}

// Without debug mode the limiter must behave exactly as before.
func TestCriticalRateLimitActiveWhenDebugDisabled(t *testing.T) {
	withRateLimitConfig(t, false, 3)
	router := newRateLimitedRouter(t)
	clientAddr := nextClientAddr()

	for i := 0; i < config.CriticalRateLimitNum; i++ {
		if code := hitLogin(router, clientAddr); code != http.StatusUnauthorized {
			t.Fatalf("request %d within the limit should reach the handler (401), got %d", i+1, code)
		}
	}
	if code := hitLogin(router, clientAddr); code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after %d requests, got %d", config.CriticalRateLimitNum, code)
	}
}
