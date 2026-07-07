package openai

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func TestGetRequestURL(t *testing.T) {
	Convey("open router image generation", t, func() {
		adaptor := Adaptor{}
		meta := &meta.Meta{
			Mode:           relaymode.ImagesGenerations,
			ChannelType:    channeltype.OpenRouter,
			BaseURL:        "https://openrouter.ai/api",
			RequestURLPath: "/v1/images/generations",
		}
		url, err := adaptor.GetRequestURL(meta)
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "https://openrouter.ai/api/v1/images")
	})
	Convey("open router chat completions", t, func() {
		adaptor := Adaptor{}
		meta := &meta.Meta{
			Mode:           relaymode.ChatCompletions,
			ChannelType:    channeltype.OpenRouter,
			BaseURL:        "https://openrouter.ai/api",
			RequestURLPath: "/v1/chat/completions",
		}
		url, err := adaptor.GetRequestURL(meta)
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "https://openrouter.ai/api/v1/chat/completions")
	})
}
