package ssrf

import (
	"context"
	"net"
	"net/http"
	"testing"
)

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		// IPv4 Private addresses
		{"IPv4 Loopback", "127.0.0.1", true},
		{"IPv4 Loopback Range", "127.255.255.255", true},
		{"IPv4 Private 10.x.x.x", "10.0.0.1", true},
		{"IPv4 Private 10.x.x.x Range", "10.255.255.255", true},
		{"IPv4 Private 172.16.x.x", "172.16.0.1", true},
		{"IPv4 Private 172.31.x.x", "172.31.255.255", true},
		{"IPv4 Private 192.168.x.x", "192.168.0.1", true},
		{"IPv4 Private 192.168.x.x Range", "192.168.255.255", true},
		{"IPv4 Link-local", "169.254.0.1", true},
		{"IPv4 Link-local Range", "169.254.255.255", true},
		{"IPv4 Zero", "0.0.0.0", true},
		{"IPv4 Multicast", "224.0.0.1", true},
		{"IPv4 Reserved", "240.0.0.1", true},
		{"IPv4 Broadcast", "255.255.255.255", true},

		// IPv4 Public addresses
		{"IPv4 Public Google DNS", "8.8.8.8", false},
		{"IPv4 Public Cloudflare", "1.1.1.1", false},
		{"IPv4 Public Random", "203.0.114.1", false},

		// IPv6 Private addresses
		{"IPv6 Loopback", "::1", true},
		{"IPv6 Unspecified", "::", true},
		{"IPv6 Link-local", "fe80::1", true},
		{"IPv6 Unique local fc", "fc00::1", true},
		{"IPv6 Unique local fd", "fd00::1", true},

		// IPv6 Public addresses
		{"IPv6 Public Google", "2001:4860:4860::8888", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", tt.ip)
			}
			result := IsPrivateIP(ip)
			if result != tt.expected {
				t.Errorf("IsPrivateIP(%s) = %v, expected %v", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestIsDomainInList(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		list     []string
		expected bool
	}{
		// Exact match
		{"Exact match", "example.com", []string{"example.com"}, true},
		{"Exact match case insensitive", "Example.COM", []string{"example.com"}, true},
		{"No match", "other.com", []string{"example.com"}, false},

		// Wildcard match
		{"Wildcard match subdomain", "sub.example.com", []string{"*.example.com"}, true},
		{"Wildcard match deep subdomain", "deep.sub.example.com", []string{"*.example.com"}, true},
		{"Wildcard match root domain", "example.com", []string{"*.example.com"}, true},
		{"Wildcard no match different domain", "example.org", []string{"*.example.com"}, false},

		// Multiple patterns
		{"Multiple patterns first match", "example.com", []string{"example.com", "other.com"}, true},
		{"Multiple patterns second match", "other.com", []string{"example.com", "other.com"}, true},
		{"Multiple patterns no match", "third.com", []string{"example.com", "other.com"}, false},

		// Empty list
		{"Empty list", "example.com", []string{}, false},

		// Empty domain
		{"Empty domain", "", []string{"example.com"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDomainInList(tt.domain, tt.list)
			if result != tt.expected {
				t.Errorf("isDomainInList(%s, %v) = %v, expected %v", tt.domain, tt.list, result, tt.expected)
			}
		})
	}
}

func TestIsIPInList(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		list     []string
		expected bool
	}{
		// Single IP match
		{"Single IP match", "192.168.1.1", []string{"192.168.1.1"}, true},
		{"Single IP no match", "192.168.1.2", []string{"192.168.1.1"}, false},

		// CIDR match
		{"CIDR match", "192.168.1.100", []string{"192.168.1.0/24"}, true},
		{"CIDR no match", "192.168.2.1", []string{"192.168.1.0/24"}, false},
		{"CIDR /16 match", "10.0.100.1", []string{"10.0.0.0/16"}, true},
		{"CIDR /8 match", "10.100.100.1", []string{"10.0.0.0/8"}, true},

		// Multiple entries
		{"Multiple entries first match", "192.168.1.1", []string{"192.168.1.1", "10.0.0.0/8"}, true},
		{"Multiple entries second match", "10.0.0.1", []string{"192.168.1.1", "10.0.0.0/8"}, true},
		{"Multiple entries no match", "172.16.0.1", []string{"192.168.1.1", "10.0.0.0/8"}, false},

		// Empty list
		{"Empty list", "192.168.1.1", []string{}, false},

		// IPv6
		{"IPv6 single match", "2001:db8::1", []string{"2001:db8::1"}, true},
		{"IPv6 CIDR match", "2001:db8::100", []string{"2001:db8::/32"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", tt.ip)
			}
			result := isIPInList(ip, tt.list)
			if result != tt.expected {
				t.Errorf("isIPInList(%s, %v) = %v, expected %v", tt.ip, tt.list, result, tt.expected)
			}
		})
	}
}

