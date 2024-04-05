package network

import (
	"context"
	"github.com/songquanpeng/one-api/common/logger"
	"net"
)

func IsIpInSubnet(ctx context.Context, ip string, subnet string) bool {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		logger.Errorf(ctx, "failed to parse subnet: %s, subnet: %s", err.Error(), subnet)
		return false
	}
	return ipNet.Contains(net.ParseIP(ip))
}
