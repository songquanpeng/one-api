package common

import (
	"bytes"
	"io"
	"net/http"
)

func CloneRequest(old *http.Request) *http.Request {
	req := old.Clone(old.Context())
	oldBody, err := io.ReadAll(old.Body)
	if err != nil {
		return nil
	}
	err = old.Body.Close()
	if err != nil {
		return nil
	}
	old.Body = io.NopCloser(bytes.NewBuffer(oldBody))
	req.Body = io.NopCloser(bytes.NewBuffer(oldBody))
	return req
}
