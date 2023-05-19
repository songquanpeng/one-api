package model

import (
	"errors"
	"fmt"
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"one-api/common"
)

type Token struct {
	Id             int    `json:"id"`
	UserId         int    `json:"user_id"`
	Key            string `json:"key" gorm:"type:char(32);uniqueIndex"`
	Status         int    `json:"status" gorm:"default:1"`
	Name           string `json:"name" gorm:"index" `
	CreatedTime    int64  `json:"created_time" gorm:"bigint"`
	AccessedTime   int64  `json:"accessed_time" gorm:"bigint"`
	ExpiredTime    int64  `json:"expired_time" gorm:"bigint;default:-1"` // -1 means never expired
	RemainQuota    int    `json:"remain_quota" gorm:"default:0"`
	UnlimitedQuota bool   `json:"unlimited_quota" gorm:"default:false"`
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

func ValidateUserToken(key string) (token *Token, err error) {
	if key == "" {
		return nil, errors.New("N/A (no text provided) token")
	}
	token = &Token{}
	err = DB.Where("`key` = ?", key).First(token).Error
	if err == nil {
		if token.Status != common.TokenStatusEnabled {
			return nil, errors.New("The token status is not enabled.")
		}
		if token.ExpiredTime != -1 && token.ExpiredTime < common.GetTimestamp() {
			token.Status = common.TokenStatusExpired
			err := token.SelectUpdate()
			if err != nil {
				common.SysError("Failed to update token status.：" + err.Error())
			}
			return nil, errors.New("该 token 已过期")
		}
		if !token.UnlimitedQuota && token.RemainQuota <= 0 {
			token.Status = common.TokenStatusExhausted
			err := token.SelectUpdate()
			if err != nil {
				common.SysError("Failed to update token status.：" + err.Error())
			}
			return nil, errors.New("This token's quota has been exhausted.")
		}
		go func() {
			token.AccessedTime = common.GetTimestamp()
			err := token.SelectUpdate()
			if err != nil {
				common.SysError("Failed to update token.：" + err.Error())
			}
		}()
		return token, nil
	}
	return nil, errors.New("Invalid token")
}

func GetTokenByIds(id int, userId int) (*Token, error) {
	if id == 0 || userId == 0 {
		return nil, errors.New("id or userId is empty!")
	}
	token := Token{Id: id, UserId: userId}
	var err error = nil
	err = DB.First(&token, "id = ? and user_id = ?", id, userId).Error
	return &token, err
}

func GetTokenById(id int) (*Token, error) {
	if id == 0 {
		return nil, errors.New("id Is empty.！")
	}
	token := Token{Id: id}
	var err error = nil
	err = DB.First(&token, "id = ?", id).Error
	return &token, err
}

func (token *Token) Insert() error {
	var err error
	err = DB.Create(token).Error
	return err
}

// Update Make sure your token's fields is completed, because this will update non-zero values
func (token *Token) Update() error {
	var err error
	err = DB.Model(token).Select("name", "status", "expired_time", "remain_quota", "unlimited_quota").Updates(token).Error
	return err
}

func (token *Token) SelectUpdate() error {
	// This can update zero values
	return DB.Model(token).Select("accessed_time", "status").Updates(token).Error
}

func (token *Token) Delete() error {
	var err error
	err = DB.Delete(token).Error
	return err
}

func DeleteTokenById(id int, userId int) (err error) {
	// Why we need userId here? In case user want to delete other's token.
	if id == 0 || userId == 0 {
		return errors.New("id or userId is empty.！")
	}
	token := Token{Id: id, UserId: userId}
	err = DB.Where(token).First(&token).Error
	if err != nil {
		return err
	}
	return token.Delete()
}

func IncreaseTokenQuota(id int, quota int) (err error) {
	if quota < 0 {
		return errors.New("quota Cannot be negative.！")
	}
	err = DB.Model(&Token{}).Where("id = ?", id).Update("remain_quota", gorm.Expr("remain_quota + ?", quota)).Error
	return err
}

func DecreaseTokenQuota(id int, quota int) (err error) {
	if quota < 0 {
		return errors.New("quota Cannot be negative.！")
	}
	err = DB.Model(&Token{}).Where("id = ?", id).Update("remain_quota", gorm.Expr("remain_quota - ?", quota)).Error
	return err
}

func PreConsumeTokenQuota(tokenId int, quota int) (err error) {
	if quota < 0 {
		return errors.New("quota Cannot be negative.！")
	}
	token, err := GetTokenById(tokenId)
	if err != nil {
		return err
	}
	if !token.UnlimitedQuota && token.RemainQuota < quota {
		return errors.New("Token quota insufficient.")
	}
	userQuota, err := GetUserQuota(token.UserId)
	if err != nil {
		return err
	}
	if userQuota < quota {
		return errors.New("User quota insufficient.")
	}
	quotaTooLow := userQuota >= common.QuotaRemindThreshold && userQuota-quota < common.QuotaRemindThreshold
	noMoreQuota := userQuota-quota <= 0
	if quotaTooLow || noMoreQuota {
		go func() {
			email, err := GetUserEmail(token.UserId)
			if err != nil {
				common.SysError("Failed to get user email.：" + err.Error())
			}
			prompt := "Your quota is running low."
			if noMoreQuota {
				prompt = "Your quota has run out."
			}
			if email != "" {
				topUpLink := fmt.Sprintf("%s/topup", common.ServerAddress)
				err = common.SendEmail(prompt, email,
					fmt.Sprintf("%s，Current remaining quota is %d，To avoid any disruption to your usage,，Please top up your account promptly.。<br/>Top-up link.：<a href='%s'>%s</a>", prompt, userQuota, topUpLink, topUpLink))
				if err != nil {
					common.SysError("Failed to send email.：" + err.Error())
				}
			}
		}()
	}
	if !token.UnlimitedQuota {
		err = DecreaseTokenQuota(tokenId, quota)
		if err != nil {
			return err
		}
	}
	err = DecreaseUserQuota(token.UserId, quota)
	return err
}

func PostConsumeTokenQuota(tokenId int, quota int) (err error) {
	token, err := GetTokenById(tokenId)
	if quota > 0 {
		err = DecreaseUserQuota(token.UserId, quota)
	} else {
		err = IncreaseUserQuota(token.UserId, -quota)
	}
	if err != nil {
		return err
	}
	if !token.UnlimitedQuota {
		if quota > 0 {
			err = DecreaseTokenQuota(tokenId, quota)
		} else {
			err = IncreaseTokenQuota(tokenId, -quota)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
