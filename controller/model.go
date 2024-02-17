package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/helper"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
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
	Id         string                  `json:"id"`
	Object     string                  `json:"object"`
	Created    int                     `json:"created"`
	OwnedBy    string                  `json:"owned_by"`
	Permission []OpenAIModelPermission `json:"permission"`
	Root       string                  `json:"root"`
	Parent     *string                 `json:"parent"`
}

var openAIModels []OpenAIModels
var openAIModelsMap map[string]OpenAIModels

func init() {
	var permission []OpenAIModelPermission
	permission = append(permission, OpenAIModelPermission{
		Id:                 "modelperm-LwHkVFn8AcMItP432fKKDIKJ",
		Object:             "model_permission",
		Created:            1626777600,
		AllowCreateEngine:  true,
		AllowSampling:      true,
		AllowLogprobs:      true,
		AllowSearchIndices: false,
		AllowView:          true,
		AllowFineTuning:    false,
		Organization:       "*",
		Group:              nil,
		IsBlocking:         false,
	})
	// https://platform.openai.com/docs/models/model-endpoint-compatibility
	for i := 0; i < constant.APITypeDummy; i++ {
		adaptor := helper.GetAdaptor(i)
		channelName := adaptor.GetChannelName()
		modelNames := adaptor.GetModelList()
		for _, modelName := range modelNames {
			openAIModels = append(openAIModels, OpenAIModels{
				Id:         modelName,
				Object:     "model",
				Created:    1626777600,
				OwnedBy:    channelName,
				Permission: permission,
				Root:       modelName,
				Parent:     nil,
			})
		}
	}
	// extra
	openAIModels = append(openAIModels, []OpenAIModels{
		{
			Id:         "360GPT_S2_V9",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "360",
			Permission: permission,
			Root:       "360GPT_S2_V9",
			Parent:     nil,
		},
		{
			Id:         "embedding-bert-512-v1",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "360",
			Permission: permission,
			Root:       "embedding-bert-512-v1",
			Parent:     nil,
		},
		{
			Id:         "embedding_s1_v1",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "360",
			Permission: permission,
			Root:       "embedding_s1_v1",
			Parent:     nil,
		},
		{
			Id:         "semantic_similarity_s1_v1",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "360",
			Permission: permission,
			Root:       "semantic_similarity_s1_v1",
			Parent:     nil,
		},
	}...)
	openAIModelsMap = make(map[string]OpenAIModels)
	for _, model := range openAIModels {
		openAIModelsMap[model.Id] = model
	}
}

func ListModels(c *gin.Context) {
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
		Error := relaymodel.Error{
			Message: fmt.Sprintf("The model '%s' does not exist", modelId),
			Type:    "invalid_request_error",
			Param:   "model",
			Code:    "model_not_found",
		}
		c.JSON(200, gin.H{
			"error": Error,
		})
	}
}