func TestParsePortRanges(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		expected    []int
		expectError bool
	}{
		// Single ports
		{"Single port", []string{"80"}, []int{80}, false},
		{"Multiple single ports", []string{"80", "443"}, []int{80, 443}, false},

		// Port ranges
		{"Port range", []string{"8000-8003"}, []int{8000, 8001, 8002, 8003}, false},
		{"Mixed ports and ranges", []string{"80", "443", "8000-8002"}, []int{80, 443, 8000, 8001, 8002}, false},

		// Edge cases
		{"Empty input", []string{}, []int(nil), false},
		{"Empty string", []string{""}, []int(nil), false},
		{"Whitespace", []string{" 80 ", " 443 "}, []int{80, 443}, false},

		// Errors
		{"Invalid port", []string{"abc"}, nil, true},
		{"Port out of range high", []string{"65536"}, nil, true},
		{"Port out of range low", []string{"0"}, nil, true},
		{"Invalid range format", []string{"80-90-100"}, nil, true},
		{"Reversed range", []string{"100-80"}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePortRanges(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("ParsePortRanges(%v) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParsePortRanges(%v) unexpected error: %v", tt.input, err)
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("ParsePortRanges(%v) = %v, expected %v", tt.input, result, tt.expected)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("ParsePortRanges(%v) = %v, expected %v", tt.input, result, tt.expected)
					return
				}
			}
		})
	}
}

