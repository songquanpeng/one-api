package ssrf

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/songquanpeng/one-api/common/logger"
)

// ListMode 定义列表模式类型
type ListMode string

const (
	// ModeBlacklist 黑名单模式：列表中的地址被禁止访问，其他地址放行
	ModeBlacklist ListMode = "blacklist"
	// ModeWhitelist 白名单模式：仅允许列表中的地址访问，其他地址全部拦截
	ModeWhitelist ListMode = "whitelist"
)

// SSRFProtectionConfig SSRF防护全局配置
type SSRFProtectionConfig struct {
	// Enabled 是否启用SSRF防护
	Enabled bool
	// AllowPrivateIP 是否允许访问私有IP地址
	AllowPrivateIP bool
	// ApplyIPFilterForDomain 对域名启用IP过滤（DNS解析后检查）
	ApplyIPFilterForDomain bool

	// DomainListMode 域名列表模式（whitelist/blacklist）
	// 黑名单模式：列表中的域名被禁止访问
	// 白名单模式：仅允许列表中的域名访问
	DomainListMode ListMode
	// DomainList 统一域名列表（支持通配符，如 *.example.com）
	// 根据 DomainListMode 决定是作为白名单还是黑名单使用
	DomainList []string

	// IPListMode IP列表模式（whitelist/blacklist）
	// 黑名单模式：列表中的IP被禁止访问
	// 白名单模式：仅允许列表中的IP访问
	IPListMode ListMode
	// IPList 统一IP列表（支持CIDR格式，如 192.168.1.0/24）
	// 根据 IPListMode 决定是作为白名单还是黑名单使用
	IPList []string

	// AllowedPorts 允许的端口列表
	AllowedPorts []int

	// DNSCacheTTL DNS缓存时间（秒），用于防止DNS重绑定攻击
	DNSCacheTTL int

	// MaxRedirects 最大重定向次数
	MaxRedirects int
}

// SSRFProtection SSRF防护实例
type SSRFProtection struct {
	config   *SSRFProtectionConfig
	dnsCache *dnsCache
	mu       sync.RWMutex
}

// dnsCache DNS缓存结构
type dnsCache struct {
	entries map[string]*dnsCacheEntry
	mu      sync.RWMutex
}

// dnsCacheEntry DNS缓存条目
type dnsCacheEntry struct {
	ips       []net.IP
	expiresAt time.Time
}

// 私有IP网段定义
var privateIPNets []*net.IPNet

func init() {
	// 初始化私有IP网段
	// 注意：IPv4-mapped IPv6地址 (::ffff:0:0/96) 不在此列表中，
	// 因为它会导致所有IPv4地址被错误识别为私有地址
	privateCIDRs := []string{
		// IPv4 私有地址
		"10.0.0.0/8",         // RFC 1918 - Class A
		"172.16.0.0/12",      // RFC 1918 - Class B
		"192.168.0.0/16",     // RFC 1918 - Class C
		"127.0.0.0/8",        // Loopback
		"169.254.0.0/16",     // Link-local
		"0.0.0.0/8",          // Current network
		"100.64.0.0/10",      // Carrier-grade NAT
		"192.0.0.0/24",       // IETF Protocol Assignments
		"192.0.2.0/24",       // TEST-NET-1
		"198.51.100.0/24",    // TEST-NET-2
		"203.0.113.0/24",     // TEST-NET-3
		"192.88.99.0/24",     // 6to4 Relay Anycast
		"198.18.0.0/15",      // Benchmarking
		"224.0.0.0/4",        // Multicast
		"240.0.0.0/4",        // Reserved for future use
		"255.255.255.255/32", // Broadcast

		// IPv6 私有地址
		"::1/128",       // Loopback
		"::/128",        // Unspecified
		"64:ff9b::/96",  // IPv4/IPv6 translation (NAT64)
		"100::/64",      // Discard prefix
		"2001::/32",     // Teredo
		"2001:10::/28",  // ORCHID
		"2001:20::/28",  // ORCHIDv2
		"2001:db8::/32", // Documentation
		"2002::/16",     // 6to4
		"fc00::/7",      // Unique local address
		"fe80::/10",     // Link-local
		"ff00::/8",      // Multicast
	}

	for _, cidr := range privateCIDRs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			privateIPNets = append(privateIPNets, ipNet)
		}
	}
}

