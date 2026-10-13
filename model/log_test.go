package model

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestLogDB 为每个测试创建独立的内存 SQLite（LOG_DB 及依赖的主 DB/User 表）。
func setupTestLogDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&Log{}, &User{}); err != nil {
		t.Fatalf("automigrate failed: %v", err)
	}
	oldLogDB, oldDB := LOG_DB, DB
	LOG_DB, DB = db, db
	t.Cleanup(func() {
		LOG_DB, DB = oldLogDB, oldDB
		if sqlDB, _ := db.DB(); sqlDB != nil {
			_ = sqlDB.Close()
		}
	})
}

func TestGetSuccessRateStatistics(t *testing.T) {
	setupTestLogDB(t)

	// 构造日志：gpt-4o 成功 3 次、失败 1 次；gpt-4o-mini 成功 1 次；
	// claude-3 成功 2 次；其余类型（充值/测试）不应计入
	seed := []Log{
		{UserId: 1, ModelName: "gpt-4o", Type: LogTypeConsume},
		{UserId: 1, ModelName: "gpt-4o", Type: LogTypeConsume},
		{UserId: 1, ModelName: "gpt-4o", Type: LogTypeConsume},
		{UserId: 1, ModelName: "gpt-4o", Type: LogTypeError},
		{UserId: 1, ModelName: "gpt-4o-mini", Type: LogTypeConsume},
		{UserId: 1, ModelName: "claude-3", Type: LogTypeConsume},
		{UserId: 1, ModelName: "claude-3", Type: LogTypeConsume},
		{UserId: 1, ModelName: "claude-3", Type: LogTypeTopup}, // 充值日志不计入
		{UserId: 1, ModelName: "gpt-4o", Type: LogTypeTest},   // 测试日志不计入
		{UserId: 2, ModelName: "gpt-4o", Type: LogTypeConsume}, // 其他用户（管理员视角计入）
	}
	for i := range seed {
		seed[i].CreatedAt = int64(1700000000 + i)
		if err := LOG_DB.Create(&seed[i]).Error; err != nil {
			t.Fatalf("seed log failed: %v", err)
		}
	}

	stats, err := GetSuccessRateStatistics(0, 0, 0, "", "", "", 0)
	if err != nil {
		t.Fatalf("GetSuccessRateStatistics failed: %v", err)
	}
	if len(stats) != 3 {
		t.Fatalf("expected 3 models, got %d: %+v", len(stats), stats)
	}
	// 按调用总量降序：gpt-4o (3+1+1=5), claude-3 (2), gpt-4o-mini (1)
	if stats[0].ModelName != "gpt-4o" || stats[0].SuccessCount != 4 || stats[0].ErrorCount != 1 {
		t.Errorf("gpt-4o stat mismatch: %+v", stats[0])
	}
	if stats[1].ModelName != "claude-3" || stats[1].SuccessCount != 2 || stats[1].ErrorCount != 0 {
		t.Errorf("claude-3 stat mismatch: %+v", stats[1])
	}
	if stats[2].ModelName != "gpt-4o-mini" || stats[2].SuccessCount != 1 {
		t.Errorf("gpt-4o-mini stat mismatch: %+v", stats[2])
	}

	// 用户视角：user 1 的 gpt-4o 应为 3 成功 1 失败
	userStats, err := GetSuccessRateStatistics(1, 0, 0, "gpt-4o", "", "", 0)
	if err != nil {
		t.Fatalf("GetSuccessRateStatistics(user) failed: %v", err)
	}
	if len(userStats) != 1 || userStats[0].SuccessCount != 3 || userStats[0].ErrorCount != 1 {
		t.Errorf("user stat mismatch: %+v", userStats)
	}
}

func TestRecordErrorLog(t *testing.T) {
	setupTestLogDB(t)

	RecordErrorLog(context.Background(), &Log{
		UserId:      7,
		ModelName:   "gpt-4o",
		Content:     "do_request_failed",
		ChannelId:   3,
		TokenName:   "tk",
		IsStream:    true,
		ElapsedTime: 42,
	})

	var logs []Log
	if err := LOG_DB.Find(&logs).Error; err != nil {
		t.Fatalf("find logs failed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].Type != LogTypeError {
		t.Errorf("expected type %d, got %d", LogTypeError, logs[0].Type)
	}
	if logs[0].Username != GetUsernameById(7) {
		t.Errorf("username mismatch: %s", logs[0].Username)
	}
	if logs[0].CreatedAt == 0 {
		t.Error("created_at should be set")
	}
}
