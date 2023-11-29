package zhipu

import (
	"one-api/types"
	"time"
)

type ZhipuMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ZhipuRequest struct {
	Prompt      []ZhipuMessage `json:"prompt"`
	Temperature float64        `json:"temperature,omitempty"`
	TopP        float64        `json:"top_p,omitempty"`
	RequestId   string         `json:"request_id,omitempty"`
	Incremental bool           `json:"incremental,omitempty"`
}

type ZhipuResponseData struct {
	TaskId      string         `json:"task_id"`
	RequestId   string         `json:"request_id"`
	TaskStatus  string         `json:"task_status"`
	Choices     []ZhipuMessage `json:"choices"`
	types.Usage `json:"usage"`
}

type ZhipuResponse struct {
	Code    int               `json:"code"`
	Msg     string            `json:"msg"`
	Success bool              `json:"success"`
	Data    ZhipuResponseData `json:"data"`
}

type ZhipuStreamMetaResponse struct {
	RequestId   string `json:"request_id"`
	TaskId      string `json:"task_id"`
	TaskStatus  string `json:"task_status"`
	types.Usage `json:"usage"`
}

type zhipuTokenData struct {
	Token      string
	ExpiryTime time.Time
}
