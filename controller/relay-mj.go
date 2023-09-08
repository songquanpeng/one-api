package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strings"

	"github.com/gin-gonic/gin"
)

type Midjourney struct {
	MjId        string `json:"id"`
	Action      string `json:"action"`
	Prompt      string `json:"prompt"`
	PromptEn    string `json:"promptEn"`
	Description string `json:"description"`
	State       string `json:"state"`
	SubmitTime  int64  `json:"submitTime"`
	StartTime   int64  `json:"startTime"`
	FinishTime  int64  `json:"finishTime"`
	ImageUrl    string `json:"imageUrl"`
	Status      string `json:"status"`
	Progress    string `json:"progress"`
	FailReason  string `json:"failReason"`
}

func relayMidjourneyNotify(c *gin.Context) *MidjourneyResponse {
	var midjRequest Midjourney
	err := common.UnmarshalBodyReusable(c, &midjRequest)
	if err != nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "bind_request_body_failed",
			Properties:  nil,
			Result:      "",
		}
	}
	midjourneyTask := model.GetByMJId(midjRequest.MjId)
	if midjourneyTask == nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "midjourney_task_not_found",
			Properties:  nil,
			Result:      "",
		}
	}
	midjourneyTask.Progress = midjRequest.Progress
	midjourneyTask.PromptEn = midjRequest.PromptEn
	midjourneyTask.State = midjRequest.State
	midjourneyTask.SubmitTime = midjRequest.SubmitTime
	midjourneyTask.StartTime = midjRequest.StartTime
	midjourneyTask.FinishTime = midjRequest.FinishTime
	midjourneyTask.ImageUrl = midjRequest.ImageUrl
	midjourneyTask.Status = midjRequest.Status
	midjourneyTask.FailReason = midjRequest.FailReason
	err = midjourneyTask.Update()
	if err != nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "update_midjourney_task_failed",
		}
	}

	return nil
}

func relayMidjourneyTask(c *gin.Context, relayMode int) *MidjourneyResponse {
	taskId := c.Param("id")
	originTask := model.GetByMJId(taskId)
	if originTask == nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "task_no_found",
		}
	}
	var midjourneyTask Midjourney
	midjourneyTask.MjId = originTask.MjId
	midjourneyTask.Progress = originTask.Progress
	midjourneyTask.PromptEn = originTask.PromptEn
	midjourneyTask.State = originTask.State
	midjourneyTask.SubmitTime = originTask.SubmitTime
	midjourneyTask.StartTime = originTask.StartTime
	midjourneyTask.FinishTime = originTask.FinishTime
	midjourneyTask.ImageUrl = originTask.ImageUrl
	midjourneyTask.Status = originTask.Status
	midjourneyTask.FailReason = originTask.FailReason
	midjourneyTask.Action = originTask.Action
	midjourneyTask.Description = originTask.Description
	midjourneyTask.Prompt = originTask.Prompt
	jsonMap, err := json.Marshal(midjourneyTask)
	if err != nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "unmarshal_response_body_failed",
		}
	}
	_, err = io.Copy(c.Writer, bytes.NewBuffer(jsonMap))
	if err != nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "copy_response_body_failed",
		}
	}
	return nil
}

