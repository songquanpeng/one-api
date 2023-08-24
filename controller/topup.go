package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	epay "github.com/star-horizon/go-epay"
	"log"
	"net/url"
	"one-api/common"
	"one-api/model"
	"strconv"
	"time"
)

type EpayRequest struct {
	Amount        int    `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	TopUpCode     string `json:"top_up_code"`
}

type AmountRequest struct {
	Amount    int    `json:"amount"`
	TopUpCode string `json:"top_up_code"`
}

//var client, _ = epay.NewClientWithUrl(&epay.Config{
//	PartnerID: "1096",
//	Key:       "n08V9LpE8JffA3NPP893689u8p39NV9J",
//}, "https://api.lempay.org")

var client, _ = epay.NewClientWithUrl(&epay.Config{
	PartnerID: "1064",
	Key:       "nqrrZ5RjR86mKP8rKkyrOY5Pg8NmYfKR",
}, "https://pay.yunjuw.cn")

func GetAmount(id int, count float64, topUpCode string) float64 {
	amount := count * 1.5
	if topUpCode != "" {
		if topUpCode == "nekoapi" {
			if id == 89 {
				amount = count * 0.8
			} else if id == 105 || id == 107 {
				amount = count * 1.2
			} else if id == 1 {
				amount = count * 1
			} else if id == 98 {
				amount = count * 1.1
			}
		}
	}
	return amount
}

func RequestEpay(c *gin.Context) {
	var req EpayRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": err.Error(), "data": 10})
		return
	}
	id := c.GetInt("id")
	amount := GetAmount(id, float64(req.Amount), req.TopUpCode)
	if id != 1 {
		if req.Amount < 10 {
			c.JSON(200, gin.H{"message": "最小充值10元", "data": amount, "count": 10})
			return
		}
	}
	if req.PaymentMethod == "zfb" {
		if amount > 2000 {
			c.JSON(200, gin.H{"message": "支付宝最大充值2000元", "data": amount, "count": 2000})
			return
		}
		req.PaymentMethod = "alipay"
	}
	if req.PaymentMethod == "wx" {
		if amount > 2000 {
			c.JSON(200, gin.H{"message": "微信最大充值2000元", "data": amount, "count": 2000})
			return
		}
		req.PaymentMethod = "wxpay"
	}

	returnUrl, _ := url.Parse("https://nekoapi.com/log")
	notifyUrl, _ := url.Parse("https://nekoapi.com/api/user/epay/notify")
	tradeNo := strconv.FormatInt(time.Now().Unix(), 10)
	payMoney := amount
	//if payMoney < 400 {
	//	payMoney = amount * 0.99
	//	if amount-payMoney > 2 {
	//		payMoney = amount - 2
	//	}
	//}
	uri, params, err := client.Purchase(&epay.PurchaseArgs{
		Type:           epay.PurchaseType(req.PaymentMethod),
		ServiceTradeNo: "A" + tradeNo,
		Name:           "B" + tradeNo,
		Money:          strconv.FormatFloat(payMoney, 'f', 2, 64),
		Device:         epay.PC,
		NotifyUrl:      notifyUrl,
		ReturnUrl:      returnUrl,
	})
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}
	topUp := &model.TopUp{
		UserId:     id,
		Amount:     req.Amount,
		Money:      int(amount),
		TradeNo:    "A" + tradeNo,
		CreateTime: time.Now().Unix(),
		Status:     "pending",
	}
	err = topUp.Insert()
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}
	c.JSON(200, gin.H{"message": "success", "data": params, "url": uri})
}

func EpayNotify(c *gin.Context) {
	params := lo.Reduce(lo.Keys(c.Request.URL.Query()), func(r map[string]string, t string, i int) map[string]string {
		r[t] = c.Request.URL.Query().Get(t)
		return r
	}, map[string]string{})
	verifyInfo, err := client.Verify(params)
	if err == nil && verifyInfo.VerifyStatus {
		_, err := c.Writer.Write([]byte("success"))
		if err != nil {
			log.Println("易支付回调写入失败")
		}
	} else {
		_, err := c.Writer.Write([]byte("fail"))
		if err != nil {
			log.Println("易支付回调写入失败")
		}
	}

	if verifyInfo.TradeStatus == epay.StatusTradeSuccess {
		log.Println(verifyInfo)
		topUp := model.GetTopUpByTradeNo(verifyInfo.ServiceTradeNo)
		if topUp.Status == "pending" {
			topUp.Status = "success"
			err := topUp.Update()
			if err != nil {
				log.Printf("易支付回调更新订单失败: %v", topUp)
				return
			}
			//user, _ := model.GetUserById(topUp.UserId, false)
			//user.Quota += topUp.Amount * 500000
			err = model.IncreaseUserQuota(topUp.UserId, topUp.Amount*500000)
			if err != nil {
				log.Printf("易支付回调更新用户失败: %v", topUp)
				return
			}
			log.Printf("易支付回调更新用户成功 %v", topUp)
			model.RecordLog(topUp.UserId, model.LogTypeTopup, fmt.Sprintf("使用在线充值成功，充值金额: %v", common.LogQuota(topUp.Amount*500000)))
		}
	} else {
		log.Printf("易支付异常回调: %v", verifyInfo)
	}
}

func RequestAmount(c *gin.Context) {
	var req AmountRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": err.Error(), "data": 10})
		return
	}
	id := c.GetInt("id")
	if id != 1 {
		if req.Amount < 10 {
			c.JSON(200, gin.H{"message": "最小充值10刀", "data": GetAmount(id, 10, req.TopUpCode), "count": 10})
			return
		}
		//if req.Amount > 1500 {
		//	c.JSON(200, gin.H{"message": "最大充值1000刀", "data": GetAmount(id, 1000, req.TopUpCode), "count": 1500})
		//	return
		//}
	}

	c.JSON(200, gin.H{"message": "success", "data": GetAmount(id, float64(req.Amount), req.TopUpCode)})
}
