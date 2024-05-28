package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"one-api/common/utils"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const (
	loggerINFO  = "INFO"
	loggerWarn  = "WARN"
	loggerError = "ERR"
)
const (
	RequestIdKey = "X-Oneapi-Request-Id"
)

const maxLogCount = 1000000

var logCount int
var setupLogLock sync.Mutex
var setupLogWorking bool

var defaultLogDir = "./logs"

func SetupLogger() {
	logDir := getLogDir()
	if logDir == "" {
		return
	}

	ok := setupLogLock.TryLock()
	if !ok {
		log.Println("setup log is already working")
		return
	}
	defer func() {
		setupLogLock.Unlock()
		setupLogWorking = false
	}()
	logPath := filepath.Join(logDir, fmt.Sprintf("oneapi-%s.log", time.Now().Format("20060102")))
	fd, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("failed to open log file")
	}
	gin.DefaultWriter = io.MultiWriter(os.Stdout, fd)
	gin.DefaultErrorWriter = io.MultiWriter(os.Stderr, fd)
}

func getLogDir() string {
	logDir := viper.GetString("log_dir")
	if logDir == "" {
		logDir = defaultLogDir
	}

	var err error
	logDir, err = filepath.Abs(logDir)
	if err != nil {
		log.Fatal(err)
		return ""
	}

	if !utils.IsFileExist(logDir) {
		err = os.Mkdir(logDir, 0777)
		if err != nil {
			log.Fatal(err)
			return ""
		}
	}

	return logDir
}

func SysLog(s string) {
	t := time.Now()
	_, _ = fmt.Fprintf(gin.DefaultWriter, "[SYS] %v | %s \n", t.Format("2006/01/02 - 15:04:05"), s)
}

func SysError(s string) {
	t := time.Now()
	_, _ = fmt.Fprintf(gin.DefaultErrorWriter, "[SYS] %v | %s \n", t.Format("2006/01/02 - 15:04:05"), s)
}

func LogInfo(ctx context.Context, msg string) {
	logHelper(ctx, loggerINFO, msg)
}

func LogWarn(ctx context.Context, msg string) {
	logHelper(ctx, loggerWarn, msg)
}

func LogError(ctx context.Context, msg string) {
	logHelper(ctx, loggerError, msg)
}

func logHelper(ctx context.Context, level string, msg string) {
	writer := gin.DefaultErrorWriter
	if level == loggerINFO {
		writer = gin.DefaultWriter
	}
	id, ok := ctx.Value(RequestIdKey).(string)
	if !ok {
		id = "unknown"
	}
	now := time.Now()
	_, _ = fmt.Fprintf(writer, "[%s] %v | %s | %s \n", level, now.Format("2006/01/02 - 15:04:05"), id, msg)
	logCount++ // we don't need accurate count, so no lock here
	if logCount > maxLogCount && !setupLogWorking {
		logCount = 0
		setupLogWorking = true
		go func() {
			SetupLogger()
		}()
	}
}

func FatalLog(v ...any) {
	t := time.Now()
	_, _ = fmt.Fprintf(gin.DefaultErrorWriter, "[FATAL] %v | %v \n", t.Format("2006/01/02 - 15:04:05"), v)
	os.Exit(1)
}
