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

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
)

const (
	loggerDEBUG = "DEBUG"
	loggerINFO  = "INFO"
	loggerWarn  = "WARN"
	loggerError = "ERR"
)

var setupLogOnce sync.Once

func SetupLogger() {
	setupLogOnce.Do(func() {
		if LogDir != "" {
			logPath := filepath.Join(LogDir, fmt.Sprintf("oneapi-%s.log", time.Now().Format("20060102")))
			fd, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal("failed to open log file")
			}
			gin.DefaultWriter = io.MultiWriter(os.Stdout, fd)
			gin.DefaultErrorWriter = io.MultiWriter(os.Stderr, fd)
		}
	})
}

func SysLog(s string) {
	t := time.Now()
	_, _ = fmt.Fprintf(gin.DefaultWriter, "[SYS] %v | %s \n", t.Format("2006/01/02 - 15:04:05"), s)
}

func SysLogf(format string, a ...any) {
	SysLog(fmt.Sprintf(format, a...))
}

func SysError(s string) {
	t := time.Now()
	_, _ = fmt.Fprintf(gin.DefaultErrorWriter, "[SYS] %v | %s \n", t.Format("2006/01/02 - 15:04:05"), s)
}

func SysErrorf(format string, a ...any) {
	SysError(fmt.Sprintf(format, a...))
}

func Debug(ctx context.Context, msg string) {
	if config.DebugEnabled {
		logHelper(ctx, loggerDEBUG, msg)
	}
}

func Info(ctx context.Context, msg string) {
	logHelper(ctx, loggerINFO, msg)
}

func Warn(ctx context.Context, msg string) {
	logHelper(ctx, loggerWarn, msg)
}

func Error(ctx context.Context, msg string) {
	logHelper(ctx, loggerError, msg)
}

func Debugf(ctx context.Context, format string, a ...any) {
	Debug(ctx, fmt.Sprintf(format, a...))
}

func Infof(ctx context.Context, format string, a ...any) {
	Info(ctx, fmt.Sprintf(format, a...))
}

func Warnf(ctx context.Context, format string, a ...any) {
	Warn(ctx, fmt.Sprintf(format, a...))
}

func Errorf(ctx context.Context, format string, a ...any) {
	Error(ctx, fmt.Sprintf(format, a...))
}

func logHelper(ctx context.Context, level string, msg string) {
	writer := gin.DefaultErrorWriter
	if level == loggerINFO {
		writer = gin.DefaultWriter
	}
	id := ctx.Value(helper.RequestIdKey)
	if id == nil {
		id = helper.GenRequestID()
	}
	now := time.Now()
	_, _ = fmt.Fprintf(writer, "[%s] %v | %s | %s \n", level, now.Format("2006/01/02 - 15:04:05"), id, msg)
	SetupLogger()
}

func FatalLog(v ...any) {
	t := time.Now()
	_, _ = fmt.Fprintf(gin.DefaultErrorWriter, "[FATAL] %v | %v \n", t.Format("2006/01/02 - 15:04:05"), v)
	os.Exit(1)
}
