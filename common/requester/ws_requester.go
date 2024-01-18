package requester

import (
	"errors"
	"net/http"
	"one-api/common"
	"one-api/types"

	"github.com/gorilla/websocket"
)

type WSRequester struct {
	WSClient *websocket.Dialer
}

func NewWSRequester(proxyAddr string) *WSRequester {
	return &WSRequester{
		WSClient: GetWSClient(proxyAddr),
	}
}

func (w *WSRequester) NewRequest(url string, header http.Header) (*websocket.Conn, error) {
	conn, resp, err := w.WSClient.Dial(url, header)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		return nil, errors.New("ws unexpected status code")
	}

	return conn, nil
}

func SendWSJsonRequest[T streamable](conn *websocket.Conn, data any, handlerPrefix HandlerPrefix[T]) (*wsReader[T], *types.OpenAIErrorWithStatusCode) {
	err := conn.WriteJSON(data)
	if err != nil {
		return nil, common.ErrorWrapper(err, "ws_request_failed", http.StatusInternalServerError)
	}

	return &wsReader[T]{
		reader:        conn,
		handlerPrefix: handlerPrefix,
	}, nil
}

// 设置请求头
func (r *WSRequester) WithHeader(headers map[string]string) http.Header {
	header := make(http.Header)
	for k, v := range headers {
		header.Set(k, v)
	}
	return header
}
