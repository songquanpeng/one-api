package channeltype

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/songquanpeng/one-api/relay/apitype"
)

func TestChannelBaseURLs(t *testing.T) {
	Convey("channel base urls", t, func() {
		So(len(ChannelBaseURLs), ShouldEqual, Dummy)
	})
}

func TestNetlifyAIGatewayChannel(t *testing.T) {
	Convey("netlify ai gateway channel type", t, func() {
		// 基址为 Netlify AI Gateway 官方地址（不带 /v1，由 relay 拼接 /v1/chat/completions）
		So(ChannelBaseURLs[NetlifyAIGateway], ShouldEqual, "https://ai-gateway.netlify.com")
		// OpenAI 兼容渠道
		So(ToAPIType(NetlifyAIGateway), ShouldEqual, apitype.OpenAI)
		// 不与既有类型冲突
		So(NetlifyAIGateway, ShouldEqual, 61)
	})
}
