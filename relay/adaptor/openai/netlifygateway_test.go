package openai

import (
	"testing"

	"github.com/songquanpeng/one-api/relay/channeltype"
)

func TestNetlifyAIGatewayCompatibleMeta(t *testing.T) {
	// 渠道必须在 OpenAI 兼容渠道列表中注册
	found := false
	for _, ct := range CompatibleChannels {
		if ct == channeltype.NetlifyAIGateway {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("NetlifyAIGateway not registered in CompatibleChannels")
	}

	name, models := GetCompatibleChannelMeta(channeltype.NetlifyAIGateway)
	if name != "netlify-ai-gateway" {
		t.Errorf("channel name = %q, want netlify-ai-gateway", name)
	}
	expects := []string{
		"gpt-5",                  // OpenAI 直托管
		"claude-opus-4-5",        // Anthropic 直托管
		"gemini-2.5-pro",         // Gemini 直托管
		"deepseek/deepseek-chat", // OpenRouter 段
		"moonshotai/kimi-k2",     // OpenRouter 段
	}
	for _, want := range expects {
		ok := false
		for _, m := range models {
			if m == want {
				ok = true
				break
			}
		}
		if !ok {
			t.Errorf("model %q missing from netlify gateway model list", want)
		}
	}
}
