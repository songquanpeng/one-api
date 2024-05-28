package model

import (
	"one-api/common/logger"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func removeKeyIndexMigration() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202405152141",
		Migrate: func(tx *gorm.DB) error {
			dialect := tx.Dialector.Name()
			if dialect == "sqlite" {
				return nil
			}

			if !tx.Migrator().HasIndex(&Channel{}, "idx_channels_key") {
				return nil
			}

			err := tx.Migrator().DropIndex(&Channel{}, "idx_channels_key")
			if err != nil {
				logger.SysLog("remove idx_channels_key  Failure: " + err.Error())
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	}
}

func migration(db *gorm.DB) error {
	// 如果是第一次运行 直接跳过
	if !db.Migrator().HasTable("channels") {
		return nil
	}

	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		removeKeyIndexMigration(),
	})
	return m.Migrate()
}
