package replicate

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/render"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

func ChatHandler(c *gin.Context, resp *http.Response) (
	srvErr *model.ErrorWithStatusCode, usage *model.Usage) {
	if resp.StatusCode != http.StatusCreated {
		payload, _ := io.ReadAll(resp.Body)
		return openai.ErrorWrapper(
				errors.Errorf("bad_status_code [%d]%s", resp.StatusCode, string(payload)),
				"bad_status_code", http.StatusInternalServerError),
			nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}

	respData := new(ChatResponse)
	if err = json.Unmarshal(respBody, respData); err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	for {
		err = func() error {
			// get task
			taskReq, err := http.NewRequestWithContext(c.Request.Context(),
				http.MethodGet, respData.URLs.Get, nil)
			if err != nil {
				return errors.Wrap(err, "new request")
			}

			taskReq.Header.Set("Authorization", "Bearer "+meta.GetByContext(c).APIKey)
			taskResp, err := http.DefaultClient.Do(taskReq)
			if err != nil {
				return errors.Wrap(err, "get task")
			}
			defer taskResp.Body.Close()

			if taskResp.StatusCode != http.StatusOK {
				payload, _ := io.ReadAll(taskResp.Body)
				return errors.Errorf("bad status code [%d]%s",
					taskResp.StatusCode, string(payload))
			}

			taskBody, err := io.ReadAll(taskResp.Body)
			if err != nil {
				return errors.Wrap(err, "read task response")
			}

			taskData := new(ChatResponse)
			if err = json.Unmarshal(taskBody, taskData); err != nil {
				return errors.Wrap(err, "decode task response")
			}

			switch taskData.Status {
			case "succeeded":
			case "failed", "canceled":
				return errors.Errorf("task failed, [%s]%s", taskData.Status, taskData.Error)
			default:
				time.Sleep(time.Second * 3)
				return errNextLoop
			}

			if taskData.URLs.Stream == "" {
				return errors.New("stream url is empty")
			}

			// request stream url
			responseText, err := chatStreamHandler(c, taskData.URLs.Stream)
			if err != nil {
				return errors.Wrap(err, "chat stream handler")
			}

			ctxMeta := meta.GetByContext(c)
			usage = openai.ResponseText2Usage(responseText,
				ctxMeta.ActualModelName, ctxMeta.PromptTokens)
			return nil
		}()
		if err != nil {
			if errors.Is(err, errNextLoop) {
				continue
			}

			return openai.ErrorWrapper(err, "chat_task_failed", http.StatusInternalServerError), nil
		}

		break
	}

	return nil, usage
}

const (
	eventPrefix = "event: "
	dataPrefix  = "data: "
	done        = "[DONE]"
)

func chatStreamHandler(c *gin.Context, streamUrl string) (responseText string, err error) {
	// request stream endpoint
	streamReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, streamUrl, nil)
	if err != nil {
		return "", errors.Wrap(err, "new request to stream")
	}

	streamReq.Header.Set("Authorization", "Bearer "+meta.GetByContext(c).APIKey)
	streamReq.Header.Set("Accept", "text/event-stream")
	streamReq.Header.Set("Cache-Control", "no-store")

	resp, err := http.DefaultClient.Do(streamReq)
	if err != nil {
		return "", errors.Wrap(err, "do request to stream")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(resp.Body)
		return "", errors.Errorf("bad status code [%d]%s", resp.StatusCode, string(payload))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	common.SetEventStreamHeaders(c)
	doneRendered := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Handle comments starting with ':'
		if strings.HasPrefix(line, ":") {
			continue
		}

		// Parse SSE fields
		if strings.HasPrefix(line, eventPrefix) {
			event := strings.TrimSpace(line[len(eventPrefix):])
			var data string
			// Read the following lines to get data and id
			for scanner.Scan() {
				nextLine := scanner.Text()
				if nextLine == "" {
					break
				}
				if strings.HasPrefix(nextLine, dataPrefix) {
					data = nextLine[len(dataPrefix):]
				} else if strings.HasPrefix(nextLine, "id:") {
					// id = strings.TrimSpace(nextLine[len("id:"):])
				}
			}

			if event == "output" {
				render.StringData(c, data)
				responseText += data
			} else if event == "done" {
				render.Done(c)
				doneRendered = true
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", errors.Wrap(err, "scan stream")
	}

	if !doneRendered {
		render.Done(c)
	}

	return responseText, nil
}
