package relay

import (
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/model"
	"one-api/relay/relay_util"
	"one-api/types"
	"sort"

	"github.com/gin-gonic/gin"
)

// https://platform.openai.com/docs/api-reference/models/list

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

type ModelPrice struct {
	Type   string `json:"type"`
	Input  string `json:"input"`
	Output string `json:"output"`
}
type OpenAIModels struct {
	Id         string                   `json:"id"`
	Object     string                   `json:"object"`
	Created    int                      `json:"created"`
	OwnedBy    *string                  `json:"owned_by"`
	Permission *[]OpenAIModelPermission `json:"permission"`
	Root       *string                  `json:"root"`
	Parent     *string                  `json:"parent"`
	Price      *ModelPrice              `json:"price"`
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
		c.JSON(200, gin.H{
			"object": "list",
			"data":   []string{},
		})
		return
	}
	sort.Strings(models)

	var groupOpenAIModels []*OpenAIModels
	for _, modelName := range models {
		groupOpenAIModels = append(groupOpenAIModels, getOpenAIModelWithName(modelName))
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
	prices := relay_util.PricingInstance.GetAllPrices()
	var openAIModels []OpenAIModels
	for modelId, price := range prices {
		openAIModels = append(openAIModels, OpenAIModels{
			Id:         modelId,
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    getModelOwnedBy(price.ChannelType),
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
	modelName := c.Param("model")
	openaiModel := getOpenAIModelWithName(modelName)
	if *openaiModel.OwnedBy != relay_util.UnknownOwnedBy {
		c.JSON(200, openaiModel)
	} else {
		openAIError := types.OpenAIError{
			Message: fmt.Sprintf("The model '%s' does not exist", modelName),
			Type:    "invalid_request_error",
			Param:   "model",
			Code:    "model_not_found",
		}
		c.JSON(200, gin.H{
			"error": openAIError,
		})
	}
}

func getModelOwnedBy(channelType int) (ownedBy *string) {
	if ownedByName, ok := relay_util.ModelOwnedBy[channelType]; ok {
		return &ownedByName
	}

	return &relay_util.UnknownOwnedBy
}

func getOpenAIModelWithName(modelName string) *OpenAIModels {
	price := relay_util.PricingInstance.GetPrice(modelName)

	return &OpenAIModels{
		Id:         modelName,
		Object:     "model",
		Created:    1677649963,
		OwnedBy:    getModelOwnedBy(price.ChannelType),
		Permission: nil,
		Root:       nil,
		Parent:     nil,
	}
}

func GetModelOwnedBy(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    relay_util.ModelOwnedBy,
	})
}
