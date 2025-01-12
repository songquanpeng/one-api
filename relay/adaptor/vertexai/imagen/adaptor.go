package imagen

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

var ModelList = []string{
	"imagen-3.0-generate-001",
}

type Adaptor struct {
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	meta := meta.GetByContext(c)

	if request.ResponseFormat != "b64_json" {
		return nil, errors.New("only support b64_json response format")
	}
	if request.N <= 0 {
		return nil, errors.New("n must be greater than 0")
	}

	switch meta.Mode {
	case relaymode.ImagesGenerations:
		return convertImageCreateRequest(request)
	default:
		return nil, errors.New("not implemented")
	}
}

func convertImageCreateRequest(request *model.ImageRequest) (any, error) {
	return CreateImageRequest{
		Instances: []createImageInstance{
			{
				Prompt: request.Prompt,
			},
		},
		Parameters: createImageParameters{
			SampleCount: request.N,
		},
	}, nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	return nil, errors.New("not implemented")
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, wrapErr *model.ErrorWithStatusCode) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, openai.ErrorWrapper(
			errors.Wrap(err, "failed to read response body"),
			"read_response_body",
			http.StatusInternalServerError,
		)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, openai.ErrorWrapper(
			errors.Errorf("upstream response status code: %d, body: %s", resp.StatusCode, string(respBody)),
			"upstream_response",
			http.StatusInternalServerError,
		)
	}

	imagenResp := new(CreateImageResponse)
	if err := json.Unmarshal(respBody, imagenResp); err != nil {
		return nil, openai.ErrorWrapper(
			errors.Wrap(err, "failed to decode response body"),
			"unmarshal_upstream_response",
			http.StatusInternalServerError,
		)
	}

	if len(imagenResp.Predictions) == 0 {
		return nil, openai.ErrorWrapper(
			errors.New("empty predictions"),
			"empty_predictions",
			http.StatusInternalServerError,
		)
	}

	oaiResp := openai.ImageResponse{
		Created: time.Now().Unix(),
	}
	for _, prediction := range imagenResp.Predictions {
		oaiResp.Data = append(oaiResp.Data, openai.ImageData{
			B64Json: prediction.BytesBase64Encoded,
		})
	}

	c.JSON(http.StatusOK, oaiResp)
	return nil, nil
}
