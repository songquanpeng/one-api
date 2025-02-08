package monitor

import (
	"fmt"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/message"
	"github.com/songquanpeng/one-api/model"
)

func notifyRootUser(subject string, content string) {
	if config.MessagePusherAddress != "" {
		err := message.SendMessage(subject, content, content)
		if err != nil {
			logger.SysError(fmt.Sprintf("failed to send message: %s", err.Error()))
		} else {
			return
		}
	}
	if config.RootUserEmail == "" {
		config.RootUserEmail = model.GetRootUserEmail()
	}
	err := message.SendEmail(subject, config.RootUserEmail, content)
	if err != nil {
		logger.SysError(fmt.Sprintf("failed to send email: %s", err.Error()))
	}
}

// DisableChannel disable & notify
func DisableChannel(channelId int, channelName string, reason string) {
	model.UpdateChannelStatusById(channelId, model.ChannelStatusAutoDisabled)
	logger.SysLog(fmt.Sprintf("channel #%d has been disabled: %s", channelId, reason))
	subject := fmt.Sprintf("渠道状态变更提醒")
	content := message.EmailTemplate(
		subject,
		fmt.Sprintf(`
			<p>您好！</p>
			<p>渠道「<strong>%s</strong>」（#%d）已被禁用。</p>
			<p>禁用原因：</p>
			<p style="background-color: #f8f8f8; padding: 10px; border-radius: 4px;">%s</p>
		`, channelName, channelId, reason),
	)
	notifyRootUser(subject, content)
}

func MetricDisableChannel(channelId int, successRate float64) {
	model.UpdateChannelStatusById(channelId, model.ChannelStatusAutoDisabled)
	logger.SysLog(fmt.Sprintf("channel #%d has been disabled due to low success rate: %.2f", channelId, successRate*100))
	subject := fmt.Sprintf("渠道状态变更提醒")
	content := message.EmailTemplate(
		subject,
		fmt.Sprintf(`
			<p>您好！</p>
			<p>渠道 #%d 已被系统自动禁用。</p>
			<p>禁用原因：</p>
			<p style="background-color: #f8f8f8; padding: 10px; border-radius: 4px;">该渠道在最近 %d 次调用中成功率为 <strong>%.2f%%</strong>，低于系统阈值 <strong>%.2f%%</strong>。</p>
		`, channelId, config.MetricQueueSize, successRate*100, config.MetricSuccessRateThreshold*100),
	)
	notifyRootUser(subject, content)
}

// EnableChannel enable & notify
func EnableChannel(channelId int, channelName string) {
	model.UpdateChannelStatusById(channelId, model.ChannelStatusEnabled)
	logger.SysLog(fmt.Sprintf("channel #%d has been enabled", channelId))
	subject := fmt.Sprintf("渠道状态变更提醒")
	content := message.EmailTemplate(
		subject,
		fmt.Sprintf(`
			<p>您好！</p>
			<p>渠道「<strong>%s</strong>」（#%d）已被重新启用。</p>
			<p>您现在可以继续使用该渠道了。</p>
		`, channelName, channelId),
	)
	notifyRootUser(subject, content)
}
