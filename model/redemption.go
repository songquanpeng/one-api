package model

import (
	"errors"
	"fmt"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/utils"

	"gorm.io/gorm"
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

var allowedRedemptionslOrderFields = map[string]bool{
	"id":            true,
	"name":          true,
	"status":        true,
	"quota":         true,
	"created_time":  true,
	"redeemed_time": true,
}

func GetRedemptionsList(params *GenericParams) (*DataResult[Redemption], error) {
	var redemptions []*Redemption
	db := DB
	if params.Keyword != "" {
		db = db.Where("id = ? or name LIKE ?", utils.String2Int(params.Keyword), params.Keyword+"%")
	}

	return PaginateAndOrder[Redemption](db, &params.PaginationParams, &redemptions, allowedRedemptionslOrderFields)
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

func Redeem(key string, userId int) (quota int, err error) {
	if key == "" {
		return 0, errors.New("未提供兑换码")
	}
	if userId == 0 {
		return 0, errors.New("无效的 user id")
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
		if redemption.Status != config.RedemptionCodeStatusEnabled {
			return errors.New("该兑换码已被使用")
		}
		err = tx.Model(&User{}).Where("id = ?", userId).Update("quota", gorm.Expr("quota + ?", redemption.Quota)).Error
		if err != nil {
			return err
		}
		redemption.RedeemedTime = utils.GetTimestamp()
		redemption.Status = config.RedemptionCodeStatusUsed
		err = tx.Save(redemption).Error
		return err
	})
	if err != nil {
		return 0, errors.New("兑换失败，" + err.Error())
	}
	RecordLog(userId, LogTypeTopup, fmt.Sprintf("通过兑换码充值 %s", common.LogQuota(redemption.Quota)))
	return redemption.Quota, nil
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

type RedemptionStatistics struct {
	Count  int64 `json:"count"`
	Quota  int64 `json:"quota"`
	Status int   `json:"status"`
}

func GetStatisticsRedemption() (redemptionStatistics []*RedemptionStatistics, err error) {
	err = DB.Model(&Redemption{}).Select("status", "count(*) as count", "sum(quota) as quota").Where("status != ?", 2).Group("status").Scan(&redemptionStatistics).Error
	return redemptionStatistics, err
}

type RedemptionStatisticsGroup struct {
	Date      string `json:"date"`
	Quota     int64  `json:"quota"`
	UserCount int64  `json:"user_count"`
}

func GetStatisticsRedemptionByPeriod(startTimestamp, endTimestamp int64) (redemptionStatistics []*RedemptionStatisticsGroup, err error) {
	groupSelect := getTimestampGroupsSelect("redeemed_time", "day", "date")

	err = DB.Raw(`
		SELECT `+groupSelect+`,
		sum(quota) as quota,
		count(distinct user_id) as user_count
		FROM redemptions
		WHERE status=3
		AND redeemed_time BETWEEN ? AND ?
		GROUP BY date
		ORDER BY date
	`, startTimestamp, endTimestamp).Scan(&redemptionStatistics).Error

	return redemptionStatistics, err
}
