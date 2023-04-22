package common

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	Port         = flag.Int("port", 3000, "the listening port")
	PrintVersion = flag.Bool("version", false, "print version and exit")
	LogDir       = flag.String("log-dir", "", "specify the log directory")
	//Host         = flag.String("host", "localhost", "the server's ip address or domain")
	//Path         = flag.String("path", "", "specify a local path to public")
	//VideoPath    = flag.String("video", "", "specify a video folder to public")
	//NoBrowser    = flag.Bool("no-browser", false, "open browser or not")
)

// UploadPath Maybe override by ENV_VAR
var UploadPath = "upload"

//var ExplorerRootPath = UploadPath
//var ImageUploadPath = "upload/images"
//var VideoServePath = "upload"

func init() {
	flag.Parse()

	if *PrintVersion {
		fmt.Println(Version)
		os.Exit(0)
	}

	if os.Getenv("SESSION_SECRET") != "" {
		SessionSecret = os.Getenv("SESSION_SECRET")
	}
	if os.Getenv("SQLITE_PATH") != "" {
		SQLitePath = os.Getenv("SQLITE_PATH")
	}
	if os.Getenv("UPLOAD_PATH") != "" {
		UploadPath = os.Getenv("UPLOAD_PATH")
		//ExplorerRootPath = UploadPath
		//ImageUploadPath = path.Join(UploadPath, "images")
		//VideoServePath = UploadPath
	}
	if *LogDir != "" {
		var err error
		*LogDir, err = filepath.Abs(*LogDir)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := os.Stat(*LogDir); os.IsNotExist(err) {
			err = os.Mkdir(*LogDir, 0777)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	//if *Path != "" {
	//	ExplorerRootPath = *Path
	//}
	//if *VideoPath != "" {
	//	VideoServePath = *VideoPath
	//}
	//
	//ExplorerRootPath, _ = filepath.Abs(ExplorerRootPath)
	//VideoServePath, _ = filepath.Abs(VideoServePath)
	//ImageUploadPath, _ = filepath.Abs(ImageUploadPath)
	//
	if _, err := os.Stat(UploadPath); os.IsNotExist(err) {
		_ = os.Mkdir(UploadPath, 0777)
	}
	//if _, err := os.Stat(ImageUploadPath); os.IsNotExist(err) {
	//	_ = os.Mkdir(ImageUploadPath, 0777)
	//}
}
