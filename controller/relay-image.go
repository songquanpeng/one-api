package controller

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func relayImageHelper(c *gin.Context, relayMode int) *OpenAIErrorWithStatusCode {
	// TODO: this part is not finished
	req, err := http.NewRequest(c.Request.Method, c.Request.RequestURI, c.Request.Body)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errorWrapper(err, "do_request_failed", http.StatusOK)
	}
	err = req.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_request_body_failed", http.StatusOK)
	}
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return errorWrapper(err, "copy_response_body_failed", http.StatusOK)
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusOK)
	}
	return nil
}
