package model

import (
	_ "gorm.io/driver/sqlite"
)

type Channel struct {
	Id           int    `json:"id"`
	Type         int    `json:"type" gorm:"default:0"`
	Key          string `json:"key"`
	Status       int    `json:"status" gorm:"default:1"`
	Name         string `json:"name" gorm:"unique;index"`
	Weight       int    `json:"weight"`
	CreatedTime  int64  `json:"created_time" gorm:"bigint"`
	AccessedTime int64  `json:"accessed_time" gorm:"bigint"`
}

func GetAllChannels(startIdx int, num int) ([]*Channel, error) {
	var channels []*Channel
	var err error
	err = DB.Order("id desc").Limit(num).Offset(startIdx).Omit("key").Find(&channels).Error
	return channels, err
}

func SearchChannels(keyword string) (channels []*Channel, err error) {
	err = DB.Omit("key").Where("id = ? or name LIKE ?", keyword, keyword+"%").Find(&channels).Error
	return channels, err
}

func GetChannelById(id int) (*Channel, error) {
	channel := Channel{Id: id}
	var err error = nil
	err = DB.Omit("key").First(&channel, "id = ?", id).Error
	return &channel, err
}

func (channel *Channel) Insert() error {
	var err error
	err = DB.Create(channel).Error
	return err
}

func (channel *Channel) Update() error {
	var err error
	err = DB.Model(channel).Updates(channel).Error
	return err
}

func (channel *Channel) Delete() error {
	var err error
	err = DB.Delete(channel).Error
	return err
}
