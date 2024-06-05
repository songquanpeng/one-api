package model

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending OrderStatus = "pending"
	OrderStatusSuccess OrderStatus = "success"
	OrderStatusFailed  OrderStatus = "failed"
	OrderStatusClosed  OrderStatus = "closed"
)

type Order struct {
	ID            int            `json:"id"`
	UserId        int            `json:"user_id"`
	TradeNo       string         `json:"trade_no" gorm:"type:varchar(50);uniqueIndex"`
	GatewayNo     string         `json:"gateway_no" gorm:"type:varchar(100)"`
	Amount        int            `json:"amount" gorm:"default:0"`
	OrderAmount   float64        `json:"order_amount" gorm:"type:decimal(10,2);default:0"`
	OrderCurrency CurrencyType   `json:"order_currency" gorm:"type:varchar(16)"`
	Quota         int            `json:"quota" gorm:"type:int;default:0"`
	Fee           float64        `json:"fee" gorm:"type:decimal(10,2);default:0"`
	Status        OrderStatus    `json:"status" gorm:"type:varchar(32)"`
	CreatedAt     int            `json:"created_at"`
	UpdatedAt     int            `json:"-"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// 查询并关闭未完成的订单
func CloseUnfinishedOrder() error {
	// 关闭超过 3 小时未支付的订单
	unixTime := time.Now().Unix() - 3*3600
	return DB.Model(&Order{}).Where("status = ? AND created_at < ?", OrderStatusPending, unixTime).Update("status", OrderStatusClosed).Error
}

func GetOrderByTradeNo(tradeNo string) (*Order, error) {
	var order Order
	err := DB.Where("trade_no = ?", tradeNo).First(&order).Error
	return &order, err
}

func GetUserOrder(userId int, tradeNo string) (*Order, error) {
	var order Order
	err := DB.Where("user_id = ? AND trade_no = ?", userId, tradeNo).First(&order).Error
	return &order, err
}

func (o *Order) Insert() error {
	return DB.Create(o).Error
}

func (o *Order) Update() error {
	return DB.Save(o).Error
}

var allowedOrderFields = map[string]bool{
	"id":         true,
	"user_id":    true,
	"status":     true,
	"created_at": true,
}

type SearchOrderParams struct {
	UserId         int    `form:"user_id"`
	TradeNo        string `form:"trade_no"`
	GatewayNo      string `form:"gateway_no"`
	Status         string `form:"status"`
	StartTimestamp int64  `form:"start_timestamp"`
	EndTimestamp   int64  `form:"end_timestamp"`
	PaginationParams
}

func GetOrderList(params *SearchOrderParams) (*DataResult[Order], error) {
	var orders []*Order

	db := DB.Omit("key")

	if params.UserId != 0 {
		db = db.Where("user_id = ?", params.UserId)
	}

	if params.TradeNo != "" {
		db = db.Where("trade_no = ?", params.TradeNo)
	}

	if params.GatewayNo != "" {
		db = db.Where("gateway_no = ?", params.GatewayNo)
	}

	if params.Status != "" {
		db = db.Where("status = ?", params.Status)
	}

	if params.StartTimestamp != 0 {
		db = db.Where("created_at >= ?", params.StartTimestamp)
	}
	if params.EndTimestamp != 0 {
		db = db.Where("created_at <= ?", params.EndTimestamp)
	}

	return PaginateAndOrder(db, &params.PaginationParams, &orders, allowedOrderFields)
}
