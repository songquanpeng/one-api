package model

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var quotaTestSeq atomic.Int64

func setupQuotaTestDB(t *testing.T) {
	t.Helper()
	dsn := fmt.Sprintf("file:quota-%s-%d?mode=memory&cache=shared&_busy_timeout=5000", t.Name(), quotaTestSeq.Add(1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&User{}, &Token{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	DB = db
}

func createTestUser(t *testing.T, quota int64) *User {
	t.Helper()
	seq := quotaTestSeq.Add(1)
	user := &User{
		Username:    fmt.Sprintf("u%d", seq),
		Password:    "password1234",
		Quota:       quota,
		Status:      UserStatusEnabled,
		Role:        RoleCommonUser,
		AccessToken: fmt.Sprintf("%032d", seq),
		AffCode:     fmt.Sprintf("a%d", seq),
	}
	if err := DB.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func createTestToken(t *testing.T, userId int, remain int64) *Token {
	t.Helper()
	seq := quotaTestSeq.Add(1)
	token := &Token{
		UserId:      userId,
		Key:         fmt.Sprintf("%048d", seq),
		Name:        "quota-test",
		Status:      TokenStatusEnabled,
		RemainQuota: remain,
		ExpiredTime: -1,
	}
	if err := DB.Create(token).Error; err != nil {
		t.Fatalf("create token: %v", err)
	}
	return token
}

func TestDecreaseUserQuotaSequentialOversellDoesNotGoNegative(t *testing.T) {
	setupQuotaTestDB(t)
	user := createTestUser(t, 100)

	if err := decreaseUserQuota(user.Id, 60); err != nil {
		t.Fatalf("first decrease: %v", err)
	}
	err := decreaseUserQuota(user.Id, 60)
	remaining, qErr := GetUserQuota(user.Id)
	if qErr != nil {
		t.Fatalf("GetUserQuota: %v", qErr)
	}
	if remaining < 0 {
		t.Fatalf("user quota went negative: %d", remaining)
	}
	if err == nil {
		t.Fatalf("expected oversell to fail, remaining=%d", remaining)
	}
	if remaining != 40 {
		t.Fatalf("remaining quota = %d, want 40", remaining)
	}
}

func TestDecreaseTokenQuotaSequentialOversellDoesNotGoNegative(t *testing.T) {
	setupQuotaTestDB(t)
	user := createTestUser(t, 1000)
	token := createTestToken(t, user.Id, 100)

	if err := decreaseTokenQuota(token.Id, 60); err != nil {
		t.Fatalf("first decrease: %v", err)
	}
	err := decreaseTokenQuota(token.Id, 60)
	got, gErr := GetTokenById(token.Id)
	if gErr != nil {
		t.Fatalf("GetTokenById: %v", gErr)
	}
	if got.RemainQuota < 0 {
		t.Fatalf("token remain_quota went negative: %d", got.RemainQuota)
	}
	if err == nil {
		t.Fatalf("expected oversell to fail, remain_quota=%d", got.RemainQuota)
	}
	if got.RemainQuota != 40 {
		t.Fatalf("remain_quota = %d, want 40", got.RemainQuota)
	}
	if got.UsedQuota != 60 {
		t.Fatalf("used_quota = %d, want 60", got.UsedQuota)
	}
}

func TestDecreaseUserQuotaExactDrainThenReject(t *testing.T) {
	setupQuotaTestDB(t)
	user := createTestUser(t, 100)

	if err := decreaseUserQuota(user.Id, 100); err != nil {
		t.Fatalf("exact drain: %v", err)
	}
	err := decreaseUserQuota(user.Id, 1)
	remaining, qErr := GetUserQuota(user.Id)
	if qErr != nil {
		t.Fatalf("GetUserQuota: %v", qErr)
	}
	if remaining != 0 {
		t.Fatalf("remaining quota = %d, want 0", remaining)
	}
	if err == nil {
		t.Fatal("expected decrease below zero to fail")
	}
}

func TestDecreaseUserQuotaConcurrentOversellDoesNotGoNegative(t *testing.T) {
	setupQuotaTestDB(t)
	user := createTestUser(t, 100)

	const workers = 10
	const each = int64(60)
	var success atomic.Int64
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if err := decreaseUserQuota(user.Id, each); err == nil {
				success.Add(1)
			}
		}()
	}
	wg.Wait()

	remaining, err := GetUserQuota(user.Id)
	if err != nil {
		t.Fatalf("GetUserQuota: %v", err)
	}
	if remaining < 0 {
		t.Fatalf("user quota went negative under concurrent decrease: %d (successes=%d)", remaining, success.Load())
	}
	if success.Load() != 1 {
		t.Fatalf("successful decreases = %d, want 1; remaining=%d", success.Load(), remaining)
	}
	if remaining != 40 {
		t.Fatalf("remaining quota = %d, want 40", remaining)
	}
}

func TestDecreaseTokenQuotaConcurrentOversellDoesNotGoNegative(t *testing.T) {
	setupQuotaTestDB(t)
	user := createTestUser(t, 1000)
	token := createTestToken(t, user.Id, 100)

	const workers = 10
	const each = int64(60)
	var success atomic.Int64
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if err := decreaseTokenQuota(token.Id, each); err == nil {
				success.Add(1)
			}
		}()
	}
	wg.Wait()

	got, err := GetTokenById(token.Id)
	if err != nil {
		t.Fatalf("GetTokenById: %v", err)
	}
	if got.RemainQuota < 0 {
		t.Fatalf("token remain_quota went negative under concurrent decrease: %d (successes=%d)", got.RemainQuota, success.Load())
	}
	if success.Load() != 1 {
		t.Fatalf("successful decreases = %d, want 1; remain_quota=%d", success.Load(), got.RemainQuota)
	}
	if got.RemainQuota != 40 {
		t.Fatalf("remain_quota = %d, want 40", got.RemainQuota)
	}
	if got.UsedQuota != 60 {
		t.Fatalf("used_quota = %d, want 60", got.UsedQuota)
	}
}
