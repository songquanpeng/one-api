package ali

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
	"io"
	"net/http"
)

// https://help.aliyun.com/zh/dashscope/developer-reference/api-details

type Adaptor struct {
	meta *meta.Meta
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.meta = meta
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	fullRequestURL := ""
	switch meta.Mode {
	case relaymode.Embeddings:
		fullRequestURL = fmt.Sprintf("%s/api/v1/services/embeddings/text-embedding/text-embedding", meta.BaseURL)
	case relaymode.ImagesGenerations:
		fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/text2image/image-synthesis", meta.BaseURL)
	default:
		fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/text-generation/generation", meta.BaseURL)
	}

	return fullRequestURL, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	if meta.IsStream {
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("X-DashScope-SSE", "enable")
	}
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)

	if meta.Mode == relaymode.ImagesGenerations {
		req.Header.Set("X-DashScope-Async", "enable")
	}
	if a.meta.Config.Plugin != "" {
		req.Header.Set("X-DashScope-Plugin", a.meta.Config.Plugin)
	}
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	switch relayMode {
	case relaymode.Embeddings:
		aliEmbeddingRequest := ConvertEmbeddingRequest(*request)
		return aliEmbeddingRequest, nil
	default:
		aliRequest := ConvertRequest(*request)
		return aliRequest, nil
	}
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	aliRequest := ConvertImageRequest(*request)
	return aliRequest, nil
}

func (a *Adaptor) ConvertTextToSpeechRequest(request *model.TextToSpeechRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	aliRequest := ConvertTextToSpeechRequest(*request)
	return aliRequest, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	if meta.Mode == relaymode.AudioSpeech {
		return a.DoWSSRequest(c, meta, requestBody)
	}
	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		err, usage = StreamHandler(c, resp)
	} else {
		switch meta.Mode {
		case relaymode.Embeddings:
			err, usage = EmbeddingHandler(c, resp)
		case relaymode.ImagesGenerations:
			err, usage = ImageHandler(c, resp)
		case relaymode.AudioSpeech:
			err, usage = AudioSpeechHandler(c, resp)
		default:
			err, usage = Handler(c, resp)
		}
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "ali"
}

func (a *Adaptor) DoWSSRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	baseURL := "wss://dashscope.aliyuncs.com/api-ws/v1/inference"
	var usage Usage
	// Create an empty http.Response object
	response := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(nil),
	}

	conn, _, err := websocket.DefaultDialer.Dial(baseURL, http.Header{"Authorization": {"Bearer " + meta.APIKey}})
	if err != nil {
		return response, errors.New("ali_wss_conn_failed")
	}
	defer conn.Close()

	var requestBodyBytes []byte
	requestBodyBytes, err = io.ReadAll(requestBody)
	if err != nil {
		return response, errors.New("ali_failed_to_read_request_body")
	}

	// Convert JSON strings to map[string]interface{}
	var requestBodyMap map[string]interface{}
	err = json.Unmarshal(requestBodyBytes, &requestBodyMap)
	if err != nil {
		return response, errors.New("ali_failed_to_parse_request_body")
	}

	if err := conn.WriteJSON(requestBodyMap); err != nil {
		return response, errors.New("ali_wss_write_msg_failed")
	}

	const chunkSize = 1024

	for {
		messageType, audioData, err := conn.ReadMessage()
		if err != nil {
			if err == io.EOF {
				break
			}
			return response, errors.New("ali_wss_read_msg_failed")
		}

		var msg WSSMessage
		switch messageType {
		case websocket.TextMessage:
			err = json.Unmarshal(audioData, &msg)
			if msg.Header.Event == "task-finished" {
				response.StatusCode = http.StatusOK
				usage.TotalTokens = msg.Payload.Usage.Characters
				return response, nil
			}
		case websocket.BinaryMessage:
			for i := 0; i < len(audioData); i += chunkSize {
				end := i + chunkSize
				if end > len(audioData) {
					end = len(audioData)
				}
				chunk := audioData[i:end]

				_, writeErr := c.Writer.Write(chunk)
				if writeErr != nil {
					return response, errors.New("wss_write_chunk_failed")
				}
			}
		}
	}

	return response, nil
}
