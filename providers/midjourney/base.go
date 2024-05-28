package midjourney

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"one-api/common/config"
	"one-api/common/logger"
	"one-api/common/requester"
	"one-api/model"
	"one-api/providers/base"
	"time"
)

// 定义供应商工厂
type MidjourneyProviderFactory struct{}

// 创建 MidjourneyProvider
func (f MidjourneyProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &MidjourneyProvider{
		BaseProvider: base.BaseProvider{
			Config:    getConfig(),
			Channel:   channel,
			Requester: requester.NewHTTPRequester(*channel.Proxy, nil),
		},
	}
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL: "",
	}
}

type MidjourneyProvider struct {
	base.BaseProvider
}

func (p *MidjourneyProvider) Send(timeout int, requestURL string) (*MidjourneyResponseWithStatusCode, []byte, error) {
	var nullBytes []byte
	var mapResult map[string]interface{}
	if p.Context.Request.Method != "GET" {
		err := json.NewDecoder(p.Context.Request.Body).Decode(&mapResult)
		if err != nil {
			return MidjourneyErrorWithStatusCodeWrapper(MjErrorUnknown, "read_request_body_failed", http.StatusInternalServerError), nullBytes, err
		}
		delete(mapResult, "accountFilter")
		if !config.MjNotifyEnabled {
			delete(mapResult, "notifyHook")
		}
	}

	reqBody, err := json.Marshal(mapResult)
	if err != nil {
		return MidjourneyErrorWithStatusCodeWrapper(MjErrorUnknown, "marshal_request_body_failed", http.StatusInternalServerError), nullBytes, err
	}

	fullRequestURL := p.GetFullRequestURL(requestURL, "")

	var cancel context.CancelFunc
	p.Requester.Context, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	headers := p.GetRequestHeaders()
	defer cancel()

	req, err := p.Requester.NewRequest(p.Context.Request.Method, fullRequestURL, p.Requester.WithBody(bytes.NewBuffer(reqBody)), p.Requester.WithHeader(headers))
	if err != nil {
		return MidjourneyErrorWithStatusCodeWrapper(MjErrorUnknown, "create_request_failed", http.StatusInternalServerError), nullBytes, err
	}

	resp, errWith := p.Requester.SendRequestRaw(req)
	if errWith != nil {
		logger.SysError("do request failed: " + errWith.Error())
		return MidjourneyErrorWithStatusCodeWrapper(MjErrorUnknown, "do_request_failed", http.StatusInternalServerError), nullBytes, err
	}
	statusCode := resp.StatusCode
	err = req.Body.Close()
	if err != nil {
		return MidjourneyErrorWithStatusCodeWrapper(MjErrorUnknown, "close_request_body_failed", statusCode), nullBytes, err
	}
	err = p.Context.Request.Body.Close()
	if err != nil {
		return MidjourneyErrorWithStatusCodeWrapper(MjErrorUnknown, "close_request_body_failed", statusCode), nullBytes, err
	}
	var midjResponse MidjourneyResponse

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return MidjourneyErrorWithStatusCodeWrapper(MjErrorUnknown, "read_response_body_failed", statusCode), nullBytes, err
	}
	err = resp.Body.Close()
	if err != nil {
		return MidjourneyErrorWithStatusCodeWrapper(MjErrorUnknown, "close_response_body_failed", statusCode), responseBody, err
	}
	respStr := string(responseBody)
	log.Printf("responseBody: %s", respStr)
	if respStr == "" {
		return MidjourneyErrorWithStatusCodeWrapper(MjErrorUnknown, "empty_response_body", statusCode), responseBody, nil
	} else {
		err = json.Unmarshal(responseBody, &midjResponse)
		if err != nil {
			return MidjourneyErrorWithStatusCodeWrapper(MjErrorUnknown, "unmarshal_response_body_failed", statusCode), responseBody, err
		}
	}

	return &MidjourneyResponseWithStatusCode{
		StatusCode: statusCode,
		Response:   midjResponse,
	}, responseBody, nil

}

func (p *MidjourneyProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	headers["mj-api-secret"] = p.Channel.Key
	headers["Content-Type"] = p.Context.Request.Header.Get("Content-Type")
	headers["Accept"] = p.Context.Request.Header.Get("Accept")

	return headers
}
