package channel

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"one-api/common/requester"
	"one-api/types"
	"time"
)

const dingTalkURL = "https://oapi.dingtalk.com/robot/send?"

type DingTalk struct {
	token   string
	secret  string
	keyWord string
}

type dingTalkMessage struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	} `json:"markdown"`
}

type dingTalkResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func NewDingTalk(token string, secret string) *DingTalk {
	return &DingTalk{
		token:  token,
		secret: secret,
	}
}

func NewDingTalkWithKeyWord(token string, keyWord string) *DingTalk {
	return &DingTalk{
		token:   token,
		keyWord: keyWord,
	}
}

func (d *DingTalk) Name() string {
	return "DingTalk"
}

func (d *DingTalk) Send(ctx context.Context, title, message string) error {
	msg := dingTalkMessage{
		MsgType: "markdown",
	}
	msg.Markdown.Title = title
	msg.Markdown.Text = message

	if d.keyWord != "" {
		msg.Markdown.Text = fmt.Sprintf("%s\n%s", d.keyWord, msg.Markdown.Text)
	}

	query := url.Values{}
	query.Set("access_token", d.token)
	if d.secret != "" {
		t := time.Now().UnixMilli()
		query.Set("timestamp", fmt.Sprintf("%d", t))
		query.Set("sign", d.sign(t))
	}
	uri := dingTalkURL + query.Encode()

	client := requester.NewHTTPRequester("", dingtalkErrFunc)
	client.Context = ctx
	client.IsOpenAI = false

	req, err := client.NewRequest(http.MethodPost, uri, client.WithHeader(requester.GetJsonHeaders()), client.WithBody(msg))
	if err != nil {
		return err
	}

	resp, errWithOP := client.SendRequestRaw(req)
	if errWithOP != nil {
		return fmt.Errorf("%s", errWithOP.Message)
	}
	defer resp.Body.Close()

	dingtalkErr := dingtalkErrFunc(resp)
	if dingtalkErr != nil {
		return fmt.Errorf("%s", dingtalkErr.Message)
	}

	return nil
}

func (d *DingTalk) sign(timestamp int64) string {
	stringToHash := fmt.Sprintf("%d\n%s", timestamp, d.secret)
	hmac256 := hmac.New(sha256.New, []byte(d.secret))
	hmac256.Write([]byte(stringToHash))
	data := hmac256.Sum(nil)
	signature := base64.StdEncoding.EncodeToString(data)

	return url.QueryEscape(signature)
}

func dingtalkErrFunc(resp *http.Response) *types.OpenAIError {
	respMsg := &dingTalkResponse{}

	err := json.NewDecoder(resp.Body).Decode(respMsg)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if respMsg.ErrCode == 0 {
		return nil
	}

	return &types.OpenAIError{
		Message: fmt.Sprintf("send msg err. err msg: %s", respMsg.ErrMsg),
		Type:    "dingtalk_error",
		Code:    fmt.Sprintf("%d", respMsg.ErrCode),
	}
}