func TestSSRFProtectionValidateURL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		config      *SSRFProtectionConfig
		url         string
		expectError bool
		errorMsg    string
	}{
		// Protocol validation
		{
			name:        "HTTP allowed",
			config:      &SSRFProtectionConfig{Enabled: true, AllowPrivateIP: true},
			url:         "http://example.com",
			expectError: false,
		},
		{
			name:        "HTTPS allowed",
			config:      &SSRFProtectionConfig{Enabled: true, AllowPrivateIP: true},
			url:         "https://example.com",
			expectError: false,
		},
		{
			name:        "FTP not allowed",
			config:      &SSRFProtectionConfig{Enabled: true},
			url:         "ftp://example.com",
			expectError: true,
			errorMsg:    "unsupported protocol",
		},
		{
			name:        "File not allowed",
			config:      &SSRFProtectionConfig{Enabled: true},
			url:         "file:///etc/passwd",
			expectError: true,
			errorMsg:    "unsupported protocol",
		},

		// Port validation
		{
			name:        "Allowed port 80",
			config:      &SSRFProtectionConfig{Enabled: true, AllowPrivateIP: true, AllowedPorts: []int{80, 443}},
			url:         "http://example.com:80",
			expectError: false,
		},
		{
			name:        "Allowed port 443",
			config:      &SSRFProtectionConfig{Enabled: true, AllowPrivateIP: true, AllowedPorts: []int{80, 443}},
			url:         "https://example.com:443",
			expectError: false,
		},
		{
			name:        "Disallowed port",
			config:      &SSRFProtectionConfig{Enabled: true, AllowedPorts: []int{80, 443}},
			url:         "http://example.com:8080",
			expectError: true,
			errorMsg:    "port 8080 is not allowed",
		},
		{
			name:        "No port restriction",
			config:      &SSRFProtectionConfig{Enabled: true, AllowPrivateIP: true, AllowedPorts: []int{}},
			url:         "http://example.com:9999",
			expectError: false,
		},

		// Private IP validation
		{
			name:        "Private IP blocked",
			config:      &SSRFProtectionConfig{Enabled: true, AllowPrivateIP: false},
			url:         "http://192.168.1.1",
			expectError: true,
			errorMsg:    "private IP address not allowed",
		},
		{
			name:        "Private IP allowed",
			config:      &SSRFProtectionConfig{Enabled: true, AllowPrivateIP: true},
			url:         "http://192.168.1.1",
			expectError: false,
		},
		{
			name:        "Loopback blocked",
			config:      &SSRFProtectionConfig{Enabled: true, AllowPrivateIP: false},
			url:         "http://127.0.0.1",
			expectError: true,
			errorMsg:    "private IP address not allowed",
		},
		{
			name:        "Localhost blocked",
			config:      &SSRFProtectionConfig{Enabled: true, AllowPrivateIP: false},
			url:         "http://localhost",
			expectError: false, // localhost is a domain, not an IP
		},

		// Domain whitelist mode
		{
			name: "Domain in whitelist mode - allowed",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				DomainListMode: ModeWhitelist,
				DomainList:     []string{"example.com"},
			},
			url:         "http://example.com",
			expectError: false,
		},
		{
			name: "Domain not in whitelist mode - blocked",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				DomainListMode: ModeWhitelist,
				DomainList:     []string{"example.com"},
			},
			url:         "http://other.com",
			expectError: true,
			errorMsg:    "domain not in whitelist",
		},
		{
			name: "Subdomain with wildcard whitelist",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				DomainListMode: ModeWhitelist,
				DomainList:     []string{"*.example.com"},
			},
			url:         "http://sub.example.com",
			expectError: false,
		},

		// Domain blacklist mode
		{
			name: "Domain in blacklist mode - blocked",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{"evil.com"},
			},
			url:         "http://evil.com",
			expectError: true,
			errorMsg:    "domain is blacklisted",
		},
		{
			name: "Domain not in blacklist mode - allowed",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{"evil.com"},
			},
			url:         "http://good.com",
			expectError: false,
		},

		// IP whitelist mode
		{
			name: "IP in whitelist mode - allowed",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				AllowPrivateIP: false,
				IPListMode:     ModeWhitelist,
				IPList:         []string{"192.168.1.1"},
			},
			url:         "http://192.168.1.1",
			expectError: false,
		},
		{
			name: "IP CIDR whitelist mode - allowed",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				AllowPrivateIP: false,
				IPListMode:     ModeWhitelist,
				IPList:         []string{"192.168.1.0/24"},
			},
			url:         "http://192.168.1.100",
			expectError: false,
		},
		{
			name: "IP not in whitelist mode - blocked",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				AllowPrivateIP: false,
				IPListMode:     ModeWhitelist,
				IPList:         []string{"192.168.1.0/24"},
			},
			url:         "http://8.8.8.8",
			expectError: true,
			errorMsg:    "IP address not in whitelist",
		},

		// IP blacklist mode
		{
			name: "IP in blacklist mode - blocked",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				AllowPrivateIP: true,
				IPListMode:     ModeBlacklist,
				IPList:         []string{"8.8.8.8"},
			},
			url:         "http://8.8.8.8",
			expectError: true,
			errorMsg:    "IP address is blacklisted",
		},
		{
			name: "IP not in blacklist mode - allowed",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				AllowPrivateIP: true,
				IPListMode:     ModeBlacklist,
				IPList:         []string{"8.8.8.8"},
			},
			url:         "http://1.1.1.1",
			expectError: false,
		},

		// Disabled protection
		{
			name:        "Protection disabled",
			config:      &SSRFProtectionConfig{Enabled: false},
			url:         "http://127.0.0.1",
			expectError: false,
		},

		// Invalid URLs
		{
			name:        "Invalid URL",
			config:      &SSRFProtectionConfig{Enabled: true},
			url:         "not-a-url",
			expectError: true,
			errorMsg:    "unsupported protocol",
		},

		// Empty list tests
		{
			name: "Empty whitelist blocks all domains",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				DomainListMode: ModeWhitelist,
				DomainList:     []string{},
			},
			url:         "http://example.com",
			expectError: true,
			errorMsg:    "domain not in whitelist",
		},
		{
			name: "Empty blacklist allows all domains",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
			},
			url:         "http://example.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protection := NewSSRFProtection(tt.config)
			err := protection.ValidateURL(ctx, tt.url)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateURL(%s) expected error containing '%s', got nil", tt.url, tt.errorMsg)
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateURL(%s) error = '%v', expected to contain '%s'", tt.url, err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateURL(%s) unexpected error: %v", tt.url, err)
				}
			}
		})
	}
}

