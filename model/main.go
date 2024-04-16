package model

import (
	"fmt"
	"one-api/common"
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
		common.FatalLog("failed to initialize database: " + err.Error())
	}
	ChannelGroup.Load()
	common.RootUserEmail = GetRootUserEmail()

	if viper.GetBool("BATCH_UPDATE_ENABLED") {
		common.BatchUpdateEnabled = true
		common.BatchUpdateInterval = common.GetOrDefault("BATCH_UPDATE_INTERVAL", 5)
		common.SysLog("batch update enabled with interval " + strconv.Itoa(common.BatchUpdateInterval) + "s")
		InitBatchUpdater()
	}
}

func createRootAccountIfNeed() error {
	var user User
	//if user.Status != common.UserStatusEnabled {
	if err := DB.First(&user).Error; err != nil {
		common.SysLog("no user exists, create a root user for you: username is root, password is 123456")
		hashedPassword, err := common.Password2Hash("123456")
		if err != nil {
			return err
		}
		rootUser := User{
			Username:    "root",
			Password:    hashedPassword,
			Role:        common.RoleRootUser,
			Status:      common.UserStatusEnabled,
			DisplayName: "Root User",
			AccessToken: common.GetUUID(),
			Quota:       100000000,
		}
		DB.Create(&rootUser)
	}
	return nil
}

func chooseDB() (*gorm.DB, error) {
	if viper.IsSet("SQL_DSN") {
		dsn := viper.GetString("SQL_DSN")
		if strings.HasPrefix(dsn, "postgres://") {
			// Use PostgreSQL
			common.SysLog("using PostgreSQL as database")
			common.UsingPostgreSQL = true
			return gorm.Open(postgres.New(postgres.Config{
				DSN:                  dsn,
				PreferSimpleProtocol: true, // disables implicit prepared statement usage
			}), &gorm.Config{
				PrepareStmt: true, // precompile SQL
			})
		}
		// Use MySQL
		common.SysLog("using MySQL as database")
		return gorm.Open(mysql.Open(dsn), &gorm.Config{
			PrepareStmt: true, // precompile SQL
		})
	}
	// Use SQLite
	common.SysLog("SQL_DSN not set, using SQLite as database")
	common.UsingSQLite = true
	config := fmt.Sprintf("?_busy_timeout=%d", common.GetOrDefault("SQLITE_BUSY_TIMEOUT", 3000))
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

		sqlDB.SetMaxIdleConns(common.GetOrDefault("SQL_MAX_IDLE_CONNS", 100))
		sqlDB.SetMaxOpenConns(common.GetOrDefault("SQL_MAX_OPEN_CONNS", 1000))
		sqlDB.SetConnMaxLifetime(time.Second * time.Duration(common.GetOrDefault("SQL_MAX_LIFETIME", 60)))

		if !common.IsMasterNode {
			return nil
		}
		common.SysLog("database migration started")
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
		common.SysLog("database migrated")
		err = createRootAccountIfNeed()
		return err
	} else {
		common.FatalLog(err)
	}
	return err
}

func CloseDB() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	return err
}
