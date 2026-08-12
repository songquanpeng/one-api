package channeltype

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/songquanpeng/one-api/relay/apitype"
	"testing"
)

func TestChannelBaseURLs(t *testing.T) {
	Convey("channel base urls", t, func() {
		So(len(ChannelBaseURLs), ShouldEqual, Dummy)
	})
}

func TestSiftQChannelRegistration(t *testing.T) {
	Convey("SiftQ channel registration", t, func() {
		So(ChannelBaseURLs[SiftQ], ShouldEqual, "https://siftq.com/api/minimax/")
		So(ToAPIType(SiftQ), ShouldEqual, apitype.SiftQ)
	})
}