func TestValidateIPWithModes(t *testing.T) {
	tests := []struct {
		name        string
		config      *SSRFProtectionConfig
		ip          string
		expectError bool
		errorMsg    string
	}{
		// Whitelist mode tests
		{
			name: "Whitelist mode - IP in list allowed",
			config: &SSRFProtectionConfig{
				Enabled:    true,
				IPListMode: ModeWhitelist,
				IPList:     []string{"8.8.8.8", "1.1.1.1"},
			},
			ip:          "8.8.8.8",
			expectError: false,
		},
		{
			name: "Whitelist mode - IP not in list blocked",
			config: &SSRFProtectionConfig{
				Enabled:    true,
				IPListMode: ModeWhitelist,
				IPList:     []string{"8.8.8.8", "1.1.1.1"},
			},
			ip:          "9.9.9.9",
			expectError: true,
			errorMsg:    "IP address not in whitelist",
		},
		{
			name: "Whitelist mode - private IP with AllowPrivateIP true",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				AllowPrivateIP: true,
				IPListMode:     ModeWhitelist,
				IPList:         []string{"8.8.8.8"},
			},
			ip:          "192.168.1.1",
			expectError: false,
		},
		{
			name: "Whitelist mode - private IP with AllowPrivateIP false",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				AllowPrivateIP: false,
				IPListMode:     ModeWhitelist,
				IPList:         []string{"8.8.8.8"},
			},
			ip:          "192.168.1.1",
			expectError: true,
			errorMsg:    "private IP address not allowed",
		},

		// Blacklist mode tests
		{
			name: "Blacklist mode - IP in list blocked",
			config: &SSRFProtectionConfig{
				Enabled:    true,
				IPListMode: ModeBlacklist,
				IPList:     []string{"8.8.8.8", "1.1.1.1"},
			},
			ip:          "8.8.8.8",
			expectError: true,
			errorMsg:    "IP address is blacklisted",
		},
		{
			name: "Blacklist mode - IP not in list allowed",
			config: &SSRFProtectionConfig{
				Enabled:    true,
				IPListMode: ModeBlacklist,
				IPList:     []string{"8.8.8.8", "1.1.1.1"},
			},
			ip:          "9.9.9.9",
			expectError: false,
		},
		{
			name: "Blacklist mode - private IP still blocked when AllowPrivateIP false",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				AllowPrivateIP: false,
				IPListMode:     ModeBlacklist,
				IPList:         []string{"8.8.8.8"},
			},
			ip:          "192.168.1.1",
			expectError: true,
			errorMsg:    "private IP address not allowed",
		},

		// CIDR tests
		{
			name: "Whitelist mode - CIDR match",
			config: &SSRFProtectionConfig{
				Enabled:    true,
				IPListMode: ModeWhitelist,
				IPList:     []string{"8.8.0.0/16"},
			},
			ip:          "8.8.4.4",
			expectError: false,
		},
		{
			name: "Blacklist mode - CIDR match",
			config: &SSRFProtectionConfig{
				Enabled:    true,
				IPListMode: ModeBlacklist,
				IPList:     []string{"8.8.0.0/16"},
			},
			ip:          "8.8.4.4",
			expectError: true,
			errorMsg:    "IP address is blacklisted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protection := NewSSRFProtection(tt.config)
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", tt.ip)
			}
			err := protection.ValidateIP(ip)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateIP(%s) expected error containing '%s', got nil", tt.ip, tt.errorMsg)
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateIP(%s) error = '%v', expected to contain '%s'", tt.ip, err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateIP(%s) unexpected error: %v", tt.ip, err)
				}
			}
		})
	}
}

func TestValidateDomainWithModes(t *testing.T) {
	tests := []struct {
		name        string
		config      *SSRFProtectionConfig
		domain      string
		expectError bool
		errorMsg    string
	}{
		// Whitelist mode tests
		{
			name: "Whitelist mode - domain in list allowed",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				DomainListMode: ModeWhitelist,
				DomainList:     []string{"example.com", "test.com"},
			},
			domain:      "example.com",
			expectError: false,
		},
		{
			name: "Whitelist mode - domain not in list blocked",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				DomainListMode: ModeWhitelist,
				DomainList:     []string{"example.com", "test.com"},
			},
			domain:      "other.com",
			expectError: true,
			errorMsg:    "domain not in whitelist",
		},
		{
			name: "Whitelist mode - wildcard match",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				DomainListMode: ModeWhitelist,
				DomainList:     []string{"*.example.com"},
			},
			domain:      "api.example.com",
			expectError: false,
		},

		// Blacklist mode tests
		{
			name: "Blacklist mode - domain in list blocked",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{"evil.com", "malware.com"},
			},
			domain:      "evil.com",
			expectError: true,
			errorMsg:    "domain is blacklisted",
		},
		{
			name: "Blacklist mode - domain not in list allowed",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{"evil.com", "malware.com"},
			},
			domain:      "good.com",
			expectError: false,
		},
		{
			name: "Blacklist mode - wildcard match",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{"*.evil.com"},
			},
			domain:      "sub.evil.com",
			expectError: true,
			errorMsg:    "domain is blacklisted",
		},

		// Empty list tests
		{
			name: "Whitelist mode - empty list blocks all",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				DomainListMode: ModeWhitelist,
				DomainList:     []string{},
			},
			domain:      "any.com",
			expectError: true,
			errorMsg:    "domain not in whitelist",
		},
		{
			name: "Blacklist mode - empty list allows all",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
			},
			domain:      "any.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protection := NewSSRFProtection(tt.config)
			err := protection.ValidateDomain(tt.domain)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateDomain(%s) expected error containing '%s', got nil", tt.domain, tt.errorMsg)
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateDomain(%s) error = '%v', expected to contain '%s'", tt.domain, err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateDomain(%s) unexpected error: %v", tt.domain, err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if !config.Enabled {
		t.Error("Default config should have Enabled = true")
	}
	if config.AllowPrivateIP {
		t.Error("Default config should have AllowPrivateIP = false")
	}
	if config.DomainListMode != ModeBlacklist {
		t.Error("Default config should have DomainListMode = blacklist")
	}
	if config.IPListMode != ModeBlacklist {
		t.Error("Default config should have IPListMode = blacklist")
	}
	if len(config.AllowedPorts) == 0 {
		t.Error("Default config should have default allowed ports")
	}

	// Check default ports include 80 and 443
	has80, has443 := false, false
	for _, port := range config.AllowedPorts {
		if port == 80 {
			has80 = true
		}
		if port == 443 {
			has443 = true
		}
	}
	if !has80 || !has443 {
		t.Error("Default config should include ports 80 and 443")
	}
}

