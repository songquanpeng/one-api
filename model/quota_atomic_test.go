package model

import (
	"testing"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupQuotaAtomicTest(t *testing.T) *gorm.DB {
	t.Helper()
	previousDB := DB
	previousBatchUpdate := config.BatchUpdateEnabled
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}, &Token{}))
	DB = db
	config.BatchUpdateEnabled = false
	t.Cleanup(func() {
		DB = previousDB
		config.BatchUpdateEnabled = previousBatchUpdate
	})
	return db
}

func seedQuotaAtomicTestData(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Create(&User{
		Id:       1,
		Username: "quota-user",
		Password: "password123",
		Quota:    5,
		Status:   UserStatusEnabled,
	}).Error)
	require.NoError(t, db.Create(&Token{
		Id:             1,
		UserId:         1,
		Key:            "quota-token",
		Status:         TokenStatusEnabled,
		RemainQuota:    10,
		UnlimitedQuota: false,
	}).Error)
}

func TestDecreaseQuotaRemainsAtomicWhenBatchUpdatesAreEnabled(t *testing.T) {
	db := setupQuotaAtomicTest(t)
	seedQuotaAtomicTestData(t, db)
	config.BatchUpdateEnabled = true

	require.NoError(t, DecreaseTokenQuota(1, 3))
	require.NoError(t, DecreaseUserQuota(1, 3))

	var token Token
	var user User
	require.NoError(t, db.First(&token, 1).Error)
	require.NoError(t, db.First(&user, 1).Error)
	require.Equal(t, int64(7), token.RemainQuota)
	require.Equal(t, int64(2), user.Quota)
}

func TestPreConsumeQuotaRollsBackTokenWhenUserQuotaIsInsufficient(t *testing.T) {
	db := setupQuotaAtomicTest(t)
	seedQuotaAtomicTestData(t, db)

	err := PreConsumeTokenQuota(1, 6)
	require.Error(t, err)
	require.ErrorContains(t, err, "insufficient")

	var token Token
	var user User
	require.NoError(t, db.First(&token, 1).Error)
	require.NoError(t, db.First(&user, 1).Error)
	require.Equal(t, int64(10), token.RemainQuota)
	require.Equal(t, int64(5), user.Quota)
}
