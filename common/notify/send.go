package notify

import (
	"context"
	"fmt"
	"one-api/common/logger"
)

func (n *Notify) Send(ctx context.Context, title, message string) {
	if ctx == nil {
		ctx = context.Background()
	}

	for channelName, channel := range n.notifiers {
		if channel == nil {
			continue
		}
		err := channel.Send(ctx, title, message)
		if err != nil {
			logger.LogError(ctx, fmt.Sprintf("%s err: %s", channelName, err.Error()))
		}
	}
}

func Send(title, message string) {
	//lint:ignore SA1029 reason: 需要使用该类型作为错误处理
	ctx := context.WithValue(context.Background(), logger.RequestIdKey, "NotifyTask")

	notifyChannels.Send(ctx, title, message)
}
