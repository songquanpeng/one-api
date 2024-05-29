package cli

import (
	"flag"
	"fmt"
	"one-api/common/config"
	"one-api/common/utils"
	"os"

	"github.com/spf13/viper"
)

var (
	port         = flag.Int("port", 0, "the listening port")
	printVersion = flag.Bool("version", false, "print version and exit")
	printHelp    = flag.Bool("help", false, "print help and exit")
	logDir       = flag.String("log-dir", "", "specify the log directory")
	Config       = flag.String("config", "config.yaml", "specify the config.yaml path")
	export       = flag.Bool("export", false, "Exports prices to a JSON file.")
)

func InitCli() {
	flag.Parse()

	if *printVersion {
		fmt.Println(config.Version)
		os.Exit(0)
	}

	if *printHelp {
		help()
		os.Exit(0)
	}

	if *port != 0 {
		viper.Set("port", *port)
	}

	if *logDir != "" {
		viper.Set("log_dir", *logDir)
	}

	if *export {
		ExportPrices()
		os.Exit(0)
	}

	if !utils.IsFileExist(*Config) {
		return
	}

	viper.SetConfigFile(*Config)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

}

func help() {
	fmt.Println("One API " + config.Version + " - All in one API service for OpenAI API.")
	fmt.Println("Copyright (C) 2024 MartialBE. All rights reserved.")
	fmt.Println("Original copyright holder: JustSong")
	fmt.Println("GitHub: https://github.com/MartialBE/one-api")
	fmt.Println("Usage: one-api [--port <port>] [--log-dir <log directory>] [--config <config.yaml path>] [--version] [--help]")
}
