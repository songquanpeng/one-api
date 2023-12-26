package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"one-api/common"
	"one-api/model"
	"one-api/providers"
	providersBase "one-api/providers/base"
	"one-api/types"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func GetValidFieldName(err error, obj interface{}) string {
	getObj := reflect.TypeOf(obj)
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			if f, exist := getObj.Elem().FieldByName(e.Field()); exist {
				return f.Name
			}
		}
	}
	return err.Error()
}

func fetchChannel(c *gin.Context, modelName string) (channel *model.Channel, pass bool) {
	channelId, ok := c.Get("channelId")
	if ok {
		channel, pass = fetchChannelById(c, channelId.(int))
		if pass {
			return
		}

	}
	channel, pass = fetchChannelByModel(c, modelName)
	if pass {
		return
	}

	return
}

func fetchChannelById(c *gin.Context, channelId any) (*model.Channel, bool) {
	id, err := strconv.Atoi(channelId.(string))
	if err != nil {
		common.AbortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
		return nil, true
	}
	channel, err := model.GetChannelById(id, true)
	if err != nil {
		common.AbortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
		return nil, true
	}
	if channel.Status != common.ChannelStatusEnabled {
		common.AbortWithMessage(c, http.StatusForbidden, "该渠道已被禁用")
		return nil, true
	}

	return channel, false
}

func fetchChannelByModel(c *gin.Context, modelName string) (*model.Channel, bool) {
	group := c.GetString("group")
	channel, err := model.CacheGetRandomSatisfiedChannel(group, modelName)
	if err != nil {
		message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", group, modelName)
		if channel != nil {
			common.SysError(fmt.Sprintf("渠道不存在：%d", channel.Id))
			message = "数据库一致性已被破坏，请联系管理员"
		}
		common.AbortWithMessage(c, http.StatusServiceUnavailable, message)
		return nil, true
	}

	return channel, false
}

func getProvider(c *gin.Context, channel *model.Channel, relayMode int) (providersBase.ProviderInterface, bool) {
	provider := providers.GetProvider(channel, c)
	if provider == nil {
		common.AbortWithMessage(c, http.StatusNotImplemented, "channel not found")
		return nil, true
	}

	if !provider.SupportAPI(relayMode) {
		common.AbortWithMessage(c, http.StatusNotImplemented, "channel does not support this API")
		return nil, true
	}

	return provider, false
}

func shouldDisableChannel(err *types.OpenAIError, statusCode int) bool {
	if !common.AutomaticDisableChannelEnabled {
		return false
	}
	if err == nil {
		return false
	}
	if statusCode == http.StatusUnauthorized {
		return true
	}
	if err.Type == "insufficient_quota" || err.Code == "invalid_api_key" || err.Code == "account_deactivated" {
		return true
	}
	return false
}

func shouldEnableChannel(err error, openAIErr *types.OpenAIError) bool {
	if !common.AutomaticEnableChannelEnabled {
		return false
	}
	if err != nil {
		return false
	}
	if openAIErr != nil {
		return false
	}
	return true
}

func postConsumeQuota(ctx context.Context, tokenId int, quotaDelta int, totalQuota int, userId int, channelId int, modelRatio float64, groupRatio float64, modelName string, tokenName string) {
	// quotaDelta is remaining quota to be consumed
	err := model.PostConsumeTokenQuota(tokenId, quotaDelta)
	if err != nil {
		common.SysError("error consuming token remain quota: " + err.Error())
	}
	err = model.CacheUpdateUserQuota(userId)
	if err != nil {
		common.SysError("error update user quota cache: " + err.Error())
	}
	// totalQuota is total quota consumed
	if totalQuota != 0 {
		logContent := fmt.Sprintf("模型倍率 %.2f，分组倍率 %.2f", modelRatio, groupRatio)
		model.RecordConsumeLog(ctx, userId, channelId, totalQuota, 0, modelName, tokenName, totalQuota, logContent)
		model.UpdateUserUsedQuotaAndRequestCount(userId, totalQuota)
		model.UpdateChannelUsedQuota(channelId, totalQuota)
	}
	if totalQuota <= 0 {
		common.LogError(ctx, fmt.Sprintf("totalQuota consumed is %d, something is wrong", totalQuota))
	}
}

