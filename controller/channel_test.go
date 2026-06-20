package controller

import (
	"testing"

	"github.com/songquanpeng/one-api/model"
)

// TestChannelQuotaExhausted is a regression test for channels not auto-disabling
// at quota exhaustion. A channel must be considered exhausted as soon as
// used_quota >= quota (boundary inclusive), while a quota of 0 means unlimited.
func TestChannelQuotaExhausted(t *testing.T) {
	cases := []struct {
		name      string
		quota     int64
		usedQuota int64
		want      bool
	}{
		{name: "unlimited quota is never exhausted", quota: 0, usedQuota: 1_000_000, want: false},
		{name: "negative quota treated as unlimited", quota: -1, usedQuota: 5, want: false},
		{name: "usage below quota", quota: 100, usedQuota: 99, want: false},
		{name: "usage exactly at quota", quota: 100, usedQuota: 100, want: true},
		{name: "usage over quota", quota: 100, usedQuota: 150, want: true},
		{name: "no usage yet", quota: 100, usedQuota: 0, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			channel := &model.Channel{Quota: tc.quota, UsedQuota: tc.usedQuota}
			if got := channelQuotaExhausted(channel); got != tc.want {
				t.Errorf("channelQuotaExhausted(quota=%d, used=%d) = %v, want %v",
					tc.quota, tc.usedQuota, got, tc.want)
			}
		})
	}
}
