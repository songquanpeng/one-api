package drives

import (
	"bytes"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type AliOSSUpload struct {
	Endpoint        string
	AccessKeyId     string
	AccessKeySecret string
	BucketName      string
}

func NewAliOSSUpload(endpoint, accessKeyId, accessKeySecret, bucketName string) *AliOSSUpload {
	return &AliOSSUpload{
		Endpoint:        endpoint,
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
		BucketName:      bucketName,
	}
}

func (a *AliOSSUpload) Name() string {
	return "AliOSS"
}

func (a *AliOSSUpload) Upload(data []byte, fileName string) (string, error) {
	// Create OSS Client
	client, err := oss.New(a.Endpoint, a.AccessKeyId, a.AccessKeySecret)
	if err != nil {
		return "", fmt.Errorf("creating OSS client: %w", err)
	}

	// Create Bucket
	bucket, err := client.Bucket(a.BucketName)
	if err != nil {
		return "", fmt.Errorf("getting bucket: %w", err)
	}

	// Upload File
	reader := bytes.NewReader(data)
	err = bucket.PutObject(fileName, reader)
	if err != nil {
		return "", fmt.Errorf("uploading file: %w", err)
	}

	// Get Object URL
	objectURL, err := bucket.SignURL(fileName, oss.HTTPGet, 3600)
	if err != nil {
		return "", fmt.Errorf("signing object URL: %w", err)
	}

	return objectURL, nil
}
