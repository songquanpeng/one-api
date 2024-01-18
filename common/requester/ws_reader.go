package requester

import (
	"io"

	"github.com/gorilla/websocket"
)

type wsReader[T streamable] struct {
	isFinished bool

	reader        *websocket.Conn
	handlerPrefix HandlerPrefix[T]
}

func (stream *wsReader[T]) Recv() (response *[]T, err error) {
	if stream.isFinished {
		err = io.EOF
		return
	}

	response, err = stream.processLines()
	return
}

func (stream *wsReader[T]) processLines() (*[]T, error) {
	for {
		_, msg, err := stream.reader.ReadMessage()
		if err != nil {
			return nil, err
		}

		var response []T
		err = stream.handlerPrefix(&msg, &stream.isFinished, &response)

		if err != nil {
			return nil, err
		}

		if stream.isFinished {
			if len(response) > 0 {
				return &response, io.EOF
			}
			return nil, io.EOF
		}

		if msg == nil || len(response) == 0 {
			continue
		}

		return &response, nil

	}
}

func (stream *wsReader[T]) Close() {
	stream.reader.Close()
}
