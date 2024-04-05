package network

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsIpInSubnet(t *testing.T) {
	ctx := context.Background()
	ip1 := "192.168.0.5"
	ip2 := "125.216.250.89"
	subnet := "192.168.0.0/24"
	Convey("TestIsIpInSubnet", t, func() {
		So(isIpInSubnet(ctx, ip1, subnet), ShouldBeTrue)
		So(isIpInSubnet(ctx, ip2, subnet), ShouldBeFalse)
	})
}