func relayMidjourneySubmit(c *gin.Context, relayMode int) *MidjourneyResponse {
	imageModel := "midjourney"

	tokenId := c.GetInt("token_id")
	channelType := c.GetInt("channel")
	userId := c.GetInt("id")
	consumeQuota := c.GetBool("consume_quota")
	group := c.GetString("group")

	var midjRequest MidjourneyRequest
	if consumeQuota {
		err := common.UnmarshalBodyReusable(c, &midjRequest)
		if err != nil {
			return &MidjourneyResponse{
				Code:        4,
				Description: "bind_request_body_failed",
			}
		}
	}
	if relayMode == RelayModeMidjourneyImagine {
		if midjRequest.Prompt == "" {
			return &MidjourneyResponse{
				Code:        4,
				Description: "prompt_is_required",
			}
		}
		midjRequest.Action = "IMAGINE"
	} else if midjRequest.TaskId != "" {
		originTask := model.GetByMJId(midjRequest.TaskId)
		if originTask == nil {
			return &MidjourneyResponse{
				Code:        4,
				Description: "task_no_found",
			}
		} else if originTask.Action == "UPSCALE" {
			//return errorWrapper(errors.New("upscale task can not be change"), "request_params_error", http.StatusBadRequest).
			return &MidjourneyResponse{
				Code:        4,
				Description: "upscale_task_can_not_be_change",
			}
		} else if originTask.Status != "SUCCESS" {
			return &MidjourneyResponse{
				Code:        4,
				Description: "task_status_is_not_success",
			}

		}
		midjRequest.Prompt = originTask.Prompt
	} else if relayMode == RelayModeMidjourneyChange {
		if midjRequest.TaskId == "" {
			return &MidjourneyResponse{
				Code:        4,
				Description: "taskId_is_required",
			}
		} else if midjRequest.Action == "" {
			return &MidjourneyResponse{
				Code:        4,
				Description: "action_is_required",
			}
		} else if midjRequest.Index == 0 {
			return &MidjourneyResponse{
				Code:        4,
				Description: "index_can_only_be_1_2_3_4",
			}
		}
	}

	// map model name
	modelMapping := c.GetString("model_mapping")
	isModelMapped := false
	if modelMapping != "" {
		modelMap := make(map[string]string)
		err := json.Unmarshal([]byte(modelMapping), &modelMap)
		if err != nil {
			//return errorWrapper(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
			return &MidjourneyResponse{
				Code:        4,
				Description: "unmarshal_model_mapping_failed",
			}
		}
		if modelMap[imageModel] != "" {
			imageModel = modelMap[imageModel]
			isModelMapped = true
		}
	}

	baseURL := common.ChannelBaseURLs[channelType]
	requestURL := c.Request.URL.String()

	if c.GetString("base_url") != "" {
		baseURL = c.GetString("base_url")
	}

	//midjRequest.NotifyHook = "http://127.0.0.1:3000/mj/notify"

	fullRequestURL := fmt.Sprintf("%s%s", baseURL, requestURL)

	var requestBody io.Reader
	if isModelMapped {
		jsonStr, err := json.Marshal(midjRequest)
		if err != nil {
			return &MidjourneyResponse{
				Code:        4,
				Description: "marshal_text_request_failed",
			}
		}
		requestBody = bytes.NewBuffer(jsonStr)
	} else {
		requestBody = c.Request.Body
	}

	modelRatio := common.GetModelRatio(imageModel)
	groupRatio := common.GetGroupRatio(group)
	ratio := modelRatio * groupRatio
	userQuota, err := model.CacheGetUserQuota(userId)

	sizeRatio := 1.0

	quota := int(ratio * sizeRatio * 1000)

	if consumeQuota && userQuota-quota < 0 {
		return &MidjourneyResponse{
			Code:        4,
			Description: "quota_not_enough",
		}
	}

	req, err := http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
	if err != nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "create_request_failed",
		}
	}
	//req.Header.Set("Authorization", c.Request.Header.Get("Authorization"))

	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Accept", c.Request.Header.Get("Accept"))
	//mjToken := ""
	//if c.Request.Header.Get("Authorization") != "" {
	//	mjToken = strings.Split(c.Request.Header.Get("Authorization"), " ")[1]
	//}
	req.Header.Set("mj-api-secret", strings.Split(c.Request.Header.Get("Authorization"), " ")[1])
	// print request header
	log.Printf("request header: %s", req.Header)
	log.Printf("request body: %s", midjRequest.Prompt)

	resp, err := httpClient.Do(req)
	if err != nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "do_request_failed",
		}
	}

	err = req.Body.Close()
	if err != nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "close_request_body_failed",
		}
	}
	err = c.Request.Body.Close()
	if err != nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "close_request_body_failed",
		}
	}
	var midjResponse MidjourneyResponse

	defer func() {
		if consumeQuota {
			err := model.PostConsumeTokenQuota(tokenId, quota)
			if err != nil {
				common.SysError("error consuming token remain quota: " + err.Error())
			}
			err = model.CacheUpdateUserQuota(userId)
			if err != nil {
				common.SysError("error update user quota cache: " + err.Error())
			}
			if quota != 0 {
				tokenName := c.GetString("token_name")
				logContent := fmt.Sprintf("模型倍率 %.2f，分组倍率 %.2f", modelRatio, groupRatio)
				model.RecordConsumeLog(userId, 0, 0, imageModel, tokenName, quota, logContent, tokenId)
				model.UpdateUserUsedQuotaAndRequestCount(userId, quota)
				channelId := c.GetInt("channel_id")
				model.UpdateChannelUsedQuota(channelId, quota)
			}
		}
	}()

	//if consumeQuota {
	//
	//}
	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "read_response_body_failed",
		}
	}
	err = resp.Body.Close()
	if err != nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "close_response_body_failed",
		}
	}

	err = json.Unmarshal(responseBody, &midjResponse)
	log.Printf("responseBody: %s", string(responseBody))
	log.Printf("midjResponse: %v", midjResponse)
	if resp.StatusCode != 200 {
		return &MidjourneyResponse{
			Code:        4,
			Description: "fail_to_fetch_midjourney",
		}
	}
	if err != nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "unmarshal_response_body_failed",
		}
	}
	if midjResponse.Code == 24 || midjResponse.Code == 21 || midjResponse.Code == 4 {
		consumeQuota = false
	}

	midjourneyTask := &model.Midjourney{
		UserId:      userId,
		Code:        midjResponse.Code,
		Action:      midjRequest.Action,
		MjId:        midjResponse.Result,
		Prompt:      midjRequest.Prompt,
		PromptEn:    "",
		Description: midjResponse.Description,
		State:       "",
		SubmitTime:  0,
		StartTime:   0,
		FinishTime:  0,
		ImageUrl:    "",
		Status:      "",
		Progress:    "0%",
		FailReason:  "",
		ChannelId:   c.GetInt("channel_id"),
	}
	if midjResponse.Code == 4 || midjResponse.Code == 24 {
		midjourneyTask.FailReason = midjResponse.Description
	}
	err = midjourneyTask.Insert()
	if err != nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "insert_midjourney_task_failed",
		}
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))

	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)

	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "copy_response_body_failed",
		}
	}
	err = resp.Body.Close()
	if err != nil {
		return &MidjourneyResponse{
			Code:        4,
			Description: "close_response_body_failed",
		}
	}
	return nil
}
