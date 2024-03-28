package config

import (
	"strings"
	"time"

	"one-api/cli"
	"one-api/common"

	"github.com/spf13/viper"
)

func InitConf() {
	cli.FlagConfig()
	defaultConfig()
	setConfigFile()
	setEnv()

	if viper.GetBool("debug") {
		common.SysLog("running in debug mode")
	}

	common.IsMasterNode = viper.GetString("NODE_TYPE") != "slave"
	common.RequestInterval = time.Duration(viper.GetInt("POLLING_INTERVAL")) * time.Second
	common.SessionSecret = common.GetOrDefault("SESSION_SECRET", common.SessionSecret)
}

func setConfigFile() {
	if !common.IsFileExist(*cli.Config) {
		return
	}

	viper.SetConfigFile(*cli.Config)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func setEnv() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func defaultConfig() {
	viper.SetDefault("port", "3000")
	viper.SetDefault("gin_mode", "release")
	viper.SetDefault("log_dir", "./logs")
	viper.SetDefault("sqlite_path", "one-api.db")
	viper.SetDefault("sqlite_busy_timeout", 3000)
	viper.SetDefault("sync_frequency", 600)
	viper.SetDefault("batch_update_interval", 5)
	viper.SetDefault("global.api_rate_limit", 180)
	viper.SetDefault("global.web_rate_limit", 100)
	viper.SetDefault("connect_timeout", 5)
	viper.SetDefault("auto_price_updates", true)
}
