package sdk

import (
	"encoding/json"
	"strconv"
)

type Token struct {
	ID             int    `json:"id"`
	UserID         int    `json:"user_id"`
	Key            string `json:"key"`
	Status         int    `json:"status"`
	Name           string `json:"name"`
	CreatedTime    int    `json:"created_time"`
	AccessedTime   int    `json:"accessed_time"`
	ExpiredTime    int    `json:"expired_time"`
	RemainQuota    int    `json:"remain_quota"`
	UnlimitedQuota bool   `json:"unlimited_quota"`
	UsedQuota      int    `json:"used_quota"`
	Models         string `json:"models"`
	Subnet         string `json:"subnet"`
}

// define add token function
type Tokenimpl interface {
	Add(token *Token) error
	List(id int) error
	Update(token *Token) error
	Delete(id int) error
}

type Tokens struct {
	Tokens []*Token
	UserID int
	Query  map[string]string
}

type TokensImpl interface {
	List(token *Token) error
}

type TokenRespData struct {
	Data interface{} `json:"data"`
}

// add token
func (token *Token) Add(client *OneClient) error {
	client.Url = "/api/token/"
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return client.post(data)
}

// update token
func (token *Token) Update(client *OneClient) error {
	client.Url = "/api/token/"
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return client.put(data)
}

// list token
func (tokens *Tokens) List(client *OneClient) error {
	if tokens.UserID != 0 {
		if tokens.Query != nil {
			client.Url = "/api/token/search?user_id=" + strconv.Itoa(tokens.UserID)
			for k, v := range tokens.Query {
				client.Url += "&" + k + "=" + v
			}
			client.Url += "&p=0&order="
		} else {
			client.Url = "/api/token/?user_id=" + strconv.Itoa(tokens.UserID) + "&p=0&order="
		}
	} else {
		if tokens.Query != nil {
			client.Url = "/api/token/search?p=0"
			for k, v := range tokens.Query {
				client.Url += "&" + k + "=" + v
			}
		} else {
			client.Url = "/api/token/?p=0"
		}
		client.Url += "&order="
	}
	resp, err := client.get()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data := TokenRespData{Data: []*Token{}}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}
	for _, v := range data.Data.([]interface{}) {
		token := &Token{}
		tokenData, _ := json.Marshal(v)
		err = json.Unmarshal(tokenData, token)
		if err != nil {
			return err
		}
		tokens.Tokens = append(tokens.Tokens, token)
	}
	return nil
}

// delete token
func (token *Token) Delete(client *OneClient) error {
	client.Url = "/api/token/" + strconv.Itoa(token.ID) + "/" + "?user_id=" + strconv.Itoa(token.UserID)
	return client.delete(nil)
}

// get token
func (token *Token) Get(client *OneClient) error {
	client.Url = "/api/token/" + strconv.Itoa(token.ID) + "/" + "?user_id=" + strconv.Itoa(token.UserID)
	resp, err := client.get()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data := TokenRespData{Data: token}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}
	tokenData, _ := json.Marshal(data.Data)
	err = json.Unmarshal(tokenData, token)
	return nil
}
