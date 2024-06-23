package monitor

import "github.com/songquanpeng/one-api/relay/model"

type MonitorInstance interface {
	Emit(ChannelId int, success bool)
	ShouldDisableChannel(err *model.Error, statusCode int) bool
	DisableChannel(channelId int, channelName string, reason string)
}

type defaultMonitor struct {
}

func NewMonitorInstance() MonitorInstance {
	return &defaultMonitor{}
}

func (m *defaultMonitor) Emit(channelId int, success bool) {
	if success {
		metricSuccessChan <- channelId
	} else {
		metricFailChan <- channelId
	}
}

func (m *defaultMonitor) ShouldDisableChannel(err *model.Error, statusCode int) bool {
	return ShouldDisableChannel(err, statusCode)
}

func (m *defaultMonitor) DisableChannel(channelId int, channelName string, reason string) {
	DisableChannel(channelId, channelName, reason)
}
