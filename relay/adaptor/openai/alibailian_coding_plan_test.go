package openai

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func TestAliBailianCodingPlanGetRequestURL(t *testing.T) {
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
	Convey("Ali Bailian Coding Plan should expose its own channel metadata and model list", t, func() {
		name, models := GetCompatibleChannelMeta(channeltype.AliBailianCodingPlan)

		So(name, ShouldEqual, "alibailian-coding-plan")
		So(models, ShouldContain, "qwen3.6-plus")
		So(models, ShouldContain, "qwen3-coder-plus")
	})
}
