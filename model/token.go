package model

import (
	"errors"
	_ "gorm.io/driver/sqlite"
)

type Token struct {
	Id           int    `json:"id"`
	UserId       int    `json:"user_id"`
	Key          string `json:"key"`
	Status       int    `json:"status" gorm:"default:1"`
	Name         string `json:"name" gorm:"index" `
	CreatedTime  int64  `json:"created_time" gorm:"bigint"`
	AccessedTime int64  `json:"accessed_time" gorm:"bigint"`
}

func GetAllUserTokens(userId int, startIdx int, num int) ([]*Token, error) {
	var tokens []*Token
	var err error
	err = DB.Where("user_id = ?", userId).Order("id desc").Limit(num).Offset(startIdx).Find(&tokens).Error
	return tokens, err
}

func SearchUserTokens(userId int, keyword string) (tokens []*Token, err error) {
	err = DB.Where("user_id = ?", userId).Where("id = ? or name LIKE ?", keyword, keyword+"%").Find(&tokens).Error
	return tokens, err
}

func GetTokenByIds(id int, userId int) (*Token, error) {
	if id == 0 || userId == 0 {
		return nil, errors.New("id 或 userId 为空！")
	}
	token := Token{Id: id, UserId: userId}
	var err error = nil
	err = DB.First(&token, "id = ? and user_id = ?", id, userId).Error
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

func DeleteTokenById(id int, userId int) (err error) {
	// Why we need userId here? In case user want to delete other's token.
	if id == 0 || userId == 0 {
		return errors.New("id 或 userId 为空！")
	}
	token := Token{Id: id, UserId: userId}
	err = DB.Where(token).First(&token).Error
	if err != nil {
		return err
	}
	return token.Delete()
}
