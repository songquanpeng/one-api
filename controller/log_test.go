package controller

import (
	"testing"

	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func errWithCode(code string) *model.ErrorWithStatusCode {
	return &model.ErrorWithStatusCode{
		Error: model.Error{Code: code, Message: code},
	}
}

func TestShouldRecordErrorLog(t *testing.T) {
	cases := []struct {
		name     string
		mode     int
		code     string
		expected bool
	}{
		// 上游调用失败：计入分母
		{"do_request_failed", relaymode.ChatCompletions, "do_request_failed", true},
		{"upstream 400", relaymode.ChatCompletions, "invalid_api_key", true},
		{"upstream rate limit", relaymode.ChatCompletions, "rate_limit_exceeded", true},
		{"request_error", relaymode.ChatCompletions, "request_error", true},
		// 本地错误（未到达上游）：不计入分母
		{"invalid_text_request", relaymode.ChatCompletions, "invalid_text_request", false},
		{"insufficient_user_quota", relaymode.ChatCompletions, "insufficient_user_quota", false},
		{"pre_consume_token_quota_failed", relaymode.ChatCompletions, "pre_consume_token_quota_failed", false},
		{"convert_request_failed", relaymode.ChatCompletions, "convert_request_failed", false},
		{"invalid_api_type", relaymode.ChatCompletions, "invalid_api_type", false},
		{"invalid_json", relaymode.ChatCompletions, "invalid_json", false},
		{"prompt_missing", relaymode.ImagesGenerations, "prompt_missing", false},
		{"text_too_long", relaymode.AudioSpeech, "text_too_long", false},
		// Proxy 模式不记录成功日志，失败亦不记录
		{"proxy mode", relaymode.Proxy, "do_request_failed", false},
	}
	for _, tc := range cases {
		if got := shouldRecordErrorLog(tc.mode, errWithCode(tc.code)); got != tc.expected {
			t.Errorf("%s: expected %v, got %v", tc.name, tc.expected, got)
		}
	}
}

func TestCalcSuccessRate(t *testing.T) {
	cases := []struct {
		success  int64
		failures int64
		expected float64
	}{
		{0, 0, 0},
		{1, 0, 100},
		{0, 1, 0},
		{3, 1, 75},
		{19, 1, 95},                    // 95.00
		{2, 3, 40},                     // 40.00
		{1, 6, 14.29},                  // 14.2857... -> 14.29
		{999999, 1, 100},               // 99.9999 -> 100.00
		{1, 999999, 0},                 // 0.0001 -> 0.00
	}
	for _, tc := range cases {
		got := calcSuccessRate(tc.success, tc.failures)
		if got != tc.expected {
			t.Errorf("calcSuccessRate(%d, %d) = %v, expected %v", tc.success, tc.failures, got, tc.expected)
		}
	}
}
