package playground

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 允许上传的图片类型：扩展名 -> 魔数前缀
var allowedImageTypes = map[string][]byte{
	".png":  {0x89, 0x50, 0x4E, 0x47},
	".jpg":  {0xFF, 0xD8, 0xFF},
	".jpeg": {0xFF, 0xD8, 0xFF},
	".gif":  {0x47, 0x49, 0x46, 0x38}, // GIF8
	".webp": {0x52, 0x49, 0x46, 0x46}, // RIFF
}

var imageContentTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
}

// ErrImageTooLarge 图片超出大小限制
var ErrImageTooLarge = errors.New("图片大小超出限制")

// ErrUnsupportedImageType 不支持的图片类型
var ErrUnsupportedImageType = errors.New("不支持的图片类型，仅支持 PNG / JPEG / GIF / WebP")

// uploadDir 图片临时目录
func uploadDir() string {
	return filepath.Join(os.TempDir(), "one-api-playground", "uploads")
}

func init() {
	_ = os.MkdirAll(uploadDir(), 0o755)
	go startUploadCleanupLoop()
}

func startUploadCleanupLoop() {
	// 每小时清理 6 小时前的临时图片
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		CleanupUploads(6 * time.Hour)
	}
}

// CleanupUploads 清理早于 maxAge 的上传文件
func CleanupUploads(maxAge time.Duration) {
	entries, err := os.ReadDir(uploadDir())
	if err != nil {
		return
	}
	threshold := time.Now().Add(-maxAge)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(threshold) {
			_ = os.Remove(filepath.Join(uploadDir(), e.Name()))
		}
	}
}

// SaveUploadedImage 校验并保存上传的图片，返回文件名
func SaveUploadedImage(file *multipart.FileHeader, maxSizeMB int) (string, error) {
	if maxSizeMB <= 0 {
		maxSizeMB = 5
	}
	limit := int64(maxSizeMB) * 1024 * 1024
	if file.Size > limit {
		return "", ErrImageTooLarge
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	magic, ok := allowedImageTypes[ext]
	if !ok {
		return "", ErrUnsupportedImageType
	}
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// 读取文件头校验魔数，防止伪造扩展名
	header := make([]byte, 16)
	n, _ := src.Read(header)
	if n < len(magic) {
		return "", ErrUnsupportedImageType
	}
	for i, b := range magic {
		if header[i] != b {
			return "", ErrUnsupportedImageType
		}
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	name := hex.EncodeToString(b) + ext
	dstPath := filepath.Join(uploadDir(), name)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	// 限制实际写入大小，防止 Content-Length 与实际不符
	written, err := io.Copy(dst, io.LimitReader(src, limit+1))
	if err != nil {
		_ = os.Remove(dstPath)
		return "", err
	}
	if written > limit {
		_ = os.Remove(dstPath)
		return "", ErrImageTooLarge
	}
	return name, nil
}

// GetUploadContentType 根据文件名返回 Content-Type
func GetUploadContentType(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if ct, ok := imageContentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

// ReadUploadFile 读取上传的图片文件内容（防路径穿越）
func ReadUploadFile(name string) ([]byte, error) {
	base := filepath.Base(filepath.Clean("/" + name))
	if base != name || base == "." || base == ".." || base == "/" {
		return nil, fmt.Errorf("非法的文件名")
	}
	ext := strings.ToLower(filepath.Ext(base))
	if _, ok := imageContentTypes[ext]; !ok {
		return nil, ErrUnsupportedImageType
	}
	data, err := os.ReadFile(filepath.Join(uploadDir(), base))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("文件为空或已过期")
	}
	return data, nil
}

// ReadUploadFileAsDataURL 读取图片并转为 base64 data URL（用于注入模型消息）
func ReadUploadFileAsDataURL(name string) (string, error) {
	data, err := ReadUploadFile(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("data:%s;base64,%s", GetUploadContentType(name), base64.StdEncoding.EncodeToString(data)), nil
}
