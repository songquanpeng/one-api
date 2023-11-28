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
	"one-api/types"
)

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

func postConsumeQuota(ctx context.Context, tokenId int, quota int, userId int, channelId int, modelRatio float64, groupRatio float64, modelName string, tokenName string) {
	err := model.PostConsumeTokenQuota(tokenId, quota)
	if err != nil {
		common.SysError("error consuming token remain quota: " + err.Error())
	}
	err = model.CacheUpdateUserQuota(userId)
	if err != nil {
		common.SysError("error update user quota cache: " + err.Error())
	}
	if quota != 0 {
		logContent := fmt.Sprintf("模型倍率 %.2f，分组倍率 %.2f", modelRatio, groupRatio)
		model.RecordConsumeLog(ctx, userId, channelId, 0, 0, modelName, tokenName, quota, logContent)
		model.UpdateUserUsedQuotaAndRequestCount(userId, quota)
		model.UpdateChannelUsedQuota(channelId, quota)
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
		return types.ErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}

	if userQuota < q.preConsumedQuota {
		return types.ErrorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}

	err = model.CacheDecreaseUserQuota(q.userId, q.preConsumedQuota)
	if err != nil {
		return types.ErrorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
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
			return types.ErrorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
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
