package controller

import (
	"fmt"
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
	openAIModels = []OpenAIModels{
		{
			Id:         "gpt-3.5-turbo",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-3.5-turbo",
			Parent:     nil,
		},
		{
			Id:         "gpt-3.5-turbo-0301",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-3.5-turbo-0301",
			Parent:     nil,
		},
		{
			Id:         "gpt-4",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-4",
			Parent:     nil,
		},
		{
			Id:         "gpt-4-0314",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-4-0314",
			Parent:     nil,
		},
		{
			Id:         "gpt-4-32k",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-4-32k",
			Parent:     nil,
		},
		{
			Id:         "gpt-4-32k-0314",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "gpt-4-32k-0314",
			Parent:     nil,
		},
		{
			Id:         "text-embedding-ada-002",
			Object:     "model",
			Created:    1677649963,
			OwnedBy:    "openai",
			Permission: permission,
			Root:       "text-embedding-ada-002",
			Parent:     nil,
		},
	}
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
		openAIError := OpenAIError{
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
