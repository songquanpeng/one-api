package model

import (
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"one-api/common"
	"os"
	"path"
)

type File struct {
	Id              int    `json:"id"`
	Filename        string `json:"filename" gorm:"index"`
	Description     string `json:"description"`
	Uploader        string `json:"uploader"  gorm:"index"`
	UploaderId      int    `json:"uploader_id"  gorm:"index"`
	Link            string `json:"link" gorm:"unique;index"`
	UploadTime      string `json:"upload_time"`
	DownloadCounter int    `json:"download_counter"`
}

func GetAllFiles(startIdx int, num int) ([]*File, error) {
	var files []*File
	var err error
	err = DB.Order("id desc").Limit(num).Offset(startIdx).Find(&files).Error
	return files, err
}

func SearchFiles(keyword string) (files []*File, err error) {
	err = DB.Select([]string{"id", "filename", "description", "uploader", "uploader_id", "link", "upload_time", "download_counter"}).Where(
		"filename LIKE ? or uploader LIKE ? or uploader_id = ?", keyword+"%", keyword+"%", keyword).Find(&files).Error
	return files, err
}

func (file *File) Insert() error {
	var err error
	err = DB.Create(file).Error
	return err
}

// Delete Make sure link is valid! Because we will use os.Remove to delete it!
func (file *File) Delete() error {
	var err error
	err = DB.Delete(file).Error
	err = os.Remove(path.Join(common.UploadPath, file.Link))
	return err
}

func UpdateDownloadCounter(link string) {
	DB.Model(&File{}).Where("link = ?", link).UpdateColumn("download_counter", gorm.Expr("download_counter + 1"))
}
