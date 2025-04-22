package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type Config struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	Key  string `json:"key"`
}

type OneClient struct {
	Client *http.Client
	Config *Config
	Url    string
}

type RespMessage struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// get
func (OneClient *OneClient) get() (*http.Response, error) {
	OneClient.Client = &http.Client{}
	port := strconv.Itoa(OneClient.Config.Port)
	url := OneClient.Config.Host + ":" + port + OneClient.Url
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+OneClient.Config.Key)
	resp, err := OneClient.Client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// post
func (OneClient *OneClient) post(data []byte) error {
	OneClient.Client = &http.Client{}
	port := strconv.Itoa(OneClient.Config.Port)
	url := OneClient.Config.Host + ":" + port + OneClient.Url
	payload := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", url, payload)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+OneClient.Config.Key)
	resp, err := OneClient.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	message := RespMessage{}
	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		return err
	}
	if message.Success {
		return nil
	}
	return fmt.Errorf("create user failed: %s", message.Message)
}

// put
func (OneClient *OneClient) put(data []byte) error {
	OneClient.Client = &http.Client{}
	port := strconv.Itoa(OneClient.Config.Port)
	url := OneClient.Config.Host + ":" + port + OneClient.Url
	payload := bytes.NewBuffer(data)
	req, err := http.NewRequest("PUT", url, payload)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+OneClient.Config.Key)
	resp, err := OneClient.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	message := RespMessage{}
	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		return err
	}
	if message.Success {
		return nil
	}
	return fmt.Errorf("update user failed: %s", message.Message)
}

// delete
func (OneClient *OneClient) delete(data []byte) error {
	OneClient.Client = &http.Client{}
	port := strconv.Itoa(OneClient.Config.Port)
	url := OneClient.Config.Host + ":" + port + OneClient.Url
	payload := bytes.NewBuffer(data)
	req, err := http.NewRequest("DELETE", url, payload)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+OneClient.Config.Key)
	resp, err := OneClient.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	message := RespMessage{}
	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		return err
	}
	if message.Success {
		return nil
	}
	return fmt.Errorf("delete user failed: %s", message.Message)
}
