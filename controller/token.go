package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/network"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/model"
)

func GetAllTokens(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	myRole := c.GetInt(ctxkey.Role)
	if myRole == model.RoleRootUser {
		uId, _ := strconv.Atoi(c.Query("user_id"))
		if uId != 0 {
			userId = uId
		}
	}

	order := c.Query("order")
	tokens, err := model.GetAllUserTokens(userId, p*config.ItemsPerPage, config.ItemsPerPage, order)

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
	userId := c.GetInt(ctxkey.Id)
	keyword := c.Query("keyword")
	myRole := c.GetInt(ctxkey.Role)
	if myRole == model.RoleRootUser {
		uId, _ := strconv.Atoi(c.Query("user_id"))
		if uId != 0 {
			userId = uId
		}
	}
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
	userId := c.GetInt(ctxkey.Id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	myRole := c.GetInt(ctxkey.Role)
	if myRole == model.RoleRootUser {
		uId, _ := strconv.Atoi(c.Query("user_id"))
		if uId != 0 {
			userId = uId
		}
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
	tokenId := c.GetInt(ctxkey.TokenId)
	userId := c.GetInt(ctxkey.Id)
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

func validateToken(c *gin.Context, token model.Token) error {
	if len(token.Name) > 30 {
		return fmt.Errorf("令牌名称过长")
	}
	if token.Subnet != nil && *token.Subnet != "" {
		err := network.IsValidSubnets(*token.Subnet)
		if err != nil {
			return fmt.Errorf("无效的网段：%s", err.Error())
		}
	}
	return nil
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
	err = validateToken(c, token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("参数错误：%s", err.Error()),
		})
		return
	}

	cleanToken := model.Token{
		UserId:         c.GetInt(ctxkey.Id),
		Name:           token.Name,
		Key:            random.GenerateKey(),
		CreatedTime:    helper.GetTimestamp(),
		AccessedTime:   helper.GetTimestamp(),
		ExpiredTime:    token.ExpiredTime,
		RemainQuota:    token.RemainQuota,
		UnlimitedQuota: token.UnlimitedQuota,
		Models:         token.Models,
		Subnet:         token.Subnet,
	}
	// if the user is root and add the token for other user, set the user id
	myRole := c.GetInt(ctxkey.Role)
	if myRole == model.RoleRootUser {
		if token.UserId == 0 {
			cleanToken.UserId = c.GetInt(ctxkey.Id)
		} else {
			cleanToken.UserId = token.UserId
		}
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
		"data":    cleanToken,
	})
	return
}

func DeleteToken(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.GetInt(ctxkey.Id)
	myRole := c.GetInt(ctxkey.Role)
	// if the user is root and delete the token for other user, set the user id
	if myRole == model.RoleRootUser {
		uId, _ := strconv.Atoi(c.Query("user_id"))
		if uId != 0 {
			userId = uId
		}
	}
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

func UpdateToken(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	statusOnly := c.Query("status_only")
	token := model.Token{}
	err := c.ShouldBindJSON(&token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	err = validateToken(c, token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("参数错误：%s", err.Error()),
		})
		return
	}
	// if the user is root and update the token for other user, set the user id
	myRole := c.GetInt(ctxkey.Role)
	if myRole == model.RoleRootUser {
		if token.UserId == 0 {
			userId = token.UserId
		}
	}

	cleanToken, err := model.GetTokenByIds(token.Id, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if token.Status == model.TokenStatusEnabled {
		if cleanToken.Status == model.TokenStatusExpired && cleanToken.ExpiredTime <= helper.GetTimestamp() && cleanToken.ExpiredTime != -1 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "令牌已过期，无法启用，请先修改令牌过期时间，或者设置为永不过期",
			})
			return
		}
		if cleanToken.Status == model.TokenStatusExhausted && cleanToken.RemainQuota <= 0 && !cleanToken.UnlimitedQuota {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "令牌可用额度已用尽，无法启用，请先修改令牌剩余额度，或者设置为无限额度",
			})
			return
		}
	}
	if statusOnly != "" {
		cleanToken.Status = token.Status
	} else {
		// If you add more fields, please also update token.Update()
		cleanToken.Name = token.Name
		cleanToken.ExpiredTime = token.ExpiredTime
		cleanToken.RemainQuota = token.RemainQuota
		cleanToken.UnlimitedQuota = token.UnlimitedQuota
		cleanToken.Models = token.Models
		cleanToken.Subnet = token.Subnet
	}
	err = cleanToken.Update()
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
		"data":    cleanToken,
	})
	return
}
