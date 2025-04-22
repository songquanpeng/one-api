package sdk

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Channel struct {
	ID                 int           `json:"id"`
	Type               int           `json:"type"`
	Key                string        `json:"key"`
	Status             int           `json:"status"`
	Name               string        `json:"name"`
	Weight             int           `json:"weight"`
	CreatedTime        int           `json:"created_time"`
	TestTime           int           `json:"test_time"`
	ResponseTime       int           `json:"response_time"`
	BaseUrl            string        `json:"base_url"`
	Other              string        `json:"other"`
	Balance            int           `json:"balance"`
	BalanceUpdatedTime int           `json:"balance_updated_time"`
	Models             string        `json:"models"`
	Group              string        `json:"group"`
	UsedQuota          int           `json:"used_quota"`
	ModelMapping       string        `json:"model_mapping"`
	Priority           int           `json:"priority"`
	Config             string        `json:"config"`
	SystemPrompt       string        `json:"system_prompt"`
	ChannelConfig      ChannelConfig `json:"channel_config"`
}

type ChannelConfig struct {
	Region            string `json:"region"`
	Sk                string `json:"sk"`
	Ak                string `json:"ak"`
	UserId            string `json:"user_id"`
	VertexAiProjectId string `json:"vertex_ai_project_id"`
	VertexAiAdc       string `json:"vertex_ai_adc"`
}
type NewChannel struct {
	BaseUrl      string   `json:"base_url"`
	Config       string   `json:"config"`
	Group        string   `json:"group"`
	Groups       []string `json:"groups"`
	Key          string   `json:"key"`
	ModelMapping string   `json:"model_mapping"`
	Models       string   `json:"models"`
	Name         string   `json:"name"`
	Other        string   `json:"other"`
	SystemPrompt string   `json:"system_prompt"`
	Type         int      `json:"type"`
}

type Channels struct {
	Channels []*Channel
	Query    map[string]string
}

type ChannelRespData struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Success bool        `json:"success"`
}

type ChannelResp struct {
	Message   string  `json:"message"`
	ModelName string  `json:"modelName"`
	Success   bool    `json:"success"`
	Time      float64 `json:"time"`
}

type ChannelImpl interface {
	Add(channel *Channel) error
	Get(id int) error
	Update(channel *Channel) error
	Delete(id int) error
	Test() error
}

// list channel
func (channels *Channels) List(client *OneClient) error {
	if channels.Query != nil {
		client.Url = "/api/channel/search?"
		for k, v := range channels.Query {
			client.Url += k + "=" + v + "&"
		}
		client.Url += "p=0&order="
	} else {
		client.Url = "/api/channel/?p=0&order="
	}
	resp, err := client.get()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data := ChannelRespData{Data: []*Channel{}, Message: "", Success: false}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}
	for _, v := range data.Data.([]interface{}) {
		channel := &Channel{}
		channelData, _ := json.Marshal(v)
		err = json.Unmarshal(channelData, channel)
		if err != nil {
			return err
		}
		channels.Channels = append(channels.Channels, channel)
	}
	return nil
}

// add channel
func (channel *Channel) Add(client *OneClient) error {
	client.Url = "/api/channel/"
	channelConfigData, err := json.Marshal(channel.ChannelConfig)
	newChannel := NewChannel{
		BaseUrl:      channel.BaseUrl,
		Config:       string(channelConfigData),
		Group:        channel.Group,
		Groups:       []string{channel.Group},
		Key:          channel.Key,
		ModelMapping: channel.ModelMapping,
		Models:       channel.Models,
		Name:         channel.Name,
		Other:        channel.Other,
		SystemPrompt: channel.SystemPrompt,
		Type:         channel.Type,
	}
	data, err := json.Marshal(newChannel)
	if err != nil {
		return err
	}
	return client.post(data)
}

// update channel
func (channel *Channel) Update(client *OneClient) error {
	client.Url = "/api/channel/"
	data, err := json.Marshal(channel)
	if err != nil {
		return err
	}
	return client.put(data)
}

// delete channel
func (channel *Channel) Delete(client *OneClient) error {
	client.Url = "/api/channel/" + fmt.Sprintf("%d", channel.ID)
	return client.delete(nil)
}

// get channel
func (channel *Channel) Get(client *OneClient) error {
	client.Url = "/api/channel/" + fmt.Sprintf("%d", channel.ID) + "/"
	resp, err := client.get()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data := ChannelRespData{Data: channel, Message: "", Success: false}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}
	*channel = *(data.Data.(*Channel))
}

// test channel
func (channel *Channel) Test(client *OneClient) error {
	client.Url = "/api/channel/test/" + fmt.Sprintf("%d", channel.ID) + "/?model=" + strings.Split(channel.Models, ",")[0]
	resp, err := client.get()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data := ChannelResp{Message: "", ModelName: "", Success: false, Time: 0}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}
	if data.Success {
		return nil
	} else {
		return fmt.Errorf("test channel failed: %s", data.Message)
	}
}
