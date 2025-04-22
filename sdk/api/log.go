package sdk

import (
	"encoding/json"
	"time"
)

// get log url like http://172.18.2.63:8300/api/log/?p=0&type=0&username=&token_name=&model_name=&start_timestamp=0&end_timestamp=1745237472&channel=
// define log struct :{
//            "id": 349,
//            "user_id": 21,
//            "created_at": 1745206602,
//            "type": 3,
//            "content": "管理员将用户额度从 ＄0.000000 额度修改为 ＄1000.000000 额度",
//            "username": "test1",
//            "token_name": "",
//            "model_name": "",
//            "quota": 0,
//            "prompt_tokens": 0,
//            "completion_tokens": 0,
//            "channel": 0,
//            "request_id": "2025042111364245599153931550114",
//            "elapsed_time": 0,
//            "is_stream": false,
//            "system_prompt_reset": false
//        },

type Log struct {
	ID                int    `json:"id"`
	UserID            int    `json:"user_id"`
	CreatedAt         int    `json:"created_at"`
	Type              int    `json:"type"`
	Content           string `json:"content"`
	Username          string `json:"username"`
	TokenName         string `json:"token_name"`
	ModelName         string `json:"model_name"`
	Quota             int    `json:"quota"`
	PromptTokens      int    `json:"prompt_tokens"`
	CompletionTokens  int    `json:"completion_tokens"`
	Channel           int    `json:"channel"`
	RequestID         string `json:"request_id"`
	ElapsedTime       int    `json:"elapsed_time"`
	IsStream          bool   `json:"is_stream"`
	SystemPromptReset bool   `json:"system_prompt_reset"`
}

type Logs struct {
	Logs  []*Log
	Query map[string]string
}

type Logsimpl interface {
	Get(client *OneClient) error
}

type LogRespData struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Success bool        `json:"success"`
}

// get log
func (logs *Logs) Get(client *OneClient) error {
	client.Url = "/api/log/?"
	if logs.Query != nil {
		for k, v := range logs.Query {
			client.Url += k + "=" + v + "&"
		}
	} else {
		client.Url = "/api/log/?p=0&type=0&username=&token_name=&model_name=&start_timestamp=0&end_timestamp=" + time.Now().String() + "&channel="
	}
	resp, err := client.get()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data := LogRespData{Data: []*Log{}, Message: "", Success: false}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}
	for _, v := range data.Data.([]interface{}) {
		log := &Log{}
		logData, _ := json.Marshal(v)
		err = json.Unmarshal(logData, log)
		if err != nil {
			return err
		}
		logs.Logs = append(logs.Logs, log)
	}
	return nil
}