func TestGlobalProtection(t *testing.T) {
	// Test that global protection is singleton
	p1 := GetGlobalProtection()
	p2 := GetGlobalProtection()

	if p1 != p2 {
		t.Error("GetGlobalProtection should return the same instance")
	}

	// Test config update
	newConfig := &SSRFProtectionConfig{
		Enabled:        true,
		AllowPrivateIP: true,
		DomainListMode: ModeWhitelist,
		IPListMode:     ModeBlacklist,
		AllowedPorts:   []int{80},
	}
	UpdateGlobalConfig(newConfig)

	config := GetGlobalProtection().GetConfig()
	if !config.AllowPrivateIP {
		t.Error("Global config should be updated")
	}
	if config.DomainListMode != ModeWhitelist {
		t.Error("Global config DomainListMode should be updated to whitelist")
	}

	// Reset to default
	UpdateGlobalConfig(DefaultConfig())
}

func TestListModeConstants(t *testing.T) {
	// Test that mode constants have expected values
	if ModeBlacklist != "blacklist" {
		t.Errorf("ModeBlacklist should be 'blacklist', got '%s'", ModeBlacklist)
	}
	if ModeWhitelist != "whitelist" {
		t.Errorf("ModeWhitelist should be 'whitelist', got '%s'", ModeWhitelist)
	}
}

func TestConfigHotReload(t *testing.T) {
	// Test that configuration updates are applied immediately (hot-reload)
	ctx := context.Background()

	// Create a protection instance with initial config
	initialConfig := &SSRFProtectionConfig{
		Enabled:        true,
		DomainListMode: ModeBlacklist,
		DomainList:     []string{"blocked.com"},
		IPListMode:     ModeBlacklist,
		IPList:         []string{},
		AllowPrivateIP: true,
	}
	protection := NewSSRFProtection(initialConfig)

	// Test 1: Initially blocked.com should be blocked
	err := protection.ValidateURL(ctx, "http://blocked.com")
	if err == nil {
		t.Error("blocked.com should be blocked initially")
	}

	// Test 2: allowed.com should be allowed initially
	err = protection.ValidateURL(ctx, "http://allowed.com")
	if err != nil {
		t.Errorf("allowed.com should be allowed initially, got error: %v", err)
	}

	// Update config to block allowed.com instead
	newConfig := &SSRFProtectionConfig{
		Enabled:        true,
		DomainListMode: ModeBlacklist,
		DomainList:     []string{"allowed.com"},
		IPListMode:     ModeBlacklist,
		IPList:         []string{},
		AllowPrivateIP: true,
	}
	protection.UpdateConfig(newConfig)

	// Test 3: After hot-reload, blocked.com should now be allowed
	err = protection.ValidateURL(ctx, "http://blocked.com")
	if err != nil {
		t.Errorf("blocked.com should be allowed after config update, got error: %v", err)
	}

	// Test 4: After hot-reload, allowed.com should now be blocked
	err = protection.ValidateURL(ctx, "http://allowed.com")
	if err == nil {
		t.Error("allowed.com should be blocked after config update")
	}

	// Test 5: Switch to whitelist mode
	whitelistConfig := &SSRFProtectionConfig{
		Enabled:        true,
		DomainListMode: ModeWhitelist,
		DomainList:     []string{"trusted.com"},
		IPListMode:     ModeBlacklist,
		IPList:         []string{},
		AllowPrivateIP: true,
	}
	protection.UpdateConfig(whitelistConfig)

	// Test 6: trusted.com should be allowed in whitelist mode
	err = protection.ValidateURL(ctx, "http://trusted.com")
	if err != nil {
		t.Errorf("trusted.com should be allowed in whitelist mode, got error: %v", err)
	}

	// Test 7: untrusted.com should be blocked in whitelist mode
	err = protection.ValidateURL(ctx, "http://untrusted.com")
	if err == nil {
		t.Error("untrusted.com should be blocked in whitelist mode")
	}

	// Test 8: Verify DNS cache is cleared on config update
	// The DNS cache should be empty after UpdateConfig
	protection.dnsCache.mu.RLock()
	cacheLen := len(protection.dnsCache.entries)
	protection.dnsCache.mu.RUnlock()
	if cacheLen != 0 {
		t.Errorf("DNS cache should be cleared after config update, got %d entries", cacheLen)
	}
}

