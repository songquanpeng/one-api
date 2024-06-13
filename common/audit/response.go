package audit

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type AuditLogger struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (l *AuditLogger) Write(p []byte) (int, error) {
	l.buf.Write(p)
	return l.ResponseWriter.Write(p)
}

func CaptureResponseBody(c *gin.Context) *bytes.Buffer {
	al := &AuditLogger{
		ResponseWriter: c.Writer,
		buf:            &bytes.Buffer{},
	}
	c.Writer = al
	return al.buf
}

func B64encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

type AuditReadCloser struct {
	Reader io.Reader
	Closer io.Closer
	Buffer *bytes.Buffer
}

func (arc *AuditReadCloser) Read(p []byte) (int, error) {
	n, err := arc.Reader.Read(p)
	if n > 0 {
		arc.Buffer.Write(p[:n])
	}
	return n, err
}

func (arc *AuditReadCloser) Close() error {
	return arc.Closer.Close()
}

func CaptureHTTPResponseBody(resp *http.Response) *bytes.Buffer {
	buf := &bytes.Buffer{}
	arc := &AuditReadCloser{
		Reader: resp.Body,
		Closer: resp.Body,
		Buffer: buf,
	}
	resp.Body = arc
	return buf
}

func ParseOPENAIStreamResponse(buf *bytes.Buffer) string {
	lines := strings.Split(buf.String(), "\n")
	bts := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.Trim(line, "\n")
		if strings.HasPrefix(string(line), "data:") {
			line = line[5:]
		}
		content := gjson.Get(line, "choices.0.delta.content").String()
		bts = append(bts, content)
	}
	return strings.Join(bts, "")
}
