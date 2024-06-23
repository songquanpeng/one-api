package billing

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

// Bookkeeper 记账员逻辑，用于处理用户的配额消费
// 预扣配额检测逻辑：
//
//	开始请求前，根据不同的请求类型，预先计算需要消费的配额，根据请求用户和token的配额余量来判断是否有足够的配额来满足这个请求
//	如果余量配额不能满足这个请求，直接返回错误, 如果余量配额可以满足这个请求，那么预先消费这个配额，然后开始请求, 如果余量远远超过这个请求，那么不需要预先消费配额
//	由于预先计算的配额不是实际消费的配额，所以需要在请求结束后，根据实际消费的配额来更新用户和token的配额，退费或者扣费。
type Bookkeeper interface {
	// 获取模型的费率
	ModelRatio(model string) float64
	// 获取组的费率
	GroupRation(group string) float64
	// 获取模型的补全费率
	ModelCompletionRatio(model string) float64
	// 根据消费记录，扣除用户，token 的配额
	Consume(ctx context.Context, consumeLog *ConsumeLog)
	// 预消费配额, 当用户配额不足时，预消费配额， 预消费成功返回预消费的配额，失败返回错误, 如果预消费的配额为0，表示用户有足够的配额
	PreConsumeQuota(ctx context.Context, preConsumedQuota int64, userId, tokenId int) (int64, *relaymodel.ErrorWithStatusCode)
	// 退回预消费的配额, 这通常在调用上游api失败的时候执行
	RefundQuota(ctx context.Context, preConsumedQuota int64, tokenId int)

	// 检测用户是否有足够的配额
	// UserHasEnoughQuota(ctx context.Context, userID int, quota int64) bool
	// 检测用户是否有远远超过需求的配额, 如果用户的配额远远超过需求，那么不需要预消费配额
	// UserHasMuchMoreQuota(ctx context.Context, userID int, quota int64) bool
}

type defaultBookkeeper struct {
}

func NewBookkeeper() Bookkeeper {
	return &defaultBookkeeper{}
}

func (b *defaultBookkeeper) ModelRatio(model string) float64 {
	return billingratio.GetModelRatio(model)
}

func (b *defaultBookkeeper) GroupRation(group string) float64 {
	return billingratio.GetGroupRatio(group)
}

func (b *defaultBookkeeper) ModelCompletionRatio(model string) float64 {
	return billingratio.GetCompletionRatio(model)
}

func (b *defaultBookkeeper) Ratio(group, model string) float64 {
	modelRatio := billingratio.GetModelRatio(model)
	groupRatio := billingratio.GetGroupRatio(group)
	return modelRatio * groupRatio
}

// ConsumeLog 消费记录实体
type ConsumeLog struct {
	UserId           int
	ChannelId        int
	PromptTokens     int
	CompletionTokens int
	ModelName        string
	TokenId          int
	TokenName        string
	Quota            int64
	Content          string
	PreConsumedQuota int64
}

func (b *defaultBookkeeper) UserHasEnoughQuota(ctx context.Context, userID int, quota int64) bool {
	userQuota, err := model.CacheGetUserQuota(ctx, userID)
	if err != nil {
		return false
	}
	return userQuota >= quota
}

func (b *defaultBookkeeper) UserHasMuchMoreQuota(ctx context.Context, userID int, quota int64) bool {
	userQuota, err := model.CacheGetUserQuota(ctx, userID)
	if err != nil {
		return false
	}
	return userQuota > 100*quota
}

func (b *defaultBookkeeper) Consume(ctx context.Context, consumeLog *ConsumeLog) {
	// 更新 access_token 的配额
	quotaDelta := consumeLog.Quota - consumeLog.PreConsumedQuota
	err := model.PostConsumeTokenQuota(consumeLog.TokenId, quotaDelta)
	if err != nil {
		logger.SysError("error consuming token remain quota: " + err.Error())
	}
	err = model.CacheUpdateUserQuota(ctx, consumeLog.UserId)
	if err != nil {
		logger.SysError("error update user quota cache: " + err.Error())
	}
	// 更新用户的配额
	model.UpdateUserUsedQuotaAndRequestCount(consumeLog.UserId, consumeLog.Quota)
	// 更新渠道的配额
	model.UpdateChannelUsedQuota(consumeLog.ChannelId, consumeLog.Quota)
	// 记录消费日志
	model.RecordConsumeLog(
		ctx,
		consumeLog.UserId,
		consumeLog.ChannelId,
		consumeLog.PromptTokens,
		consumeLog.CompletionTokens,
		consumeLog.ModelName,
		consumeLog.TokenName,
		consumeLog.Quota,
		consumeLog.Content,
	)
}

func (b *defaultBookkeeper) PreConsumeQuota(ctx context.Context, preConsumedQuota int64, userId, tokenId int) (int64, *relaymodel.ErrorWithStatusCode) {
	userQuota, err := model.CacheGetUserQuota(ctx, userId)
	if err != nil {
		return preConsumedQuota, openai.ErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota-preConsumedQuota < 0 {
		return preConsumedQuota, openai.ErrorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}
	err = model.CacheDecreaseUserQuota(userId, preConsumedQuota)
	if err != nil {
		return preConsumedQuota, openai.ErrorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota > 100*preConsumedQuota {
		// in this case, we do not pre-consume quota
		// because the user has enough quota
		preConsumedQuota = 0
		logger.Info(ctx, fmt.Sprintf("user %d has enough quota %d, trusted and no need to pre-consume", userId, userQuota))
	}
	if preConsumedQuota > 0 {
		err := model.PreConsumeTokenQuota(tokenId, preConsumedQuota)
		if err != nil {
			return preConsumedQuota, openai.ErrorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
	}
	return preConsumedQuota, nil
}

func (b *defaultBookkeeper) RefundQuota(ctx context.Context, preConsumedQuota int64, tokenId int) {
	if preConsumedQuota != 0 {
		err := model.PostConsumeTokenQuota(tokenId, -preConsumedQuota)
		if err != nil {
			logger.Error(ctx, "error return pre-consumed quota: "+err.Error())
		}
	}
}