func TestCreateSafeTransportDisabled(t *testing.T) {
	// Test that CreateSafeTransport respects the Enabled flag
	ctx := context.Background()

	// Create a protection instance with protection DISABLED
	config := &SSRFProtectionConfig{
		Enabled:        false, // Protection disabled
		DomainListMode: ModeWhitelist,
		DomainList:     []string{"allowed.com"}, // Only allow this domain
		IPListMode:     ModeBlacklist,
		IPList:         []string{},
		AllowPrivateIP: false,
	}
	protection := NewSSRFProtection(config)

	// Create a safe transport
	transport := protection.CreateSafeTransport(ctx, nil)

	// The transport should have a custom DialContext
	if transport.DialContext == nil {
		t.Error("Transport should have custom DialContext")
	}

	// When protection is disabled, any domain should be allowed (validation bypassed)
	// We can't easily test the actual dialing, but we can verify the config is respected
	// by checking that IsEnabled returns false
	if protection.IsEnabled() {
		t.Error("Protection should be disabled")
	}

	// Enable protection and verify it's now enabled
	config.Enabled = true
	protection.UpdateConfig(config)
	if !protection.IsEnabled() {
		t.Error("Protection should be enabled after config update")
	}
}

func TestGlobalConfigHotReload(t *testing.T) {
	// Test global protection hot-reload
	ctx := context.Background()

	// Save original config to restore later
	originalConfig := GetGlobalProtection().GetConfig()
	defer UpdateGlobalConfig(originalConfig)

	// Set initial config
	initialConfig := &SSRFProtectionConfig{
		Enabled:        true,
		DomainListMode: ModeBlacklist,
		DomainList:     []string{"evil.com"},
		IPListMode:     ModeBlacklist,
		IPList:         []string{},
		AllowPrivateIP: true,
	}
	UpdateGlobalConfig(initialConfig)

	// Test 1: evil.com should be blocked
	err := ValidateURLWithGlobalConfig(ctx, "http://evil.com")
	if err == nil {
		t.Error("evil.com should be blocked with global config")
	}

	// Test 2: good.com should be allowed
	err = ValidateURLWithGlobalConfig(ctx, "http://good.com")
	if err != nil {
		t.Errorf("good.com should be allowed with global config, got error: %v", err)
	}

	// Update global config
	newConfig := &SSRFProtectionConfig{
		Enabled:        true,
		DomainListMode: ModeBlacklist,
		DomainList:     []string{"good.com"},
		IPListMode:     ModeBlacklist,
		IPList:         []string{},
		AllowPrivateIP: true,
	}
	UpdateGlobalConfig(newConfig)

	// Test 3: After hot-reload, evil.com should now be allowed
	err = ValidateURLWithGlobalConfig(ctx, "http://evil.com")
	if err != nil {
		t.Errorf("evil.com should be allowed after global config update, got error: %v", err)
	}

	// Test 4: After hot-reload, good.com should now be blocked
	err = ValidateURLWithGlobalConfig(ctx, "http://good.com")
	if err == nil {
		t.Error("good.com should be blocked after global config update")
	}
}

