package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
)

func setupChannelCacheTestDB(t *testing.T) {
	t.Helper()

	oldSQLitePath := common.SQLitePath
	oldUsingSQLite := common.UsingSQLite
	oldUsingMySQL := common.UsingMySQL
	oldUsingPostgreSQL := common.UsingPostgreSQL
	oldMemoryCacheEnabled := config.MemoryCacheEnabled
	oldDB := DB
	oldLogDB := LOG_DB
	oldGroup2Model2Channels := group2model2channels

	common.SQLitePath = filepath.Join(t.TempDir(), "one-api-test.db")
	common.UsingSQLite = false
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	config.MemoryCacheEnabled = true
	group2model2channels = nil

	t.Setenv("SQL_DSN", "")
	InitDB()
	LOG_DB = DB

	t.Cleanup(func() {
		if DB != nil {
			_ = closeDB(DB)
		}
		DB = oldDB
		LOG_DB = oldLogDB
		common.SQLitePath = oldSQLitePath
		common.UsingSQLite = oldUsingSQLite
		common.UsingMySQL = oldUsingMySQL
		common.UsingPostgreSQL = oldUsingPostgreSQL
		config.MemoryCacheEnabled = oldMemoryCacheEnabled
		group2model2channels = oldGroup2Model2Channels
		_ = os.Remove(common.SQLitePath)
	})
}

func TestBatchInsertChannelsRefreshesMemoryChannelCache(t *testing.T) {
	setupChannelCacheTestDB(t)

	InitChannelCache()

	err := BatchInsertChannels([]Channel{{
		Name:   "qwen",
		Type:   1,
		Key:    "test-key",
		Status: ChannelStatusEnabled,
		Models: "qwen3.6-plus",
		Group:  "default",
	}})
	if err != nil {
		t.Fatalf("BatchInsertChannels returned error: %v", err)
	}

	_, err = CacheGetRandomSatisfiedChannel("default", "qwen3.6-plus", false)
	if err != nil {
		t.Fatalf("expected inserted channel to be immediately available from memory cache, got error: %v", err)
	}
}

func TestChannelUpdateRefreshesMemoryChannelCache(t *testing.T) {
	setupChannelCacheTestDB(t)

	channel := Channel{
		Name:   "qwen",
		Type:   1,
		Key:    "test-key",
		Status: ChannelStatusEnabled,
		Models: "qwen3.6-plus",
		Group:  "default",
	}
	if err := channel.Insert(); err != nil {
		t.Fatalf("Insert returned error: %v", err)
	}
	InitChannelCache()

	channel.Models = "qwen3-coder-next"
	if err := channel.Update(); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	_, err := CacheGetRandomSatisfiedChannel("default", "qwen3-coder-next", false)
	if err != nil {
		t.Fatalf("expected updated model to be immediately available from memory cache, got error: %v", err)
	}

	_, err = CacheGetRandomSatisfiedChannel("default", "qwen3.6-plus", false)
	if err == nil {
		t.Fatalf("expected removed model to disappear from memory cache after update")
	}
}

func TestChannelDeleteRefreshesMemoryChannelCache(t *testing.T) {
	setupChannelCacheTestDB(t)

	channel := Channel{
		Name:   "qwen",
		Type:   1,
		Key:    "test-key",
		Status: ChannelStatusEnabled,
		Models: "qwen3.6-plus",
		Group:  "default",
	}
	if err := channel.Insert(); err != nil {
		t.Fatalf("Insert returned error: %v", err)
	}
	InitChannelCache()

	if err := channel.Delete(); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	_, err := CacheGetRandomSatisfiedChannel("default", "qwen3.6-plus", false)
	if err == nil {
		t.Fatalf("expected deleted channel to be removed from memory cache")
	}
}
