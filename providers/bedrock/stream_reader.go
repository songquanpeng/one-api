package bedrock

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/common/requester"
	"one-api/types"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"
	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream/eventstreamapi"
	"github.com/aws/smithy-go"
)

type streamReader[T any] struct {
	reader   *bufio.Reader
	response *http.Response

	handlerPrefix requester.HandlerPrefix[T]

	DataChan chan T
	ErrChan  chan error
}

func (stream *streamReader[T]) Recv() (<-chan T, <-chan error) {
	go stream.processLines()

	return stream.DataChan, stream.ErrChan
}

//nolint:gocognit
func (stream *streamReader[T]) processLines() {
	decode := eventstream.NewDecoder()
	payloadBuf := make([]byte, 0*1024)
	for {
		payloadBuf = payloadBuf[0:0]
		messgae, readErr := decode.Decode(stream.reader, payloadBuf)
		if readErr != nil {
			stream.ErrChan <- readErr
			return
		}

		line, err := stream.deserializeEventMessage(&messgae)
		if err != nil {
			stream.ErrChan <- common.ErrorWrapper(err, "decode_response_failed", http.StatusInternalServerError)
			return
		}

		stream.handlerPrefix(&line, stream.DataChan, stream.ErrChan)

		if line == nil {
			continue
		}

		if bytes.Equal(line, requester.StreamClosed) {
			return
		}
	}
}

func (stream *streamReader[T]) Close() {
	stream.response.Body.Close()
}

func (stream *streamReader[T]) deserializeEventMessage(msg *eventstream.Message) ([]byte, error) {
	messageType := msg.Headers.Get(eventstreamapi.MessageTypeHeader)
	if messageType == nil {
		return nil, fmt.Errorf("%s event header not present", eventstreamapi.MessageTypeHeader)
	}

	switch messageType.String() {
	case eventstreamapi.EventMessageType:
		var v BedrockResponseStream
		if err := json.Unmarshal(msg.Payload, &v); err != nil {
			return nil, err
		}
		buffer, err := base64.StdEncoding.DecodeString(v.Bytes)
		if err != nil {
			return nil, err
		}
		return buffer, nil

	case eventstreamapi.ExceptionMessageType:
		exceptionType := msg.Headers.Get(eventstreamapi.ExceptionTypeHeader)
		return nil, errors.New("Exception message :" + exceptionType.String())

	case eventstreamapi.ErrorMessageType:
		errorCode := "UnknownError"
		errorMessage := errorCode
		if header := msg.Headers.Get(eventstreamapi.ErrorCodeHeader); header != nil {
			errorCode = header.String()
		}
		if header := msg.Headers.Get(eventstreamapi.ErrorMessageHeader); header != nil {
			errorMessage = header.String()
		}
		return nil, &smithy.GenericAPIError{
			Code:    errorCode,
			Message: errorMessage,
		}

	default:
		return nil, errors.New("bedrock stream unknown error")
	}
}

func RequestStream[T any](resp *http.Response, handlerPrefix requester.HandlerPrefix[T]) (*streamReader[T], *types.OpenAIErrorWithStatusCode) {
	// 如果返回的头是json格式 说明有错误
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		return nil, requester.HandleErrorResp(resp, requestErrorHandle, true)
	}

	stream := &streamReader[T]{
		reader:        bufio.NewReader(resp.Body),
		response:      resp,
		handlerPrefix: handlerPrefix,

		DataChan: make(chan T),
		ErrChan:  make(chan error),
	}

	return stream, nil
}
