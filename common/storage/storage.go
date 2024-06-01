package storage

import (
	"one-api/common/storage/drives"

	"github.com/spf13/viper"
)

type Storage struct {
	drives map[string]StorageDrive
}

func InitStorage() {
	InitImgurStorage()
	InitSMStorage()
	InitALIOSSStorage()
}

func InitALIOSSStorage() {
	endpoint := viper.GetString("storage.alioss.endpoint")
	if endpoint == "" {
		return
	}
	accessKeyId := viper.GetString("storage.alioss.accessKeyId")
	if accessKeyId == "" {
		return
	}
	accessKeySecret := viper.GetString("storage.alioss.accessKeySecret")
	if accessKeySecret == "" {
		return
	}
	bucketName := viper.GetString("storage.alioss.bucketName")
	if bucketName == "" {

		return
	}

	aliUpload := drives.NewAliOSSUpload(endpoint, accessKeyId, accessKeySecret, bucketName)
	AddStorageDrive(aliUpload)
}

func InitSMStorage() {
	smSecret := viper.GetString("storage.smms.secret")
	if smSecret == "" {
		return
	}

	smUpload := drives.NewSMUpload(smSecret)
	AddStorageDrive(smUpload)
}

func InitImgurStorage() {
	imgurClientId := viper.GetString("storage.imgur.client_id")
	if imgurClientId == "" {
		return
	}

	imgurUpload := drives.NewImgurUpload(imgurClientId)
	AddStorageDrive(imgurUpload)
}