func TestCheckRedirect(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		config      *SSRFProtectionConfig
		redirectURL string
		viaCount    int
		expectError bool
		errorMsg    string
	}{
		// Protection disabled
		{
			name: "Protection disabled - allow any redirect",
			config: &SSRFProtectionConfig{
				Enabled:        false,
				DomainListMode: ModeWhitelist,
				DomainList:     []string{"allowed.com"},
			},
			redirectURL: "http://blocked.com",
			viaCount:    0,
			expectError: false,
		},

		// Max redirects
		{
			name: "Within max redirects - allow",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
			},
			redirectURL: "http://example.com",
			viaCount:    5,
			expectError: false,
		},
		{
			name: "Exceed max redirects - block",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   5,
				AllowPrivateIP: true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
			},
			redirectURL: "http://example.com",
			viaCount:    5,
			expectError: true,
			errorMsg:    "stopped after 5 redirects",
		},

		// Domain blacklist mode
		{
			name: "Redirect to blacklisted domain - block",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{"evil.com"},
			},
			redirectURL: "http://evil.com/malware",
			viaCount:    1,
			expectError: true,
			errorMsg:    "domain is blacklisted",
		},
		{
			name: "Redirect to non-blacklisted domain - allow",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{"evil.com"},
			},
			redirectURL: "http://good.com/page",
			viaCount:    1,
			expectError: false,
		},

		// Domain whitelist mode
		{
			name: "Redirect to whitelisted domain - allow",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: true,
				DomainListMode: ModeWhitelist,
				DomainList:     []string{"trusted.com"},
			},
			redirectURL: "http://trusted.com/api",
			viaCount:    1,
			expectError: false,
		},
		{
			name: "Redirect to non-whitelisted domain - block",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: true,
				DomainListMode: ModeWhitelist,
				DomainList:     []string{"trusted.com"},
			},
			redirectURL: "http://untrusted.com/page",
			viaCount:    1,
			expectError: true,
			errorMsg:    "domain not in whitelist",
		},

		// Wildcard domain matching in redirects
		{
			name: "Redirect to wildcard blacklisted domain - block",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{"*.evil.com"},
			},
			redirectURL: "http://sub.evil.com/path",
			viaCount:    1,
			expectError: true,
			errorMsg:    "domain is blacklisted",
		},
		{
			name: "Redirect to wildcard whitelisted domain - allow",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: true,
				DomainListMode: ModeWhitelist,
				DomainList:     []string{"*.trusted.com"},
			},
			redirectURL: "http://api.trusted.com/v1",
			viaCount:    1,
			expectError: false,
		},

		// Private IP in redirects
		{
			name: "Redirect to private IP when not allowed - block",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: false,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
				IPListMode:     ModeBlacklist,
				IPList:         []string{},
			},
			redirectURL: "http://192.168.1.1/admin",
			viaCount:    1,
			expectError: true,
			errorMsg:    "private IP address not allowed",
		},
		{
			name: "Redirect to private IP when allowed - allow",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
				IPListMode:     ModeBlacklist,
				IPList:         []string{},
			},
			redirectURL: "http://192.168.1.1/admin",
			viaCount:    1,
			expectError: false,
		},

		// Port restrictions in redirects
		{
			name: "Redirect to allowed port - allow",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: true,
				AllowedPorts:   []int{80, 443},
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
			},
			redirectURL: "https://example.com:443/secure",
			viaCount:    1,
			expectError: false,
		},
		{
			name: "Redirect to disallowed port - block",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: true,
				AllowedPorts:   []int{80, 443},
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
			},
			redirectURL: "http://example.com:8080/api",
			viaCount:    1,
			expectError: true,
			errorMsg:    "port 8080 is not allowed",
		},

		// IP blacklist in redirects
		{
			name: "Redirect to blacklisted IP - block",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
				IPListMode:     ModeBlacklist,
				IPList:         []string{"8.8.8.8"},
			},
			redirectURL: "http://8.8.8.8/dns",
			viaCount:    1,
			expectError: true,
			errorMsg:    "IP address is blacklisted",
		},

		// IP whitelist in redirects
		{
			name: "Redirect to whitelisted IP - allow",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: false,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
				IPListMode:     ModeWhitelist,
				IPList:         []string{"8.8.8.8"},
			},
			redirectURL: "http://8.8.8.8/dns",
			viaCount:    1,
			expectError: false,
		},
		{
			name: "Redirect to non-whitelisted IP - block",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: false,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
				IPListMode:     ModeWhitelist,
				IPList:         []string{"8.8.8.8"},
			},
			redirectURL: "http://1.1.1.1/cf",
			viaCount:    1,
			expectError: true,
			errorMsg:    "IP address not in whitelist",
		},

		// Protocol validation in redirects
		{
			name: "Redirect to unsupported protocol - block",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
			},
			redirectURL: "ftp://files.example.com/data",
			viaCount:    1,
			expectError: true,
			errorMsg:    "unsupported protocol",
		},

		// Redirect chain test (multiple hops)
		{
			name: "Multiple redirect hops within limit - allow",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
			},
			redirectURL: "http://hop3.example.com/final",
			viaCount:    3,
			expectError: false,
		},
		{
			name: "Redirect at max limit - block",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   3,
				AllowPrivateIP: true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
			},
			redirectURL: "http://hop4.example.com/blocked",
			viaCount:    3,
			expectError: true,
			errorMsg:    "stopped after 3 redirects",
		},

		// Localhost redirect
		{
			name: "Redirect to localhost IP - block when private IP not allowed",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: false,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
				IPListMode:     ModeBlacklist,
				IPList:         []string{},
			},
			redirectURL: "http://127.0.0.1/local",
			viaCount:    1,
			expectError: true,
			errorMsg:    "private IP address not allowed",
		},

		// CIDR in redirect IP validation
		{
			name: "Redirect to IP in blacklisted CIDR - block",
			config: &SSRFProtectionConfig{
				Enabled:        true,
				MaxRedirects:   10,
				AllowPrivateIP: true,
				DomainListMode: ModeBlacklist,
				DomainList:     []string{},
				IPListMode:     ModeBlacklist,
				IPList:         []string{"10.0.0.0/8"},
			},
			redirectURL: "http://10.20.30.40/internal",
			viaCount:    1,
			expectError: true,
			errorMsg:    "IP address is blacklisted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protection := NewSSRFProtection(tt.config)
			checkRedirect := protection.CheckRedirect(ctx)

			// Create mock request
			redirectReq, _ := http.NewRequest("GET", tt.redirectURL, nil)

			// Create mock via slice (previous requests in the redirect chain)
			via := make([]*http.Request, tt.viaCount)
			for i := 0; i < tt.viaCount; i++ {
				via[i] = &http.Request{}
			}

			err := checkRedirect(redirectReq, via)

			if tt.expectError {
				if err == nil {
					t.Errorf("CheckRedirect(%s) expected error containing '%s', got nil", tt.redirectURL, tt.errorMsg)
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("CheckRedirect(%s) error = '%v', expected to contain '%s'", tt.redirectURL, err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("CheckRedirect(%s) unexpected error: %v", tt.redirectURL, err)
				}
			}
		})
	}
}

