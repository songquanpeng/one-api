package ratio

import (
	"math"
	"testing"

	"github.com/songquanpeng/one-api/relay/channeltype"
)

func assertRatio(t *testing.T, got float64, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.0000001 {
		t.Fatalf("ratio mismatch: got %v, want %v", got, want)
	}
}

func TestUpdatedDeepSeekAndBailianModelRatios(t *testing.T) {
	assertRatio(t, GetModelRatio("deepseek-v4-flash", channeltype.DeepSeek), 0.001*RMB)
	assertRatio(t, GetCompletionRatio("deepseek-v4-flash", channeltype.DeepSeek), 2)
	assertRatio(t, GetModelRatio("deepseek-v4-pro", channeltype.DeepSeek), 0.003*RMB)
	assertRatio(t, GetCompletionRatio("deepseek-v4-pro", channeltype.DeepSeek), 2)

	assertRatio(t, GetModelRatio("qwen3.6-plus", channeltype.AliBailian), 0.002*RMB)
	assertRatio(t, GetCompletionRatio("qwen3.6-plus", channeltype.AliBailian), 6)
	assertRatio(t, GetModelRatio("qwen3.6-flash", channeltype.AliBailian), 0.0012*RMB)
	assertRatio(t, GetCompletionRatio("qwen3.6-flash", channeltype.AliBailian), 6)

	assertRatio(t, GetModelRatio("deepseek-v3.2", channeltype.AliBailian), 0.002*RMB)
	assertRatio(t, GetCompletionRatio("deepseek-v3.2", channeltype.AliBailian), 1.5)
	assertRatio(t, GetModelRatio("kimi-k2.6", channeltype.AliBailian), 0.0065*RMB)
	assertRatio(t, GetCompletionRatio("kimi-k2.6", channeltype.AliBailian), 27.0/6.5)
}