func parseModelMapping(modelMapping string) (map[string]string, error) {
	if modelMapping == "" || modelMapping == "{}" {
		return nil, nil
	}
	modelMap := make(map[string]string)
	err := json.Unmarshal([]byte(modelMapping), &modelMap)
	if err != nil {
		return nil, err
	}
	return modelMap, nil
}

type QuotaInfo struct {
	modelName         string
	promptTokens      int
	preConsumedTokens int
	modelRatio        float64
	groupRatio        float64
	ratio             float64
	preConsumedQuota  int
	userId            int
	channelId         int
	tokenId           int
	HandelStatus      bool
}

func generateQuotaInfo(c *gin.Context, modelName string, promptTokens int) (*QuotaInfo, *types.OpenAIErrorWithStatusCode) {
	quotaInfo := &QuotaInfo{
		modelName:    modelName,
		promptTokens: promptTokens,
		userId:       c.GetInt("id"),
		channelId:    c.GetInt("channel_id"),
		tokenId:      c.GetInt("token_id"),
		HandelStatus: false,
	}
	quotaInfo.initQuotaInfo(c.GetString("group"))

	errWithCode := quotaInfo.preQuotaConsumption()
	if errWithCode != nil {
		return nil, errWithCode
	}

	return quotaInfo, nil
}

func (q *QuotaInfo) initQuotaInfo(groupName string) {
	modelRatio := common.GetModelRatio(q.modelName)
	groupRatio := common.GetGroupRatio(groupName)
	preConsumedTokens := common.PreConsumedQuota
	ratio := modelRatio * groupRatio
	preConsumedQuota := int(float64(q.promptTokens+preConsumedTokens) * ratio)

	q.preConsumedTokens = preConsumedTokens
	q.modelRatio = modelRatio
	q.groupRatio = groupRatio
	q.ratio = ratio
	q.preConsumedQuota = preConsumedQuota

	return
}

func (q *QuotaInfo) preQuotaConsumption() *types.OpenAIErrorWithStatusCode {
	userQuota, err := model.CacheGetUserQuota(q.userId)
	if err != nil {
		return common.ErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}

	if userQuota < q.preConsumedQuota {
		return common.ErrorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}

	err = model.CacheDecreaseUserQuota(q.userId, q.preConsumedQuota)
	if err != nil {
		return common.ErrorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}

	if userQuota > 100*q.preConsumedQuota {
		// in this case, we do not pre-consume quota
		// because the user has enough quota
		q.preConsumedQuota = 0
		// common.LogInfo(c.Request.Context(), fmt.Sprintf("user %d has enough quota %d, trusted and no need to pre-consume", userId, userQuota))
	}

	if q.preConsumedQuota > 0 {
		err := model.PreConsumeTokenQuota(q.tokenId, q.preConsumedQuota)
		if err != nil {
			return common.ErrorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
		q.HandelStatus = true
	}

	return nil
}

func (q *QuotaInfo) completedQuotaConsumption(usage *types.Usage, tokenName string, ctx context.Context) error {
	quota := 0
	completionRatio := common.GetCompletionRatio(q.modelName)
	promptTokens := usage.PromptTokens
	completionTokens := usage.CompletionTokens
	quota = int(math.Ceil((float64(promptTokens) + float64(completionTokens)*completionRatio) * q.ratio))
	if q.ratio != 0 && quota <= 0 {
		quota = 1
	}
	totalTokens := promptTokens + completionTokens
	if totalTokens == 0 {
		// in this case, must be some error happened
		// we cannot just return, because we may have to return the pre-consumed quota
		quota = 0
	}
	quotaDelta := quota - q.preConsumedQuota
	err := model.PostConsumeTokenQuota(q.tokenId, quotaDelta)
	if err != nil {
		return errors.New("error consuming token remain quota: " + err.Error())
	}
	err = model.CacheUpdateUserQuota(q.userId)
	if err != nil {
		return errors.New("error consuming token remain quota: " + err.Error())
	}
	if quota != 0 {
		logContent := fmt.Sprintf("模型倍率 %.2f，分组倍率 %.2f", q.modelRatio, q.groupRatio)
		model.RecordConsumeLog(ctx, q.userId, q.channelId, promptTokens, completionTokens, q.modelName, tokenName, quota, logContent)
		model.UpdateUserUsedQuotaAndRequestCount(q.userId, quota)
		model.UpdateChannelUsedQuota(q.channelId, quota)
	}

	return nil
}
