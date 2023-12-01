package middleware

import (
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ModelRequestInterface interface {
	GetModel() string
	SetModel(string)
}

type ModelRequest struct {
	Model string `json:"model"`
}

func (m *ModelRequest) GetModel() string {
	return m.Model
}

func (m *ModelRequest) SetModel(model string) {
	m.Model = model
}

type ModelFormRequest struct {
	Model string `form:"model"`
}

func (m *ModelFormRequest) GetModel() string {
	return m.Model
}

func (m *ModelFormRequest) SetModel(model string) {
	m.Model = model
}

func Distribute() func(c *gin.Context) {
	return func(c *gin.Context) {
		userId := c.GetInt("id")
		userGroup, _ := model.CacheGetUserGroup(userId)
		c.Set("group", userGroup)
		var channel *model.Channel
		channelId, ok := c.Get("channelId")
		if ok {
			id, err := strconv.Atoi(channelId.(string))
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			channel, err = model.GetChannelById(id, true)
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			if channel.Status != common.ChannelStatusEnabled {
				abortWithMessage(c, http.StatusForbidden, "该渠道已被禁用")
				return
			}
		} else {
			// Select a channel for the user
			modelRequest := getModelRequest(c)
			err := common.UnmarshalBodyReusable(c, modelRequest)
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的请求")
				return
			}

			if strings.HasPrefix(c.Request.URL.Path, "/v1/moderations") {
				if modelRequest.GetModel() == "" {
					modelRequest.SetModel("text-moderation-stable")
				}
			}
			if strings.HasSuffix(c.Request.URL.Path, "embeddings") {
				if modelRequest.GetModel() == "" {
					modelRequest.SetModel(c.Param("model"))
				}
			}
			if strings.HasPrefix(c.Request.URL.Path, "/v1/images/generations") {
				if modelRequest.GetModel() == "" {
					modelRequest.SetModel("dall-e-2")
				}
			}
			if strings.HasPrefix(c.Request.URL.Path, "/v1/audio/transcriptions") || strings.HasPrefix(c.Request.URL.Path, "/v1/audio/translations") {
				if modelRequest.GetModel() == "" {
					modelRequest.SetModel("whisper-1")
				}
			}
			channel, err = model.CacheGetRandomSatisfiedChannel(userGroup, modelRequest.GetModel())
			if err != nil {
				message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", userGroup, modelRequest.GetModel())
				if channel != nil {
					common.SysError(fmt.Sprintf("渠道不存在：%d", channel.Id))
					message = "数据库一致性已被破坏，请联系管理员"
				}
				abortWithMessage(c, http.StatusServiceUnavailable, message)
				return
			}
		}
		c.Set("channel", channel.Type)
		c.Set("channel_id", channel.Id)
		c.Set("channel_name", channel.Name)
		c.Set("model_mapping", channel.GetModelMapping())
		c.Set("api_key", channel.Key)
		// c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.Key))
		c.Set("base_url", channel.GetBaseURL())
		switch channel.Type {
		case common.ChannelTypeAzure:
			c.Set("api_version", channel.Other)
		case common.ChannelTypeXunfei:
			c.Set("api_version", channel.Other)
		case common.ChannelTypeAIProxyLibrary:
			c.Set("library_id", channel.Other)
		}
		c.Next()
	}
}

func getModelRequest(c *gin.Context) (modelRequest ModelRequestInterface) {
	contentType := c.Request.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		modelRequest = &ModelRequest{}
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		modelRequest = &ModelFormRequest{}
	}

	return
}
