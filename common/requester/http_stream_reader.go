package requester

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
)

// 流处理函数，判断依据如下：
// 1.如果有错误信息，则直接返回错误信息
// 2.如果isFinished=true，则返回io.EOF，并且如果response不为空，还将返回response
// 3.如果rawLine=nil 或者 response长度为0，则直接跳过
// 4.如果以上条件都不满足，则返回response
type HandlerPrefix[T streamable] func(rawLine *[]byte, isFinished *bool, response *[]T) error

type streamable interface {
	// types.ChatCompletionStreamResponse | types.CompletionResponse
	any
}

type StreamReaderInterface[T streamable] interface {
	Recv() (*[]T, error)
	Close()
}

type streamReader[T streamable] struct {
	isFinished bool

	reader   *bufio.Reader
	response *http.Response

	handlerPrefix HandlerPrefix[T]
}

func (stream *streamReader[T]) Recv() (response *[]T, err error) {
	if stream.isFinished {
		err = io.EOF
		return
	}
	response, err = stream.processLines()
	return
}

//nolint:gocognit
func (stream *streamReader[T]) processLines() (*[]T, error) {
	for {
		rawLine, readErr := stream.reader.ReadBytes('\n')
		if readErr != nil {
			return nil, readErr
		}

		noSpaceLine := bytes.TrimSpace(rawLine)

		var response []T
		err := stream.handlerPrefix(&noSpaceLine, &stream.isFinished, &response)

		if err != nil {
			return nil, err
		}

		if stream.isFinished {
			if len(response) > 0 {
				return &response, io.EOF
			}
			return nil, io.EOF
		}

		if noSpaceLine == nil || len(response) == 0 {
			continue
		}

		return &response, nil
	}
}

func (stream *streamReader[T]) Close() {
	stream.response.Body.Close()
}
