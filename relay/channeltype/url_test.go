package channeltype

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestChannelBaseURLs(t *testing.T) {
	Convey("channel base urls", t, func() {
		So(len(ChannelBaseURLs), ShouldEqual, Dummy)
	})
}
