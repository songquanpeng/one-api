package model

import (
	"fmt"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/logger"
	"one-api/common/utils"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func SetupDB() {
	err := InitDB()
	if err != nil {
		logger.FatalLog("failed to initialize database: " + err.Error())
	}
	ChannelGroup.Load()
	config.RootUserEmail = GetRootUserEmail()

	if viper.GetBool("batch_update_enabled") {
		config.BatchUpdateEnabled = true
		config.BatchUpdateInterval = utils.GetOrDefault("batch_update_interval", 5)
		logger.SysLog("batch update enabled with interval " + strconv.Itoa(config.BatchUpdateInterval) + "s")
		InitBatchUpdater()
	}
}

func createRootAccountIfNeed() error {
	var user User
	//if user.Status != common.UserStatusEnabled {
	if err := DB.First(&user).Error; err != nil {
		logger.SysLog("no user exists, create a root user for you: username is root, password is 123456")
		hashedPassword, err := common.Password2Hash("123456")
		if err != nil {
			return err
		}
		rootUser := User{
			Username:    "root",
			Password:    hashedPassword,
			Role:        config.RoleRootUser,
			Status:      config.UserStatusEnabled,
			DisplayName: "Root User",
			AccessToken: utils.GetUUID(),
			Quota:       100000000,
		}
		DB.Create(&rootUser)
	}
	return nil
}

func chooseDB() (*gorm.DB, error) {
	if viper.IsSet("sql_dsn") {
		dsn := viper.GetString("sql_dsn")
		if strings.HasPrefix(dsn, "postgres://") {
			// Use PostgreSQL
			logger.SysLog("using PostgreSQL as database")
			common.UsingPostgreSQL = true
			return gorm.Open(postgres.New(postgres.Config{
				DSN:                  dsn,
				PreferSimpleProtocol: true, // disables implicit prepared statement usage
			}), &gorm.Config{
				PrepareStmt: true, // precompile SQL
			})
		}
		// Use MySQL
		logger.SysLog("using MySQL as database")
		return gorm.Open(mysql.Open(dsn), &gorm.Config{
			PrepareStmt: true, // precompile SQL
		})
	}
	// Use SQLite
	logger.SysLog("SQL_DSN not set, using SQLite as database")
	common.UsingSQLite = true
	config := fmt.Sprintf("?_busy_timeout=%d", utils.GetOrDefault("sqlite_busy_timeout", 3000))
	return gorm.Open(sqlite.Open(viper.GetString("sqlite_path")+config), &gorm.Config{
		PrepareStmt: true, // precompile SQL
	})
}

func InitDB() (err error) {
	db, err := chooseDB()
	if err == nil {
		if viper.GetBool("debug") {
			db = db.Debug()
		}
		DB = db
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}

		sqlDB.SetMaxIdleConns(utils.GetOrDefault("SQL_MAX_IDLE_CONNS", 100))
		sqlDB.SetMaxOpenConns(utils.GetOrDefault("SQL_MAX_OPEN_CONNS", 1000))
		sqlDB.SetConnMaxLifetime(time.Second * time.Duration(utils.GetOrDefault("SQL_MAX_LIFETIME", 60)))

		if !config.IsMasterNode {
			return nil
		}
		logger.SysLog("database migration started")

		migration(DB)

		err = db.AutoMigrate(&Channel{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Token{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&User{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Option{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Redemption{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Ability{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Log{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&TelegramMenu{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Price{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Midjourney{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&ChatCache{})
		if err != nil {
			return err
		}
		logger.SysLog("database migrated")
		err = createRootAccountIfNeed()
		return err
	} else {
		logger.FatalLog(err)
	}
	return err
}

// func MigrateDB(db *gorm.DB) error {
// 	if DB.Migrator().HasConstraint(&Price{}, "model") {
// 		fmt.Println("----Price model has constraint----")
// 		// 如果是主键，移除主键约束
// 		err := db.Migrator().DropConstraint(&Price{}, "model")
// 		if err != nil {
// 			return err
// 		}
// 		// 修改字段长度
// 		err = db.Migrator().AlterColumn(&Price{}, "model")
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

func CloseDB() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	return err
}
