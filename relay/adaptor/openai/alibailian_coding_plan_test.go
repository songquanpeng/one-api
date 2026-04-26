package openai

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func TestAliBailianCodingPlanGetRequestURL(t *testing.T) {
	Convey("Ali Bailian should accept the documented OpenAI-compatible base URL", t, func() {
		adaptor := &Adaptor{}
		m := &meta.Meta{
			ChannelType:    channeltype.AliBailian,
			Mode:           relaymode.ChatCompletions,
			BaseURL:        "https://dashscope.aliyuncs.com/compatible-mode/v1",
			RequestURLPath: "/v1/chat/completions",
		}

		requestURL, err := adaptor.GetRequestURL(m)

		So(err, ShouldBeNil)
		So(requestURL, ShouldEqual, "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions")
	})

	Convey("Ali Bailian should keep supporting the legacy root base URL", t, func() {
		adaptor := &Adaptor{}
		m := &meta.Meta{
			ChannelType:    channeltype.AliBailian,
			Mode:           relaymode.Embeddings,
			BaseURL:        "https://dashscope.aliyuncs.com",
			RequestURLPath: "/v1/embeddings",
		}

		requestURL, err := adaptor.GetRequestURL(m)

		So(err, ShouldBeNil)
		So(requestURL, ShouldEqual, "https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings")
	})

	Convey("Ali Bailian Coding Plan should use the dedicated OpenAI-compatible endpoint", t, func() {
		adaptor := &Adaptor{}
		m := &meta.Meta{
			ChannelType:    channeltype.AliBailianCodingPlan,
			Mode:           relaymode.ChatCompletions,
			BaseURL:        "https://coding.dashscope.aliyuncs.com/v1",
			RequestURLPath: "/v1/chat/completions",
		}

		requestURL, err := adaptor.GetRequestURL(m)

		So(err, ShouldBeNil)
		So(requestURL, ShouldEqual, "https://coding.dashscope.aliyuncs.com/v1/chat/completions")
	})

	Convey("Ali Bailian Coding Plan should reject unsupported embedding requests", t, func() {
		adaptor := &Adaptor{}
		m := &meta.Meta{
			ChannelType:    channeltype.AliBailianCodingPlan,
			Mode:           relaymode.Embeddings,
			BaseURL:        "https://coding.dashscope.aliyuncs.com/v1",
			RequestURLPath: "/v1/embeddings",
		}

		requestURL, err := adaptor.GetRequestURL(m)

		So(err, ShouldNotBeNil)
		So(requestURL, ShouldEqual, "")
	})
}

func TestAliBailianCodingPlanCompatibleMeta(t *testing.T) {
	Convey("Ali Bailian should expose the current Model Studio OpenAI-compatible presets", t, func() {
		name, models := GetCompatibleChannelMeta(channeltype.AliBailian)

		So(name, ShouldEqual, "alibailian")
		So(models, ShouldContain, "qwen3.6-max-preview")
		So(models, ShouldContain, "qwen3.6-flash")
		So(models, ShouldContain, "text-embedding-v4")
		So(models, ShouldContain, "deepseek-v4-pro")
		So(models, ShouldContain, "deepseek-v3.2")
		So(models, ShouldContain, "kimi-k2.6")
		So(models, ShouldContain, "glm-5.1")
		So(models, ShouldContain, "MiniMax/MiniMax-M2.7")
	})

	Convey("Ali Bailian Coding Plan should expose its own channel metadata and model list", t, func() {
		name, models := GetCompatibleChannelMeta(channeltype.AliBailianCodingPlan)

		So(name, ShouldEqual, "alibailian-coding-plan")
		So(models, ShouldContain, "qwen3.6-max-preview")
		So(models, ShouldContain, "qwen3.6-plus")
		So(models, ShouldContain, "qwen3.6-flash")
		So(models, ShouldContain, "qwen3-coder-flash")
		So(models, ShouldContain, "qwen3-coder-plus")
		So(models, ShouldContain, "kimi-k2.6")
		So(models, ShouldContain, "glm-5.1")
		So(models, ShouldContain, "MiniMax/MiniMax-M2.7")
	})

	Convey("DeepSeek should expose the V4 model names while retaining legacy aliases", t, func() {
		name, models := GetCompatibleChannelMeta(channeltype.DeepSeek)

		So(name, ShouldEqual, "deepseek")
		So(models, ShouldContain, "deepseek-v4-flash")
		So(models, ShouldContain, "deepseek-v4-pro")
		So(models, ShouldContain, "deepseek-chat")
		So(models, ShouldContain, "deepseek-reasoner")
	})
}