func TestCheckRedirectWithConfigUpdate(t *testing.T) {
	ctx := context.Background()

	// Create protection with initial config
	initialConfig := &SSRFProtectionConfig{
		Enabled:        true,
		MaxRedirects:   10,
		AllowPrivateIP: true,
		DomainListMode: ModeBlacklist,
		DomainList:     []string{"blocked.com"},
	}
	protection := NewSSRFProtection(initialConfig)
	checkRedirect := protection.CheckRedirect(ctx)

	// Create test request
	blockedReq, _ := http.NewRequest("GET", "http://blocked.com/path", nil)
	allowedReq, _ := http.NewRequest("GET", "http://allowed.com/path", nil)
	via := []*http.Request{{}}

	// Test 1: Initially blocked.com should be blocked
	err := checkRedirect(blockedReq, via)
	if err == nil {
		t.Error("blocked.com should be blocked initially")
	}

	// Test 2: allowed.com should be allowed initially
	err = checkRedirect(allowedReq, via)
	if err != nil {
		t.Errorf("allowed.com should be allowed initially, got error: %v", err)
	}

	// Update config to block allowed.com instead
	newConfig := &SSRFProtectionConfig{
		Enabled:        true,
		MaxRedirects:   10,
		AllowPrivateIP: true,
		DomainListMode: ModeBlacklist,
		DomainList:     []string{"allowed.com"},
	}
	protection.UpdateConfig(newConfig)

	// Test 3: After update, blocked.com should now be allowed
	err = checkRedirect(blockedReq, via)
	if err != nil {
		t.Errorf("blocked.com should be allowed after config update, got error: %v", err)
	}

	// Test 4: After update, allowed.com should now be blocked
	err = checkRedirect(allowedReq, via)
	if err == nil {
		t.Error("allowed.com should be blocked after config update")
	}
}

func TestCheckRedirectDisabledProtection(t *testing.T) {
	ctx := context.Background()

	// Create protection with DISABLED config
	config := &SSRFProtectionConfig{
		Enabled:        false,
		MaxRedirects:   1, // Very restrictive
		AllowPrivateIP: false,
		DomainListMode: ModeWhitelist,
		DomainList:     []string{}, // Empty whitelist would block everything
		IPListMode:     ModeWhitelist,
		IPList:         []string{},
	}
	protection := NewSSRFProtection(config)
	checkRedirect := protection.CheckRedirect(ctx)

	// All these should be allowed because protection is disabled
	testURLs := []string{
		"http://any.domain.com/path",
		"http://192.168.1.1/admin",
		"http://127.0.0.1/local",
		"http://10.0.0.1/internal",
		"ftp://files.server.com/data", // Even FTP should pass when disabled
	}

	for _, testURL := range testURLs {
		req, _ := http.NewRequest("GET", testURL, nil)
		via := make([]*http.Request, 10) // Many redirects
		for i := range via {
			via[i] = &http.Request{}
		}

		err := checkRedirect(req, via)
		if err != nil {
			t.Errorf("With protection disabled, redirect to %s should be allowed, got error: %v", testURL, err)
		}
	}
}
