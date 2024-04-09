package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common/requester"
	"one-api/types"
	"strings"
)

const pushdeerURL = "https://api2.pushdeer.com"

type Pushdeer struct {
	url     string
	pushkey string
}

type pushdeerMessage struct {
	Text string `json:"text"`
	Desp string `json:"desp"`
	Type string `json:"type"`
}

type pushdeerResponse struct {
	Code    int    `json:"code,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

func NewPushdeer(pushkey, url string) *Pushdeer {
	return &Pushdeer{
		url:     url,
		pushkey: pushkey,
	}
}

func (p *Pushdeer) Name() string {
	return "Pushdeer"
}

func (p *Pushdeer) Send(ctx context.Context, title, message string) error {
	msg := pushdeerMessage{
		Text: title,
		Desp: message,
		Type: "markdown",
	}

	url := p.url
	if url == "" {
		url = pushdeerURL
	}

	// 去除最后一个/
	url = strings.TrimSuffix(url, "/")
	uri := fmt.Sprintf("%s/message/push?pushkey=%s", url, p.pushkey)

	client := requester.NewHTTPRequester("", pushdeerErrFunc)
	client.Context = ctx
	client.IsOpenAI = false

	req, err := client.NewRequest(http.MethodPost, uri, client.WithHeader(requester.GetJsonHeaders()), client.WithBody(msg))
	if err != nil {
		return err
	}

	respMsg := &pushdeerResponse{}
	_, errWithOP := client.SendRequest(req, respMsg, false)
	if errWithOP != nil {
		return fmt.Errorf("%s", errWithOP.Message)
	}

	if respMsg.Code != 0 {
		return fmt.Errorf("send msg err. err msg: %s", respMsg.Error)
	}

	return nil
}

func pushdeerErrFunc(resp *http.Response) *types.OpenAIError {
	respMsg := &pushdeerResponse{}
	err := json.NewDecoder(resp.Body).Decode(respMsg)
	if err != nil {
		return nil
	}

	if respMsg.Message == "" {
		return nil
	}

	return &types.OpenAIError{
		Message: fmt.Sprintf("send msg err. err msg: %s", respMsg.Message),
		Type:    "pushdeer_error",
	}
}
