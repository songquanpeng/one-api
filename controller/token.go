package controller

import (
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAllTokens(c *gin.Context) {
	userId := c.GetInt("id")
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	tokens, err := model.GetAllUserTokens(userId, p*common.ItemsPerPage, common.ItemsPerPage)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    tokens,
	})
	return
}

func SearchTokens(c *gin.Context) {
	userId := c.GetInt("id")
	keyword := c.Query("keyword")
	tokens, err := model.SearchUserTokens(userId, keyword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    tokens,
	})
	return
}

func GetToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	userId := c.GetInt("id")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	token, err := model.GetTokenByIds(id, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    token,
	})
	return
}

func GetTokenStatus(c *gin.Context) {
	tokenId := c.GetInt("token_id")
	userId := c.GetInt("id")
	token, err := model.GetTokenByIds(tokenId, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	expiredAt := token.ExpiredTime
	if expiredAt == -1 {
		expiredAt = 0
	}
	c.JSON(http.StatusOK, gin.H{
		"object":          "credit_summary",
		"total_granted":   token.RemainQuota,
		"total_used":      0, // not supported currently
		"total_available": token.RemainQuota,
		"expires_at":      expiredAt * 1000,
	})
}

func AddToken(c *gin.Context) {
	token := model.Token{}
	err := c.ShouldBindJSON(&token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if len(token.Name) > 30 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "令牌名称过长",
		})
		return
	}
	cleanToken := model.Token{
		UserId:         c.GetInt("id"),
		Name:           token.Name,
		Key:            common.GenerateKey(),
		CreatedTime:    common.GetTimestamp(),
		AccessedTime:   common.GetTimestamp(),
		ExpiredTime:    token.ExpiredTime,
		RemainQuota:    token.RemainQuota,
		UnlimitedQuota: token.UnlimitedQuota,
	}
	err = cleanToken.Insert()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func DeleteToken(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.GetInt("id")
	err := model.DeleteTokenById(id, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

type updateTokenDto struct {
	Id             int     `json:"id"`
	Status         int     `json:"status" gorm:"default:1"`
	Name           *string `json:"name" gorm:"index" `
	ExpiredTime    *int64  `json:"expired_time" gorm:"bigint;default:-1"` // -1 means never expired
	RemainQuota    *int    `json:"remain_quota" gorm:"default:0"`
	UnlimitedQuota *bool   `json:"unlimited_quota" gorm:"default:false"`
	// AddRemainQuota add or subtract remain quota
	AddRemainQuota int `json:"add_remain_quota"`
	// AddUsedQuota add or subtract used quota
	AddUsedQuota int `json:"add_used_quota"`
}

func UpdateToken(c *gin.Context) {
	userId := c.GetInt("id")
	statusOnly := c.Query("status_only")
	tokenPatch := new(updateTokenDto)
	if err := c.ShouldBindJSON(tokenPatch); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "parse request join: " + err.Error(),
		})
		return
	}

	if tokenPatch.Name != nil &&
		(len(*tokenPatch.Name) > 30 || len(*tokenPatch.Name) == 0) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "令牌名称错误，长度应在 1-30 之间",
		})
		return
	}

	cleanToken, err := model.GetTokenByIds(tokenPatch.Id, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("get token by id %d: %s", tokenPatch.Id, err.Error()),
		})
		return
	}

	if tokenPatch.Status == common.TokenStatusEnabled {
		if cleanToken.Status == common.TokenStatusExpired &&
			cleanToken.ExpiredTime <= common.GetTimestamp() &&
			cleanToken.ExpiredTime != -1 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "令牌已过期，无法启用，请先修改令牌过期时间，或者设置为永不过期",
			})
			return
		}
		if cleanToken.Status == common.TokenStatusExhausted &&
			cleanToken.RemainQuota <= 0 &&
			!cleanToken.UnlimitedQuota {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "令牌可用额度已用尽，无法启用，请先修改令牌剩余额度，或者设置为无限额度",
			})
			return
		}
	}
	if statusOnly != "" {
		cleanToken.Status = tokenPatch.Status
	} else {
		// If you add more fields, please also update tokenPatch.Update()
		if tokenPatch.Name != nil {
			cleanToken.Name = *tokenPatch.Name
		}
		if tokenPatch.ExpiredTime != nil {
			cleanToken.ExpiredTime = *tokenPatch.ExpiredTime
		}
		if tokenPatch.RemainQuota != nil {
			cleanToken.RemainQuota = *tokenPatch.RemainQuota
		}
		if tokenPatch.UnlimitedQuota != nil {
			cleanToken.UnlimitedQuota = *tokenPatch.UnlimitedQuota
		}
	}

	cleanToken.RemainQuota += tokenPatch.AddRemainQuota
	cleanToken.UsedQuota += tokenPatch.AddUsedQuota

	if err = cleanToken.Update(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "update token: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    cleanToken,
	})
	return
}
