package controller

import (
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/model"
	"one-api/types"
	"sort"

	"github.com/gin-gonic/gin"
)

// https://platform.openai.com/docs/api-reference/models/list

var unknownOwnedBy = "未知"

type OpenAIModelPermission struct {
	Id                 string  `json:"id"`
	Object             string  `json:"object"`
	Created            int     `json:"created"`
	AllowCreateEngine  bool    `json:"allow_create_engine"`
	AllowSampling      bool    `json:"allow_sampling"`
	AllowLogprobs      bool    `json:"allow_logprobs"`
	AllowSearchIndices bool    `json:"allow_search_indices"`
	AllowView          bool    `json:"allow_view"`
	AllowFineTuning    bool    `json:"allow_fine_tuning"`
	Organization       string  `json:"organization"`
	Group              *string `json:"group"`
	IsBlocking         bool    `json:"is_blocking"`
}

type OpenAIModels struct {
	Id         string                   `json:"id"`
	Object     string                   `json:"object"`
	Created    int                      `json:"created"`
	OwnedBy    *string                  `json:"owned_by"`
	Permission *[]OpenAIModelPermission `json:"permission"`
	Root       *string                  `json:"root"`
	Parent     *string                  `json:"parent"`
}

var modelOwnedBy map[int]string

func init() {
	modelOwnedBy = map[int]string{
		common.ChannelTypeOpenAI:    "OpenAI",
		common.ChannelTypeAnthropic: "Anthropic",
		common.ChannelTypeBaidu:     "Baidu",
		common.ChannelTypePaLM:      "Google PaLM",
		common.ChannelTypeGemini:    "Google Gemini",
		common.ChannelTypeZhipu:     "Zhipu",
		common.ChannelTypeAli:       "Ali",
		common.ChannelTypeXunfei:    "Xunfei",
		common.ChannelType360:       "360",
		common.ChannelTypeTencent:   "Tencent",
		common.ChannelTypeBaichuan:  "Baichuan",
		common.ChannelTypeMiniMax:   "MiniMax",
		common.ChannelTypeDeepseek:  "Deepseek",
		common.ChannelTypeMoonshot:  "Moonshot",
		common.ChannelTypeMistral:   "Mistral",
		common.ChannelTypeGroq:      "Groq",
	}
}

func ListModels(c *gin.Context) {
	groupName := c.GetString("group")
	if groupName == "" {
		id := c.GetInt("id")
		user, err := model.GetUserById(id, false)
		if err != nil {
			common.AbortWithMessage(c, http.StatusServiceUnavailable, err.Error())
			return
		}
		groupName = user.Group
	}

	models, err := model.ChannelGroup.GetGroupModels(groupName)
	if err != nil {
		common.AbortWithMessage(c, http.StatusServiceUnavailable, err.Error())
		return
	}
	sort.Strings(models)

	groupOpenAIModels := make([]OpenAIModels, 0, len(models))
	for _, modelId := range models {
		groupOpenAIModels = append(groupOpenAIModels, OpenAIModels{
			Id:         modelId,
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    getModelOwnedBy(modelId),
			Permission: nil,
			Root:       nil,
			Parent:     nil,
		})
	}

	// 根据 OwnedBy 排序
	sort.Slice(groupOpenAIModels, func(i, j int) bool {
		if groupOpenAIModels[i].OwnedBy == nil {
			return true // 假设 nil 值小于任何非 nil 值
		}
		if groupOpenAIModels[j].OwnedBy == nil {
			return false // 假设任何非 nil 值大于 nil 值
		}
		return *groupOpenAIModels[i].OwnedBy < *groupOpenAIModels[j].OwnedBy
	})

	c.JSON(200, gin.H{
		"object": "list",
		"data":   groupOpenAIModels,
	})
}

func ListModelsForAdmin(c *gin.Context) {
	openAIModels := make([]OpenAIModels, 0, len(common.ModelRatio))
	for modelId := range common.ModelRatio {
		openAIModels = append(openAIModels, OpenAIModels{
			Id:         modelId,
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    getModelOwnedBy(modelId),
			Permission: nil,
			Root:       nil,
			Parent:     nil,
		})
	}
	// 根据 OwnedBy 排序
	sort.Slice(openAIModels, func(i, j int) bool {
		if openAIModels[i].OwnedBy == nil {
			return true // 假设 nil 值小于任何非 nil 值
		}
		if openAIModels[j].OwnedBy == nil {
			return false // 假设任何非 nil 值大于 nil 值
		}
		return *openAIModels[i].OwnedBy < *openAIModels[j].OwnedBy
	})

	c.JSON(200, gin.H{
		"object": "list",
		"data":   openAIModels,
	})
}

func RetrieveModel(c *gin.Context) {
	modelId := c.Param("model")
	ownedByName := getModelOwnedBy(modelId)
	if *ownedByName != unknownOwnedBy {
		c.JSON(200, OpenAIModels{
			Id:         modelId,
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    ownedByName,
			Permission: nil,
			Root:       nil,
			Parent:     nil,
		})
	} else {
		openAIError := types.OpenAIError{
			Message: fmt.Sprintf("The model '%s' does not exist", modelId),
			Type:    "invalid_request_error",
			Param:   "model",
			Code:    "model_not_found",
		}
		c.JSON(200, gin.H{
			"error": openAIError,
		})
	}
}

func getModelOwnedBy(modelId string) (ownedBy *string) {
	if modelType, ok := common.ModelTypes[modelId]; ok {
		if ownedByName, ok := modelOwnedBy[modelType.Type]; ok {
			return &ownedByName
		}
	}

	return &unknownOwnedBy
}
