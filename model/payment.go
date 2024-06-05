package model

import (
	"one-api/common/utils"

	"gorm.io/gorm"
)

type CurrencyType string

const (
	CurrencyTypeUSD CurrencyType = "USD"
	CurrencyTypeCNY CurrencyType = "CNY"
)

type Payment struct {
	ID           int            `json:"id"`
	Type         string         `json:"type" form:"type" gorm:"type:varchar(16)"`
	UUID         string         `json:"uuid" form:"uuid" gorm:"type:char(32);uniqueIndex"`
	Name         string         `json:"name" form:"name" gorm:"type:varchar(255); not null"`
	Icon         string         `json:"icon" form:"icon" gorm:"type:varchar(300)"`
	NotifyDomain string         `json:"notify_domain" form:"notify_domain" gorm:"type:varchar(300)"`
	FixedFee     float64        `json:"fixed_fee" form:"fixed_fee" gorm:"type:decimal(10,2); default:0.00"`
	PercentFee   float64        `json:"percent_fee" form:"percent_fee" gorm:"type:decimal(10,2); default:0.00"`
	Currency     CurrencyType   `json:"currency" form:"currency" gorm:"type:varchar(5)"`
	Config       string         `json:"config" form:"config" gorm:"type:text"`
	Sort         int            `json:"sort" form:"sort" gorm:"default:1"`
	Enable       *bool          `json:"enable" form:"enable" gorm:"default:true"`
	CreatedAt    int64          `json:"created_at" gorm:"bigint"`
	UpdatedAt    int64          `json:"-" gorm:"bigint"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func GetPaymentByID(id int) (*Payment, error) {
	var payment Payment
	err := DB.First(&payment, id).Error
	return &payment, err
}

func GetPaymentByUUID(uuid string) (*Payment, error) {
	var payment Payment
	err := DB.Where("uuid = ? AND enable = ?", uuid, true).First(&payment).Error
	return &payment, err
}

var allowedPaymentOrderFields = map[string]bool{
	"id":         true,
	"uuid":       true,
	"name":       true,
	"type":       true,
	"sort":       true,
	"enable":     true,
	"created_at": true,
}

type SearchPaymentParams struct {
	Payment
	PaginationParams
}

func GetPanymentList(params *SearchPaymentParams) (*DataResult[Payment], error) {
	var payments []*Payment

	db := DB.Omit("key")

	if params.Type != "" {
		db = db.Where("type = ?", params.Type)
	}

	if params.Name != "" {
		db = db.Where("name LIKE ?", params.Name+"%")
	}

	if params.UUID != "" {
		db = db.Where("uuid = ?", params.UUID)
	}

	if params.Currency != "" {
		db = db.Where("currency = ?", params.Currency)
	}

	return PaginateAndOrder(db, &params.PaginationParams, &payments, allowedPaymentOrderFields)
}

func GetUserPaymentList() ([]*Payment, error) {
	var payments []*Payment
	err := DB.Model(payments).Select("uuid, name, icon, fixed_fee, percent_fee, currency, sort").Where("enable = ?", true).Find(&payments).Error
	return payments, err
}

func (p *Payment) Insert() error {
	p.UUID = utils.GetUUID()
	return DB.Create(p).Error
}

func (p *Payment) Update(overwrite bool) error {
	var err error

	if overwrite {
		err = DB.Model(p).Select("*").Updates(p).Error
	} else {
		err = DB.Model(p).Updates(p).Error
	}

	return err
}

func (p *Payment) Delete() error {
	return DB.Delete(p).Error
}
