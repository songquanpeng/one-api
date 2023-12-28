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

var openAIModels []OpenAIModels
var openAIModelsMap map[string]OpenAIModels

func init() {
	// https://platform.openai.com/docs/models/model-endpoint-compatibility
	keys := make([]string, 0, len(common.ModelRatio))
	for k := range common.ModelRatio {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, modelId := range keys {
		openAIModels = append(openAIModels, OpenAIModels{
			Id:         modelId,
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    nil,
			Permission: nil,
			Root:       nil,
			Parent:     nil,
		})
	}

	openAIModelsMap = make(map[string]OpenAIModels)
	for _, model := range openAIModels {
		openAIModelsMap[model.Id] = model
	}
}

func ListModels(c *gin.Context) {
	groupName := c.GetString("group")

	models, err := model.CacheGetGroupModels(groupName)
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
			OwnedBy:    nil,
			Permission: nil,
			Root:       nil,
			Parent:     nil,
		})
	}

	c.JSON(200, gin.H{
		"object": "list",
		"data":   groupOpenAIModels,
	})
}

func ListModelsForAdmin(c *gin.Context) {
	c.JSON(200, gin.H{
		"object": "list",
		"data":   openAIModels,
	})
}

func RetrieveModel(c *gin.Context) {
	modelId := c.Param("model")
	if model, ok := openAIModelsMap[modelId]; ok {
		c.JSON(200, model)
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
