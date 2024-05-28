package model

import (
	"errors"
	"one-api/common/utils"
)

type TelegramMenu struct {
	Id           int    `json:"id"`
	Command      string `json:"command" gorm:"type:varchar(32);uniqueIndex"`
	Description  string `json:"description" gorm:"type:varchar(255);default:''"`
	ParseMode    string `json:"parse_mode" gorm:"type:varchar(255);default:'MarkdownV2'"`
	ReplyMessage string `json:"reply_message"`
}

var allowedTelegramMenusOrderFields = map[string]bool{
	"id":      true,
	"command": true,
}

func GetTelegramMenusList(params *GenericParams) (*DataResult[TelegramMenu], error) {
	var menus []*TelegramMenu
	db := DB
	if params.Keyword != "" {
		db = db.Where("id = ? or command LIKE ?", utils.String2Int(params.Keyword), params.Keyword+"%")
	}

	return PaginateAndOrder[TelegramMenu](db, &params.PaginationParams, &menus, allowedTelegramMenusOrderFields)
}

// 查询菜单列表  只查询command和description
func GetTelegramMenus() ([]*TelegramMenu, error) {
	var menus []*TelegramMenu
	err := DB.Select("command, description").Find(&menus).Error
	return menus, err
}

// 根据command查询菜单
func GetTelegramMenuByCommand(command string) (*TelegramMenu, error) {
	menu := &TelegramMenu{}
	err := DB.Where("command = ?", command).First(menu).Error
	return menu, err
}

func GetTelegramMenuById(id int) (*TelegramMenu, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	telegramMenu := TelegramMenu{Id: id}
	var err error = nil
	err = DB.First(&telegramMenu, "id = ?", id).Error
	return &telegramMenu, err
}

func IsTelegramCommandAlreadyTaken(command string, id int) bool {
	query := DB.Where("command = ?", command)
	if id != 0 {
		query = query.Not("id", id)
	}
	return query.Find(&TelegramMenu{}).RowsAffected == 1
}

func (menu *TelegramMenu) Insert() error {
	return DB.Create(menu).Error
}

func (menu *TelegramMenu) Update() error {
	return DB.Model(menu).Updates(menu).Error
}

func (menu *TelegramMenu) Delete() error {
	return DB.Delete(menu).Error
}
