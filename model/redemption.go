package model

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"one-api/common"
)

type Redemption struct {
	Id           int    `json:"id"`
	UserId       int    `json:"user_id"`
	Key          string `json:"key" gorm:"type:char(32);uniqueIndex"`
	Status       int    `json:"status" gorm:"default:1"`
	Name         string `json:"name" gorm:"index"`
	Quota        int    `json:"quota" gorm:"default:100"`
	CreatedTime  int64  `json:"created_time" gorm:"bigint"`
	RedeemedTime int64  `json:"redeemed_time" gorm:"bigint"`
	Count        int    `json:"count" gorm:"-:all"` // only for api request
}

func GetAllRedemptions(startIdx int, num int) ([]*Redemption, error) {
	var redemptions []*Redemption
	var err error
	err = DB.Order("id desc").Limit(num).Offset(startIdx).Find(&redemptions).Error
	return redemptions, err
}

func SearchRedemptions(keyword string) (redemptions []*Redemption, err error) {
	err = DB.Where("id = ? or name LIKE ?", keyword, keyword+"%").Find(&redemptions).Error
	return redemptions, err
}

func GetRedemptionById(id int) (*Redemption, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	redemption := Redemption{Id: id}
	var err error = nil
	err = DB.First(&redemption, "id = ?", id).Error
	return &redemption, err
}

func Redeem(key string, userId int) (quota int, upgradedToVIP bool, err error) {
	upgradedToVIP = false // 初始化升级状态为 false（注意：这里不需要使用 var）
	if key == "" {
		return 0, false, errors.New("未提供兑换码")
	}
	if userId == 0 {
		return 0, false, errors.New("无效的 user id")
	}
	redemption := &Redemption{}

	keyCol := "`key`"
	if common.UsingPostgreSQL {
		keyCol = `"key"`
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Set("gorm:query_option", "FOR UPDATE").Where(keyCol+" = ?", key).First(redemption).Error
		if err != nil {
			return errors.New("无效的兑换码")
		}
		if redemption.Status != common.RedemptionCodeStatusEnabled {
			return errors.New("该兑换码已被使用")
		}
		err = tx.Model(&User{}).Where("id = ?", userId).Update("quota", gorm.Expr("quota + ?", redemption.Quota)).Error
		if err != nil {
			return err
		}
		redemption.RedeemedTime = common.GetTimestamp()
		redemption.Status = common.RedemptionCodeStatusUsed
		err = tx.Save(redemption).Error
		return err
	})
	if err != nil {
		return 0, false, errors.New("兑换失败，" + err.Error())
	}
	RecordLog(userId, LogTypeTopup, fmt.Sprintf("通过兑换码充值 %s", common.LogQuota(redemption.Quota)))

    // 获取用户信息
    user := &User{}
    err = DB.Where("id = ?", userId).First(user).Error
    if err != nil {
        return redemption.Quota, upgradedToVIP, errors.New("查询用户信息失败")
    }

    // 检查是否需要升级为 VIP
    if user.Group != "vip" && user.Quota >= 5*500000 {
        // 升级用户到 VIP
        err = DB.Model(&User{}).Where("id = ?", userId).Update("group", "vip").Error
        if err != nil {
            return redemption.Quota, upgradedToVIP, errors.New("升级 VIP 失败")
        }
        upgradedToVIP = true // 设置升级状态为 true
    }

    return redemption.Quota, upgradedToVIP, nil
}

func (redemption *Redemption) Insert() error {
	var err error
	err = DB.Create(redemption).Error
	return err
}

func (redemption *Redemption) SelectUpdate() error {
	// This can update zero values
	return DB.Model(redemption).Select("redeemed_time", "status").Updates(redemption).Error
}

// Update Make sure your token's fields is completed, because this will update non-zero values
func (redemption *Redemption) Update() error {
	var err error
	err = DB.Model(redemption).Select("name", "status", "quota", "redeemed_time").Updates(redemption).Error
	return err
}

func (redemption *Redemption) Delete() error {
	var err error
	err = DB.Delete(redemption).Error
	return err
}

func DeleteRedemptionById(id int) (err error) {
	if id == 0 {
		return errors.New("id 为空！")
	}
	redemption := Redemption{Id: id}
	err = DB.Where(redemption).First(&redemption).Error
	if err != nil {
		return err
	}
	return redemption.Delete()
}
