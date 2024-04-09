package channel

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common/requester"
	"one-api/types"
	"strconv"
	"time"
)

const larkURL = "https://open.feishu.cn/open-apis/bot/v2/hook/"

type Lark struct {
	token   string
	secret  string
	keyWord string
}

type larkMessage struct {
	MessageType string          `json:"msg_type"`
	Timestamp   string          `json:"timestamp,omitempty"`
	Sign        string          `json:"sign,omitempty"`
	Card        larkCardContent `json:"card"`
}

type larkCardContent struct {
	Config struct {
		WideScreenMode bool `json:"wide_screen_mode"`
		EnableForward  bool `json:"enable_forward"`
	}
	Elements []larkMessageRequestCardElement `json:"elements"`
}

type larkMessageRequestCardElementText struct {
	Content string `json:"content"`
	Tag     string `json:"tag"`
}

type larkMessageRequestCardElement struct {
	Tag  string                            `json:"tag"`
	Text larkMessageRequestCardElementText `json:"text"`
}

type larkResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func NewLark(token, secret string) *Lark {
	return &Lark{
		token:  token,
		secret: secret,
	}
}

func NewLarkWithKeyWord(token, keyWord string) *Lark {
	return &Lark{
		token:   token,
		keyWord: keyWord,
	}
}

func (l *Lark) Name() string {
	return "Lark"
}

func (l *Lark) Send(ctx context.Context, title, message string) error {
	msg := larkMessage{
		MessageType: "interactive",
	}

	if l.keyWord != "" {
		title = fmt.Sprintf("%s(%s)", title, l.keyWord)
	}

	msg.Card.Config.WideScreenMode = true
	msg.Card.Config.EnableForward = true
	msg.Card.Elements = append(msg.Card.Elements, larkMessageRequestCardElement{
		Tag: "div",
		Text: larkMessageRequestCardElementText{
			Content: fmt.Sprintf("**%s**\n%s", title, message),
			Tag:     "lark_md",
		},
	})

	if l.secret != "" {
		t := time.Now().Unix()
		msg.Timestamp = strconv.FormatInt(t, 10)
		msg.Sign = l.sign(t)
	}

	uri := larkURL + l.token
	client := requester.NewHTTPRequester("", larkErrFunc)
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

	larkErr := larkErrFunc(resp)
	if larkErr != nil {
		return fmt.Errorf("%s", larkErr.Message)
	}

	return nil

}

func (l *Lark) sign(timestamp int64) string {
	//timestamp + key 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + l.secret
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	h.Write(data)

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func larkErrFunc(resp *http.Response) *types.OpenAIError {
	respMsg := &larkResponse{}
	err := json.NewDecoder(resp.Body).Decode(respMsg)
	if err != nil {
		return nil
	}

	if respMsg.Code == 0 {
		return nil
	}

	return &types.OpenAIError{
		Message: fmt.Sprintf("send msg err. err msg: %s", respMsg.Message),
		Type:    "lark_error",
		Code:    respMsg.Code,
	}
}