// DefaultConfig 返回默认SSRF防护配置
func DefaultConfig() *SSRFProtectionConfig {
	return &SSRFProtectionConfig{
		Enabled:                true,
		AllowPrivateIP:         false,
		ApplyIPFilterForDomain: false,
		DomainListMode:         ModeBlacklist,
		DomainList:             []string{},
		IPListMode:             ModeBlacklist,
		IPList:                 []string{},
		AllowedPorts:           []int{80, 443},
		DNSCacheTTL:            60,
		MaxRedirects:           10,
	}
}

// globalProtection 全局SSRF防护实例
var globalProtection *SSRFProtection
var globalProtectionOnce sync.Once

// GetGlobalProtection 获取全局SSRF防护实例
func GetGlobalProtection() *SSRFProtection {
	globalProtectionOnce.Do(func() {
		globalProtection = NewSSRFProtection(DefaultConfig())
	})
	return globalProtection
}

// NewSSRFProtection 创建新的SSRF防护实例
func NewSSRFProtection(config *SSRFProtectionConfig) *SSRFProtection {
	if config == nil {
		config = DefaultConfig()
	}
	return &SSRFProtection{
		config: config,
		dnsCache: &dnsCache{
			entries: make(map[string]*dnsCacheEntry),
		},
	}
}

// UpdateConfig 更新SSRF防护配置
func (p *SSRFProtection) UpdateConfig(config *SSRFProtectionConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config = config
	// 清空DNS缓存
	p.dnsCache.mu.Lock()
	p.dnsCache.entries = make(map[string]*dnsCacheEntry)
	p.dnsCache.mu.Unlock()
}

// GetConfig 获取当前配置
func (p *SSRFProtection) GetConfig() *SSRFProtectionConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config
}

// IsEnabled 检查SSRF防护是否启用
func (p *SSRFProtection) IsEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config.Enabled
}

// IsPrivateIP 检查IP是否为私有地址
func IsPrivateIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	// 检查是否为零地址
	if ip.IsUnspecified() {
		return true
	}

	// 检查是否为环回地址
	if ip.IsLoopback() {
		return true
	}

	// 检查是否为链路本地地址
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// 检查是否为组播地址
	if ip.IsMulticast() {
		return true
	}

	// 检查是否在私有网段中
	for _, ipNet := range privateIPNets {
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}

// isIPInList 检查IP是否在列表中（支持CIDR格式）
func isIPInList(ip net.IP, list []string) bool {
	if len(list) == 0 {
		return false
	}

	for _, item := range list {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		// 尝试解析为CIDR
		if strings.Contains(item, "/") {
			_, ipNet, err := net.ParseCIDR(item)
			if err == nil && ipNet.Contains(ip) {
				return true
			}
		} else {
			// 作为单个IP处理
			listIP := net.ParseIP(item)
			if listIP != nil && ip.Equal(listIP) {
				return true
			}
		}
	}
	return false
}

// isDomainInList 检查域名是否在列表中（支持通配符）
func isDomainInList(domain string, list []string) bool {
	if len(list) == 0 {
		return false
	}

	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return false
	}

	for _, pattern := range list {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}

		// 精确匹配
		if domain == pattern {
			return true
		}

		// 通配符匹配 (*.example.com)
		if strings.HasPrefix(pattern, "*.") {
			suffix := strings.TrimPrefix(pattern, "*")
			// *.example.com 匹配 sub.example.com 和 example.com
			if strings.HasSuffix(domain, suffix) || domain == strings.TrimPrefix(suffix, ".") {
				return true
			}
		}
	}
	return false
}

// isPortAllowed 检查端口是否被允许
func (p *SSRFProtection) isPortAllowed(port int) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.config.AllowedPorts) == 0 {
		return true // 如果没有配置端口限制，则允许所有端口
	}

	for _, allowedPort := range p.config.AllowedPorts {
		if port == allowedPort {
			return true
		}
	}
	return false
}

