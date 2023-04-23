package model

import (
	_ "gorm.io/driver/sqlite"
)

type Token struct {
	Id           int    `json:"id"`
	UserId       int    `json:"user_id"`
	Key          string `json:"key"`
	Status       int    `json:"status" gorm:"default:1"`
	Name         string `json:"name" gorm:"unique;index"`
	CreatedTime  int64  `json:"created_time" gorm:"bigint"`
	AccessedTime int64  `json:"accessed_time" gorm:"bigint"`
}

func GetAllUserTokens(userId int, startIdx int, num int) ([]*Token, error) {
	var tokens []*Token
	var err error
	err = DB.Where("userId = ?", userId).Order("id desc").Limit(num).Offset(startIdx).Omit("key").Find(&tokens).Error
	return tokens, err
}

func SearchUserTokens(userId int, keyword string) (tokens []*Token, err error) {
	err = DB.Where("userId = ?", userId).Omit("key").Where("id = ? or name LIKE ?", keyword, keyword+"%").Find(&tokens).Error
	return tokens, err
}

func GetTokenById(id int) (*Token, error) {
	token := Token{Id: id}
	var err error = nil
	err = DB.Omit("key").Select([]string{"id", "type"}).First(&token, "id = ?", id).Error
	return &token, err
}

func (token *Token) Insert() error {
	var err error
	err = DB.Create(token).Error
	return err
}

func (token *Token) Update() error {
	var err error
	err = DB.Model(token).Updates(token).Error
	return err
}

func (token *Token) Delete() error {
	var err error
	err = DB.Delete(token).Error
	return err
}
