package network

import (
	"context"
	"fmt"
	"github.com/songquanpeng/one-api/common/logger"
	"net"
)

func IsValidSubnet(subnet string) error {
	_, _, err := net.ParseCIDR(subnet)
	if err != nil {
		return fmt.Errorf("failed to parse subnet: %w", err)
	}
	return nil
}

func IsIpInSubnet(ctx context.Context, ip string, subnet string) bool {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		logger.Errorf(ctx, "failed to parse subnet: %s", err.Error())
		return false
	}
	return ipNet.Contains(net.ParseIP(ip))
}