// resolveWithCache 带缓存的DNS解析（防止DNS重绑定攻击）
func (p *SSRFProtection) resolveWithCache(ctx context.Context, host string) ([]net.IP, error) {
	p.dnsCache.mu.RLock()
	entry, exists := p.dnsCache.entries[host]
	p.dnsCache.mu.RUnlock()

	if exists && time.Now().Before(entry.expiresAt) {
		return entry.ips, nil
	}

	// DNS解析
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("DNS resolution failed for %s: %v", host, err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no IP addresses found for %s", host)
	}

	// 更新缓存
	p.mu.RLock()
	ttl := p.config.DNSCacheTTL
	p.mu.RUnlock()

	p.dnsCache.mu.Lock()
	p.dnsCache.entries[host] = &dnsCacheEntry{
		ips:       ips,
		expiresAt: time.Now().Add(time.Duration(ttl) * time.Second),
	}
	p.dnsCache.mu.Unlock()

	return ips, nil
}

// ValidateIP 验证IP地址是否允许访问
func (p *SSRFProtection) ValidateIP(ip net.IP) error {
	p.mu.RLock()
	config := p.config
	p.mu.RUnlock()

	// 根据IP列表模式进行验证
	inList := isIPInList(ip, config.IPList)

	if config.IPListMode == ModeWhitelist {
		// 白名单模式：仅允许列表中的IP访问
		if !inList {
			// 不在白名单中，检查是否为私有IP（私有IP有单独的控制开关）
			if IsPrivateIP(ip) {
				if config.AllowPrivateIP {
					return nil // 允许私有IP
				}
				return fmt.Errorf("private IP address not allowed: %s", ip.String())
			}
			return fmt.Errorf("IP address not in whitelist: %s", ip.String())
		}
		return nil // 在白名单中，允许访问
	} else {
		// 黑名单模式：列表中的IP被禁止访问
		if inList {
			return fmt.Errorf("IP address is blacklisted: %s", ip.String())
		}
	}

	// 检查私有IP（无论什么模式都需要检查）
	if !config.AllowPrivateIP && IsPrivateIP(ip) {
		return fmt.Errorf("private IP address not allowed: %s", ip.String())
	}

	return nil
}

// ValidateDomain 验证域名是否允许访问
func (p *SSRFProtection) ValidateDomain(domain string) error {
	p.mu.RLock()
	config := p.config
	p.mu.RUnlock()

	// 根据域名列表模式进行验证
	inList := isDomainInList(domain, config.DomainList)
	if config.DomainListMode == ModeWhitelist {
		// 白名单模式：仅允许列表中的域名访问
		if !inList {
			return fmt.Errorf("domain not in whitelist: %s", domain)
		}
		return nil // 在白名单中，允许访问
	} else {
		// 黑名单模式：列表中的域名被禁止访问
		if inList {
			return fmt.Errorf("domain is blacklisted: %s", domain)
		}
	}

	return nil
}

func (p *SSRFProtection) validateHost(ctx context.Context, host string) error {
	if ip := net.ParseIP(host); ip != nil {
		return p.ValidateIP(ip)
	}

	if err := p.ValidateDomain(host); err != nil {
		return err
	}

	p.mu.RLock()
	applyIPFilter := p.config.ApplyIPFilterForDomain
	p.mu.RUnlock()

	if applyIPFilter {
		ips, err := p.resolveWithCache(ctx, host)
		if err != nil {
			return err
		}
		for _, ip := range ips {
			if err := p.ValidateIP(ip); err != nil {
				return fmt.Errorf("%s resolves to blocked IP: %v", host, err)
			}
		}
	}

	return nil
}

func (p *SSRFProtection) ValidateURL(ctx context.Context, urlStr string) error {
	if !p.IsEnabled() {
		return nil
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("unsupported protocol: %s (only http/https allowed)", u.Scheme)
	}

	host := u.Hostname()
	portStr := u.Port()
	if portStr == "" {
		if scheme == "https" {
			portStr = "443"
		} else {
			portStr = "80"
		}
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port: %s", portStr)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("port out of range: %d", port)
	}

	if !p.isPortAllowed(port) {
		return fmt.Errorf("port %d is not allowed", port)
	}

	return p.validateHost(ctx, host)
}

// ValidateURLString 验证URL字符串（便捷方法）
func (p *SSRFProtection) ValidateURLString(urlStr string) error {
	return p.ValidateURL(context.Background(), urlStr)
}

