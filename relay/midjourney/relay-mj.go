// Author: Calcium-Ion
// GitHub: https://github.com/Calcium-Ion/new-api
// Path: relay/relay-mj.go
package midjourney

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/controller"
	"one-api/model"
	provider "one-api/providers/midjourney"
	"one-api/relay"
	"one-api/relay/relay_util"
	"one-api/types"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func RelayMidjourneyImage(c *gin.Context) {
	taskId := c.Param("id")
	midjourneyTask := model.GetByOnlyMJId(taskId)
	if midjourneyTask == nil {
		c.JSON(400, gin.H{
			"error": "midjourney_task_not_found",
		})
		return
	}
	resp, err := http.Get(midjourneyTask.ImageUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "http_get_image_failed",
		})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{
			"error": string(responseBody),
		})
		return
	}
	// 从Content-Type头获取MIME类型
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// 如果无法确定内容类型，则默认为jpeg
		contentType = "image/jpeg"
	}
	// 设置响应的内容类型
	c.Writer.Header().Set("Content-Type", contentType)
	// 将图片流式传输到响应体
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		log.Println("Failed to stream image:", err)
	}
}

func RelayMidjourneyNotify(c *gin.Context) *provider.MidjourneyResponse {
	var midjRequest provider.MidjourneyDto
	err := common.UnmarshalBodyReusable(c, &midjRequest)
	if err != nil {
		return &provider.MidjourneyResponse{
			Code:        4,
			Description: "bind_request_body_failed",
			Properties:  nil,
			Result:      "",
		}
	}
	midjourneyTask := model.GetByOnlyMJId(midjRequest.MjId)
	if midjourneyTask == nil {
		return &provider.MidjourneyResponse{
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
		return &provider.MidjourneyResponse{
			Code:        4,
			Description: "update_midjourney_task_failed",
		}
	}

	return nil
}

func coverMidjourneyTaskDto(originTask *model.Midjourney) (midjourneyTask provider.MidjourneyDto) {
	midjourneyTask.MjId = originTask.MjId
	midjourneyTask.Progress = originTask.Progress
	midjourneyTask.PromptEn = originTask.PromptEn
	midjourneyTask.State = originTask.State
	midjourneyTask.SubmitTime = originTask.SubmitTime
	midjourneyTask.StartTime = originTask.StartTime
	midjourneyTask.FinishTime = originTask.FinishTime
	midjourneyTask.ImageUrl = ""
	if originTask.ImageUrl != "" {
		midjourneyTask.ImageUrl = config.ServerAddress + "/mj/image/" + originTask.MjId
		if originTask.Status != "SUCCESS" {
			midjourneyTask.ImageUrl += "?rand=" + strconv.FormatInt(time.Now().UnixNano(), 10)
		}
	}
	midjourneyTask.Status = originTask.Status
	midjourneyTask.FailReason = originTask.FailReason
	midjourneyTask.Action = originTask.Action
	midjourneyTask.Description = originTask.Description
	midjourneyTask.Prompt = originTask.Prompt
	if originTask.Buttons != "" {
		var buttons []provider.ActionButton
		err := json.Unmarshal([]byte(originTask.Buttons), &buttons)
		if err == nil {
			midjourneyTask.Buttons = buttons
		}
	}
	if originTask.Properties != "" {
		var properties provider.Properties
		err := json.Unmarshal([]byte(originTask.Properties), &properties)
		if err == nil {
			midjourneyTask.Properties = &properties
		}
	}
	return
}

func RelaySwapFace(c *gin.Context) *provider.MidjourneyResponse {
	mjProvider, errWithMJ := getMJProviderWithRequest(c, provider.RelayModeMidjourneySwapFace, nil)
	if errWithMJ != nil {
		return errWithMJ
	}

	startTime := time.Now().UnixNano() / int64(time.Millisecond)
	userId := c.GetInt("id")
	var swapFaceRequest provider.SwapFaceRequest
	err := common.UnmarshalBodyReusable(c, &swapFaceRequest)
	if err != nil {
		return provider.MidjourneyErrorWrapper(provider.MjRequestError, "bind_request_body_failed")
	}
	if swapFaceRequest.SourceBase64 == "" || swapFaceRequest.TargetBase64 == "" {
		return provider.MidjourneyErrorWrapper(provider.MjRequestError, "sour_base64_and_target_base64_is_required")
	}

	quotaInstance, errWithOA := getQuota(c, provider.MjActionSwapFace)
	if errWithOA != nil {
		return &provider.MidjourneyResponse{
			Code:        4,
			Description: errWithOA.Message,
		}
	}
	requestURL := getMjRequestPath(c.Request.URL.String())
	mjResp, _, err := mjProvider.Send(60, requestURL)
	if err != nil {
		quotaInstance.Undo(c)
		return &mjResp.Response
	}
	if mjResp.StatusCode == 200 && mjResp.Response.Code == 1 {
		quotaInstance.Consume(c, &types.Usage{CompletionTokens: 0, PromptTokens: 1000, TotalTokens: 1000})
	} else {
		quotaInstance.Undo(c)
	}

	quota := int(quotaInstance.GetInputRatio() * 1000)

	midjResponse := &mjResp.Response
	midjourneyTask := &model.Midjourney{
		UserId:      userId,
		Code:        midjResponse.Code,
		Action:      provider.MjActionSwapFace,
		MjId:        midjResponse.Result,
		Prompt:      "InsightFace",
		PromptEn:    "",
		Description: midjResponse.Description,
		State:       "",
		SubmitTime:  startTime,
		StartTime:   time.Now().UnixNano() / int64(time.Millisecond),
		FinishTime:  0,
		ImageUrl:    "",
		Status:      "",
		Progress:    "0%",
		FailReason:  "",
		ChannelId:   c.GetInt("channel_id"),
		Quota:       quota,
	}
	err = midjourneyTask.Insert()
	if err != nil {
		return provider.MidjourneyErrorWrapper(provider.MjRequestError, "insert_midjourney_task_failed")
	}
	// 开始激活任务
	controller.ActivateUpdateMidjourneyTaskBulk()

	c.Writer.WriteHeader(mjResp.StatusCode)
	respBody, err := json.Marshal(midjResponse)
	if err != nil {
		return provider.MidjourneyErrorWrapper(provider.MjRequestError, "unmarshal_response_body_failed")
	}
	_, err = io.Copy(c.Writer, bytes.NewBuffer(respBody))
	if err != nil {
		return provider.MidjourneyErrorWrapper(provider.MjRequestError, "copy_response_body_failed")
	}
	return nil
}

func RelayMidjourneyTaskImageSeed(c *gin.Context) *provider.MidjourneyResponse {
	taskId := c.Param("id")
	userId := c.GetInt("id")
	originTask := model.GetByMJId(userId, taskId)
	if originTask == nil {
		return provider.MidjourneyErrorWrapper(provider.MjRequestError, "task_no_found")
	}

	mjProvider, errWithMJ := getMJProviderWithChannelId(c, originTask.ChannelId)
	if errWithMJ != nil {
		return errWithMJ
	}

	requestURL := getMjRequestPath(c.Request.URL.String())
	midjResponseWithStatus, _, err := mjProvider.Send(30, requestURL)
	if err != nil {
		return &midjResponseWithStatus.Response
	}
	midjResponse := &midjResponseWithStatus.Response
	c.Writer.WriteHeader(midjResponseWithStatus.StatusCode)
	respBody, err := json.Marshal(midjResponse)
	if err != nil {
		return provider.MidjourneyErrorWrapper(provider.MjRequestError, "unmarshal_response_body_failed")
	}
	_, err = io.Copy(c.Writer, bytes.NewBuffer(respBody))
	if err != nil {
		return provider.MidjourneyErrorWrapper(provider.MjRequestError, "copy_response_body_failed")
	}
	return nil
}

func RelayMidjourneyTask(c *gin.Context, relayMode int) *provider.MidjourneyResponse {
	userId := c.GetInt("id")
	var err error
	var respBody []byte
	switch relayMode {
	case provider.RelayModeMidjourneyTaskFetch:
		taskId := c.Param("id")
		originTask := model.GetByMJId(userId, taskId)
		if originTask == nil {
			return &provider.MidjourneyResponse{
				Code:        4,
				Description: "task_no_found",
			}
		}
		midjourneyTask := coverMidjourneyTaskDto(originTask)
		respBody, err = json.Marshal(midjourneyTask)
		if err != nil {
			return &provider.MidjourneyResponse{
				Code:        4,
				Description: "unmarshal_response_body_failed",
			}
		}
	case provider.RelayModeMidjourneyTaskFetchByCondition:
		var condition = struct {
			IDs []string `json:"ids"`
		}{}
		err = c.BindJSON(&condition)
		if err != nil {
			return &provider.MidjourneyResponse{
				Code:        4,
				Description: "do_request_failed",
			}
		}
		var tasks []provider.MidjourneyDto
		if len(condition.IDs) != 0 {
			originTasks := model.GetByMJIds(userId, condition.IDs)
			for _, originTask := range originTasks {
				midjourneyTask := coverMidjourneyTaskDto(originTask)
				tasks = append(tasks, midjourneyTask)
			}
		}
		if tasks == nil {
			tasks = make([]provider.MidjourneyDto, 0)
		}
		respBody, err = json.Marshal(tasks)
		if err != nil {
			return &provider.MidjourneyResponse{
				Code:        4,
				Description: "unmarshal_response_body_failed",
			}
		}
	}

	c.Writer.Header().Set("Content-Type", "application/json")

	_, err = io.Copy(c.Writer, bytes.NewBuffer(respBody))
	if err != nil {
		return &provider.MidjourneyResponse{
			Code:        4,
			Description: "copy_response_body_failed",
		}
	}
	return nil
}

func RelayMidjourneySubmit(c *gin.Context, relayMode int) *provider.MidjourneyResponse {
	userId := c.GetInt("id")
	consumeQuota := true
	var midjRequest provider.MidjourneyRequest
	err := common.UnmarshalBodyReusable(c, &midjRequest)
	if err != nil {
		return provider.MidjourneyErrorWrapper(provider.MjRequestError, "bind_request_body_failed")
	}

	mjProvider, errWithMJ := getMJProviderWithRequest(c, relayMode, &midjRequest)
	if errWithMJ != nil {
		return errWithMJ
	}

	if relayMode == provider.RelayModeMidjourneyAction { // midjourney plus，需要从customId中获取任务信息
		mjErr := CoverPlusActionToNormalAction(&midjRequest)
		if mjErr != nil {
			return mjErr
		}
		relayMode = provider.RelayModeMidjourneyChange
	}

	if relayMode == provider.RelayModeMidjourneyImagine { //绘画任务，此类任务可重复
		if midjRequest.Prompt == "" {
			return provider.MidjourneyErrorWrapper(provider.MjRequestError, "prompt_is_required")
		}
		midjRequest.Action = provider.MjActionImagine
	} else if relayMode == provider.RelayModeMidjourneyDescribe { //按图生文任务，此类任务可重复
		midjRequest.Action = provider.MjActionDescribe
	} else if relayMode == provider.RelayModeMidjourneyShorten { //缩短任务，此类任务可重复，plus only
		midjRequest.Action = provider.MjActionShorten
	} else if relayMode == provider.RelayModeMidjourneyBlend { //绘画任务，此类任务可重复
		midjRequest.Action = provider.MjActionBlend
	} else if midjRequest.TaskId != "" { //放大、变换任务，此类任务，如果重复且已有结果，远端api会直接返回最终结果
		mjId := ""
		if relayMode == provider.RelayModeMidjourneyChange {
			if midjRequest.TaskId == "" {
				return provider.MidjourneyErrorWrapper(provider.MjRequestError, "task_id_is_required")
			} else if midjRequest.Action == "" {
				return provider.MidjourneyErrorWrapper(provider.MjRequestError, "action_is_required")
			} else if midjRequest.Index == 0 {
				return provider.MidjourneyErrorWrapper(provider.MjRequestError, "index_is_required")
			}
			//action = midjRequest.Action
			mjId = midjRequest.TaskId
		} else if relayMode == provider.RelayModeMidjourneySimpleChange {
			if midjRequest.Content == "" {
				return provider.MidjourneyErrorWrapper(provider.MjRequestError, "content_is_required")
			}
			params := ConvertSimpleChangeParams(midjRequest.Content)
			if params == nil {
				return provider.MidjourneyErrorWrapper(provider.MjRequestError, "content_parse_failed")
			}
			mjId = params.TaskId
			midjRequest.Action = params.Action
		} else if relayMode == provider.RelayModeMidjourneyModal {
			//if midjRequest.MaskBase64 == "" {
			//	return provider.MidjourneyErrorWrapper(provider.MjRequestError, "mask_base64_is_required")
			//}
			mjId = midjRequest.TaskId
			midjRequest.Action = provider.MjActionModal
		}

		originTask := model.GetByMJId(userId, mjId)
		if originTask == nil {
			return provider.MidjourneyErrorWrapper(provider.MjRequestError, "task_not_found")
		} else if originTask.Status != "SUCCESS" && relayMode != provider.RelayModeMidjourneyModal {
			return provider.MidjourneyErrorWrapper(provider.MjRequestError, "task_status_not_success")
		} else { //原任务的Status=SUCCESS，则可以做放大UPSCALE、变换VARIATION等动作，此时必须使用原来的请求地址才能正确处理
			mjProvider, errWithMJ = getMJProviderWithChannelId(c, originTask.ChannelId)
			if errWithMJ != nil {
				return errWithMJ
			}
			log.Printf("检测到此操作为放大、变换、重绘，获取原channel信息: %d", originTask.ChannelId)
		}
		midjRequest.Prompt = originTask.Prompt

		//if channelType == common.ChannelTypeMidjourneyPlus {
		//	// plus
		//} else {
		//	// 普通版渠道
		//
		//}
	}

	if midjRequest.Action == provider.MjActionInPaint || midjRequest.Action == provider.MjActionCustomZoom {
		consumeQuota = false
	}

	//baseURL := common.ChannelBaseURLs[channelType]
	requestURL := getMjRequestPath(c.Request.URL.String())

	//midjRequest.NotifyHook = "http://127.0.0.1:3000/mj/notify"

	quotaInstance, errWithOA := getQuota(c, midjRequest.Action)
	if errWithOA != nil {
		return &provider.MidjourneyResponse{
			Code:        4,
			Description: errWithOA.Message,
		}
	}

	midjResponseWithStatus, responseBody, err := mjProvider.Send(60, requestURL)
	if err != nil {
		quotaInstance.Undo(c)
		return &midjResponseWithStatus.Response
	}

	if consumeQuota && midjResponseWithStatus.StatusCode == 200 {
		quotaInstance.Consume(c, &types.Usage{CompletionTokens: 0, PromptTokens: 1, TotalTokens: 1})
	} else {
		quotaInstance.Undo(c)
	}
	quota := int(quotaInstance.GetInputRatio() * 1000)

	midjResponse := &midjResponseWithStatus.Response

	// 文档：https://github.com/novicezk/midjourney-proxy/blob/main/docs/api.md
	//1-提交成功
	// 21-任务已存在（处理中或者有结果了） {"code":21,"description":"任务已存在","result":"0741798445574458","properties":{"status":"SUCCESS","imageUrl":"https://xxxx"}}
	// 22-排队中 {"code":22,"description":"排队中，前面还有1个任务","result":"0741798445574458","properties":{"numberOfQueues":1,"discordInstanceId":"1118138338562560102"}}
	// 23-队列已满，请稍后再试 {"code":23,"description":"队列已满，请稍后尝试","result":"14001929738841620","properties":{"discordInstanceId":"1118138338562560102"}}
	// 24-prompt包含敏感词 {"code":24,"description":"可能包含敏感词","properties":{"promptEn":"nude body","bannedWord":"nude"}}
	// other: 提交错误，description为错误描述
	midjourneyTask := &model.Midjourney{
		UserId:      userId,
		Code:        midjResponse.Code,
		Action:      midjRequest.Action,
		MjId:        midjResponse.Result,
		Prompt:      midjRequest.Prompt,
		PromptEn:    "",
		Description: midjResponse.Description,
		State:       "",
		SubmitTime:  time.Now().UnixNano() / int64(time.Millisecond),
		StartTime:   0,
		FinishTime:  0,
		ImageUrl:    "",
		Status:      "",
		Progress:    "0%",
		FailReason:  "",
		ChannelId:   c.GetInt("channel_id"),
		Quota:       quota,
	}

	if midjResponse.Code != 1 && midjResponse.Code != 21 && midjResponse.Code != 22 {
		//非1-提交成功,21-任务已存在和22-排队中，则记录错误原因
		midjourneyTask.FailReason = midjResponse.Description
		consumeQuota = false
	}

	if midjResponse.Code == 21 { //21-任务已存在（处理中或者有结果了）
		// 将 properties 转换为一个 map
		properties, ok := midjResponse.Properties.(map[string]interface{})
		if ok {
			imageUrl, ok1 := properties["imageUrl"].(string)
			status, ok2 := properties["status"].(string)
			if ok1 && ok2 {
				midjourneyTask.ImageUrl = imageUrl
				midjourneyTask.Status = status
				if status == "SUCCESS" {
					midjourneyTask.Progress = "100%"
					midjourneyTask.StartTime = time.Now().UnixNano() / int64(time.Millisecond)
					midjourneyTask.FinishTime = time.Now().UnixNano() / int64(time.Millisecond)
					midjResponse.Code = 1
				}
			}
		}
		//修改返回值
		if midjRequest.Action != provider.MjActionInPaint && midjRequest.Action != provider.MjActionCustomZoom {
			newBody := strings.Replace(string(responseBody), `"code":21`, `"code":1`, -1)
			responseBody = []byte(newBody)
		}
	}

	err = midjourneyTask.Insert()
	if err != nil {
		return &provider.MidjourneyResponse{
			Code:        4,
			Description: "insert_midjourney_task_failed",
		}
	}
	// 开始激活任务
	controller.ActivateUpdateMidjourneyTaskBulk()

	if midjResponse.Code == 22 { //22-排队中，说明任务已存在
		//修改返回值
		newBody := strings.Replace(string(responseBody), `"code":22`, `"code":1`, -1)
		responseBody = []byte(newBody)
	}

	//resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	bodyReader := io.NopCloser(bytes.NewBuffer(responseBody))

	//for k, v := range resp.Header {
	//	c.Writer.Header().Set(k, v[0])
	//}
	c.Writer.WriteHeader(midjResponseWithStatus.StatusCode)

	_, err = io.Copy(c.Writer, bodyReader)
	if err != nil {
		return &provider.MidjourneyResponse{
			Code:        4,
			Description: "copy_response_body_failed",
		}
	}
	err = bodyReader.Close()
	if err != nil {
		return &provider.MidjourneyResponse{
			Code:        4,
			Description: "close_response_body_failed",
		}
	}
	return nil
}

func getMjRequestPath(path string) string {
	requestURL := path
	if strings.Contains(requestURL, "/mj-") {
		urls := strings.Split(requestURL, "/mj/")
		if len(urls) < 2 {
			return requestURL
		}
		requestURL = "/mj/" + urls[1]
	}
	return requestURL
}

func getQuota(c *gin.Context, action string) (*relay_util.Quota, *types.OpenAIErrorWithStatusCode) {
	modelName := CoverActionToModelName(action)

	return relay_util.NewQuota(c, modelName, 1000)
}

func getMJProviderWithRequest(c *gin.Context, relayMode int, request *provider.MidjourneyRequest) (*provider.MidjourneyProvider, *provider.MidjourneyResponse) {
	midjourneyModel, mjErr, _ := GetMjRequestModel(relayMode, request)
	if mjErr != nil {
		return nil, MidjourneyErrorFromInternal(mjErr.Code, mjErr.Description)
	}
	if midjourneyModel == "" {
		return nil, MidjourneyErrorFromInternal(provider.MjErrorUnknown, "无效的请求, 无法解析模型")
	}

	return getMJProvider(c, midjourneyModel)
}

func getMJProviderWithChannelId(c *gin.Context, channelId int) (*provider.MidjourneyProvider, *provider.MidjourneyResponse) {
	c.Set("specific_channel_id", channelId)

	return getMJProvider(c, "")
}

func getMJProvider(c *gin.Context, modelName string) (*provider.MidjourneyProvider, *provider.MidjourneyResponse) {
	baseProvider, _, err := relay.GetProvider(c, modelName)
	if err != nil {
		return nil, MidjourneyErrorFromInternal(provider.MjErrorUnknown, "无法获取provider:"+err.Error())
	}

	mjProvider, ok := baseProvider.(*provider.MidjourneyProvider)
	if !ok {
		return nil, MidjourneyErrorFromInternal(provider.MjErrorUnknown, "无效的请求, 无法获取midjourney provider")
	}

	return mjProvider, nil
}
