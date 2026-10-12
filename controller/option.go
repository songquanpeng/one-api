package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/i18n"
	"github.com/songquanpeng/one-api/model"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"

	"github.com/gin-gonic/gin"
)

func GetOptions(c *gin.Context) {
	var options []*model.Option
	config.OptionMapRWMutex.Lock()
	for k, v := range config.OptionMap {
		if strings.HasSuffix(k, "Token") || strings.HasSuffix(k, "Secret") {
			continue
		}
		options = append(options, &model.Option{
			Key:   k,
			Value: helper.Interface2String(v),
		})
	}
	config.OptionMapRWMutex.Unlock()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    options,
	})
	return
}

func UpdateOption(c *gin.Context) {
	var option model.Option
	err := json.NewDecoder(c.Request.Body).Decode(&option)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	switch option.Key {
	case "Theme":
		if !config.ValidThemes[option.Value] {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无效的主题",
			})
			return
		}
	case "GitHubOAuthEnabled":
		if option.Value == "true" && config.GitHubClientId == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 GitHub OAuth，请先填入 GitHub Client Id 以及 GitHub Client Secret！",
			})
			return
		}
	case "EmailDomainRestrictionEnabled":
		if option.Value == "true" && len(config.EmailDomainWhitelist) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用邮箱域名限制，请先填入限制的邮箱域名！",
			})
			return
		}
	case "WeChatAuthEnabled":
		if option.Value == "true" && config.WeChatServerAddress == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用微信登录，请先填入微信登录相关配置信息！",
			})
			return
		}
	case "TurnstileCheckEnabled":
		if option.Value == "true" && config.TurnstileSiteKey == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法启用 Turnstile 校验，请先填入 Turnstile 校验相关配置信息！",
			})
			return
		}
	}
	err = model.UpdateOption(option.Key, option.Value)
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

// BatchUpdateModelRatio 批量调价：把一组模型的倍率合并进 ModelRatio / CompletionRatio。
// 请求体：{"models": ["a","b"], "input": 1.25, "output": 10.0, "group": "vip"(可选，登记为 model(group) 复合键)}
func BatchUpdateModelRatio(c *gin.Context) {
	var body struct {
		Models []string `json:"models"`
		Input  float64  `json:"input"`
		Output float64  `json:"output"`
		Group  string   `json:"group"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	if len(body.Models) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "模型列表不能为空",
		})
		return
	}
	if body.Input <= 0 || body.Output < 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "倍率必须为非负数值，且输入倍率大于 0",
		})
		return
	}
	completionRatio := 1.0
	if body.Input > 0 {
		completionRatio = body.Output / body.Input
	}

	// 读取当前倍率表并合并（不影响其他模型键）
	modelRatioMap := map[string]float64{}
	if err := json.Unmarshal([]byte(billingratio.ModelRatio2JSONString()), &modelRatioMap); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	completionRatioMap := map[string]float64{}
	if err := json.Unmarshal([]byte(billingratio.CompletionRatio2JSONString()), &completionRatioMap); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	updated := 0
	for _, name := range body.Models {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		key := name
		if body.Group != "" {
			key = fmt.Sprintf("%s(%s)", name, body.Group)
		}
		modelRatioMap[key] = body.Input
		completionRatioMap[key] = completionRatio
		updated++
	}

	modelRatioJSON, err := json.Marshal(modelRatioMap)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if err := model.UpdateOption("ModelRatio", string(modelRatioJSON)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	completionRatioJSON, err := json.Marshal(completionRatioMap)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if err := model.UpdateOption("CompletionRatio", string(completionRatioJSON)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    updated,
	})
	return
}