func (p *SSRFProtection) CheckRedirect(ctx context.Context) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if !p.IsEnabled() {
			return nil
		}

		p.mu.RLock()
		maxRedirects := p.config.MaxRedirects
		p.mu.RUnlock()

		if len(via) >= maxRedirects {
			return fmt.Errorf("stopped after %d redirects", maxRedirects)
		}

		if err := p.ValidateURL(ctx, req.URL.String()); err != nil {
			logger.Warnf(ctx, "SSRF protection blocked redirect to %s: %v", req.URL.String(), err)
			return fmt.Errorf("redirect blocked by SSRF protection: %v", err)
		}

		return nil
	}
}

func (p *SSRFProtection) CreateSafeTransport(ctx context.Context, baseTransport *http.Transport) *http.Transport {
	if baseTransport == nil {
		baseTransport = &http.Transport{}
	}

	originalDialContext := baseTransport.DialContext
	if originalDialContext == nil {
		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
		originalDialContext = dialer.DialContext
	}

	baseTransport.DialContext = func(dialCtx context.Context, network, addr string) (net.Conn, error) {
		if !p.IsEnabled() {
			return originalDialContext(dialCtx, network, addr)
		}

		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		if err := p.validateHost(ctx, host); err != nil {
			logger.Warnf(ctx, "SSRF protection blocked connection to %s: %v", host, err)
			return nil, fmt.Errorf("connection blocked by SSRF protection: %v", err)
		}

		return originalDialContext(dialCtx, network, net.JoinHostPort(host, port))
	}

	return baseTransport
}

// ParsePortRanges 解析端口范围配置
// 支持格式: "80", "443", "8000-9000"
func ParsePortRanges(portConfigs []string) ([]int, error) {
	var ports []int
	seen := make(map[int]bool)

	for _, config := range portConfigs {
		config = strings.TrimSpace(config)
		if config == "" {
			continue
		}

		if strings.Contains(config, "-") {
			// 处理端口范围 "8000-9000"
			parts := strings.Split(config, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid port range format: %s", config)
			}

			startPort, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid start port in range %s: %v", config, err)
			}

			endPort, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid end port in range %s: %v", config, err)
			}

			if startPort > endPort {
				return nil, fmt.Errorf("invalid port range %s: start port cannot be greater than end port", config)
			}

			if startPort < 1 || startPort > 65535 || endPort < 1 || endPort > 65535 {
				return nil, fmt.Errorf("port range %s contains invalid port numbers (must be 1-65535)", config)
			}

			// 添加范围内的所有端口（去重）
			for port := startPort; port <= endPort; port++ {
				if !seen[port] {
					ports = append(ports, port)
					seen[port] = true
				}
			}
		} else {
			// 处理单个端口 "80"
			port, err := strconv.Atoi(config)
			if err != nil {
				return nil, fmt.Errorf("invalid port number: %s", config)
			}

			if port < 1 || port > 65535 {
				return nil, fmt.Errorf("invalid port number %d (must be 1-65535)", port)
			}

			if !seen[port] {
				ports = append(ports, port)
				seen[port] = true
			}
		}
	}

	return ports, nil
}

// ValidateURLWithConfig 使用指定配置验证URL（便捷函数）
func ValidateURLWithConfig(urlStr string, config *SSRFProtectionConfig) error {
	if config == nil || !config.Enabled {
		return nil
	}

	protection := NewSSRFProtection(config)
	return protection.ValidateURLString(urlStr)
}

// ValidateURLWithGlobalConfig 使用全局配置验证URL（便捷函数）
func ValidateURLWithGlobalConfig(ctx context.Context, urlStr string) error {
	return GetGlobalProtection().ValidateURL(ctx, urlStr)
}

// UpdateGlobalConfig 更新全局SSRF防护配置
func UpdateGlobalConfig(config *SSRFProtectionConfig) {
	GetGlobalProtection().UpdateConfig(config)
}

// IsIpInCIDRList 检查IP是否在CIDR列表中（兼容旧接口）
func IsIpInCIDRList(ip net.IP, cidrList []string) bool {
	return isIPInList(ip, cidrList)
}
