// 6ca91f95d29749db8a93d5b8903c7949

// description: User API
// use httprequest to get user list	and user info
package sdk

import (
	"encoding/json"
	"fmt"
)

type User struct {
	ID               int    `json:"id"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	DisplayName      string `json:"display_name"`
	Role             int    `json:"role"`
	Status           int    `json:"status"`
	Email            string `json:"email"`
	GithubID         string `json:"github_id"`
	WechatID         string `json:"wechat_id"`
	LarkID           string `json:"lark_id"`
	OidcID           string `json:"oidc_id"`
	VerificationCode string `json:"verification_code"`
	AccessToken      string `json:"access_token"`
	Quota            int    `json:"quota"`
	UsedQuota        int    `json:"used_quota"`
	RequestCount     int    `json:"request_count"`
	Group            string `json:"group"`
	AffCode          string `json:"aff_code"`
	InviterID        int    `json:"inviter_id"`
}

type UserRespData struct {
	Data interface{} `json:"data"`
}

type UserImpl interface {
	Add(user *User) error
	Get(id int) error
	Updater(user *User) error
	Delete(id int) error
}

type Users struct {
	Users []*User
	Query map[string]string
}

// list user
func (users *Users) List(client *OneClient) error {
	if users.Query != nil {
		client.Url = "/api/user/search?"
		for k, v := range users.Query {
			client.Url += k + "=" + v + "&"
		}
		client.Url += "p=0&order="
	} else {
		client.Url = "/api/user/?p=0&order="
	}
	resp, err := client.get()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data := UserRespData{Data: []*User{}}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}
	for _, v := range data.Data.([]interface{}) {
		user := &User{}
		userData, _ := json.Marshal(v)
		err = json.Unmarshal(userData, user)
		if err != nil {
			return err
		}
		users.Users = append(users.Users, user)
	}
	return nil
}

// add user
func (user *User) Add(client *OneClient) error {
	client.Url = "/api/user/"
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return client.post(data)
}

// delete user
func (user *User) Delete(client *OneClient) error {
	client.Url = "/api/user/manage"
	deleteData := map[string]interface{}{
		"username": user.Username,
		"action":   "delete",
	}
	data, err := json.Marshal(deleteData)
	if err != nil {
		return err
	}
	return client.post(data)
}

// update user
func (user *User) Update(client *OneClient) error {
	client.Url = "/api/user"
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return client.put(data)
}

// get user
func (user *User) Get(client *OneClient) error {
	client.Url = "/api/user/" + fmt.Sprintf("%d", user.ID)
	resp, err := client.get()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data := UserRespData{Data: user}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}
	userData, _ := json.Marshal(data.Data)
	err = json.Unmarshal(userData, user)
	if err != nil {
		return err
	}
	return nil
}
