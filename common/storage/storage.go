package storage

import (
	"one-api/common/storage/drives"

	"github.com/spf13/viper"
)

type Storage struct {
	drives map[string]StorageDrive
}

func InitStorage() {
	InitSMStorage()
}

func InitSMStorage() {
	smSecret := viper.GetString("storage.smms.secret")
	if smSecret == "" {
		return
	}

	smUpload := drives.NewSMUpload(smSecret)
	AddStorageDrive(smUpload)
}
