package storage

import (
	"context"
	"fmt"
	"one-api/common/logger"
)

func (s *Storage) Upload(ctx context.Context, data []byte, fileName string) string {
	if ctx == nil {
		ctx = context.Background()
	}

	for driveName, drive := range s.drives {
		if drive == nil {
			continue
		}
		url, err := drive.Upload(data, fileName)
		if err != nil {
			logger.LogError(ctx, fmt.Sprintf("%s err: %s", driveName, err.Error()))
		} else {
			return url
		}
	}

	return ""
}

func Upload(data []byte, fileName string) string {
	//lint:ignore SA1029 reason: 需要使用该类型作为错误处理
	ctx := context.WithValue(context.Background(), logger.RequestIdKey, "Upload")

	return storageDrives.Upload(ctx, data, fileName)
}
