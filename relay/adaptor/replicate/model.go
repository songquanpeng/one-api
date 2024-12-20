package replicate

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"time"

	"github.com/pkg/errors"
)

type OpenaiImageEditRequest struct {
	Image          *multipart.FileHeader `json:"image" form:"image" binding:"required"`
	Prompt         string                `json:"prompt" form:"prompt" binding:"required"`
	Mask           *multipart.FileHeader `json:"mask" form:"mask" binding:"required"`
	Model          string                `json:"model" form:"model" binding:"required"`
	N              int                   `json:"n" form:"n" binding:"min=0,max=10"`
	Size           string                `json:"size" form:"size"`
	ResponseFormat string                `json:"response_format" form:"response_format"`
}

// toFluxRemixRequest convert OpenAI's image edit request to Flux's remix request.
//
// Note that the mask formats of OpenAI and Flux are different:
// OpenAI's mask sets the parts to be modified as transparent (0, 0, 0, 0),
// while Flux sets the parts to be modified as black (255, 255, 255, 255),
// so we need to convert the format here.
//
// Both OpenAI's Image and Mask are browser-native ImageData,
// which need to be converted to base64 dataURI format.
func (r *OpenaiImageEditRequest) toFluxRemixRequest() (*InpaintingImageByFlusReplicateRequest, error) {
	if r.ResponseFormat != "b64_json" {
		return nil, errors.New("response_format must be b64_json for replicate models")
	}

	fluxReq := &InpaintingImageByFlusReplicateRequest{
		Input: FluxInpaintingInput{
			Prompt:           r.Prompt,
			Seed:             int(time.Now().UnixNano()),
			Steps:            30,
			Guidance:         3,
			SafetyTolerance:  5,
			PromptUnsampling: false,
			OutputFormat:     "png",
		},
	}

	imgFile, err := r.Image.Open()
	if err != nil {
		return nil, errors.Wrap(err, "open image file")
	}
	defer imgFile.Close()
	imgData, err := io.ReadAll(imgFile)
	if err != nil {
		return nil, errors.Wrap(err, "read image file")
	}

	maskFile, err := r.Mask.Open()
	if err != nil {
		return nil, errors.Wrap(err, "open mask file")
	}
	defer maskFile.Close()

	// Convert image to base64
	imageBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(imgData)
	fluxReq.Input.Image = imageBase64

	// Convert mask data to RGBA
	maskPNG, err := png.Decode(maskFile)
	if err != nil {
		return nil, errors.Wrap(err, "decode mask file")
	}

	// convert mask to RGBA
	var maskRGBA *image.RGBA
	switch converted := maskPNG.(type) {
	case *image.RGBA:
		maskRGBA = converted
	default:
		// Convert to RGBA
		bounds := maskPNG.Bounds()
		maskRGBA = image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				maskRGBA.Set(x, y, maskPNG.At(x, y))
			}
		}
	}

	maskData := maskRGBA.Pix
	invertedMask := make([]byte, len(maskData))
	for i := 0; i+4 <= len(maskData); i += 4 {
		// If pixel is transparent (alpha = 0), make it black (255)
		if maskData[i+3] == 0 {
			invertedMask[i] = 255   // R
			invertedMask[i+1] = 255 // G
			invertedMask[i+2] = 255 // B
			invertedMask[i+3] = 255 // A
		} else {
			// Copy original pixel
			copy(invertedMask[i:i+4], maskData[i:i+4])
		}
	}

	// Convert inverted mask to base64 encoded png image
	invertedMaskRGBA := &image.RGBA{
		Pix:    invertedMask,
		Stride: maskRGBA.Stride,
		Rect:   maskRGBA.Rect,
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, invertedMaskRGBA)
	if err != nil {
		return nil, errors.Wrap(err, "encode inverted mask to png")
	}

	invertedMaskBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	fluxReq.Input.Mask = invertedMaskBase64

	return fluxReq, nil
}

// DrawImageRequest draw image by fluxpro
//
// https://replicate.com/black-forest-labs/flux-pro?prediction=kg1krwsdf9rg80ch1sgsrgq7h8&output=json
type DrawImageRequest struct {
	Input ImageInput `json:"input"`
}

// ImageInput is input of DrawImageByFluxProRequest
//
// https://replicate.com/black-forest-labs/flux-1.1-pro/api/schema
type ImageInput struct {
	Steps           int    `json:"steps" binding:"required,min=1"`
	Prompt          string `json:"prompt" binding:"required,min=5"`
	ImagePrompt     string `json:"image_prompt"`
	Guidance        int    `json:"guidance" binding:"required,min=2,max=5"`
	Interval        int    `json:"interval" binding:"required,min=1,max=4"`
	AspectRatio     string `json:"aspect_ratio" binding:"required,oneof=1:1 16:9 2:3 3:2 4:5 5:4 9:16"`
	SafetyTolerance int    `json:"safety_tolerance" binding:"required,min=1,max=5"`
	Seed            int    `json:"seed"`
	NImages         int    `json:"n_images" binding:"required,min=1,max=8"`
	Width           int    `json:"width" binding:"required,min=256,max=1440"`
	Height          int    `json:"height" binding:"required,min=256,max=1440"`
}

// InpaintingImageByFlusReplicateRequest is request to inpainting image by flux pro
//
// https://replicate.com/black-forest-labs/flux-fill-pro/api/schema
type InpaintingImageByFlusReplicateRequest struct {
	Input FluxInpaintingInput `json:"input"`
}

// FluxInpaintingInput is input of DrawImageByFluxProRequest
//
// https://replicate.com/black-forest-labs/flux-fill-pro/api/schema
type FluxInpaintingInput struct {
	Mask             string `json:"mask" binding:"required"`
	Image            string `json:"image" binding:"required"`
	Seed             int    `json:"seed"`
	Steps            int    `json:"steps" binding:"required,min=1"`
	Prompt           string `json:"prompt" binding:"required,min=5"`
	Guidance         int    `json:"guidance" binding:"required,min=2,max=5"`
	OutputFormat     string `json:"output_format"`
	SafetyTolerance  int    `json:"safety_tolerance" binding:"required,min=1,max=5"`
	PromptUnsampling bool   `json:"prompt_unsampling"`
}

// ImageResponse is response of DrawImageByFluxProRequest
//
// https://replicate.com/black-forest-labs/flux-pro?prediction=kg1krwsdf9rg80ch1sgsrgq7h8&output=json
type ImageResponse struct {
	CompletedAt time.Time        `json:"completed_at"`
	CreatedAt   time.Time        `json:"created_at"`
	DataRemoved bool             `json:"data_removed"`
	Error       string           `json:"error"`
	ID          string           `json:"id"`
	Input       DrawImageRequest `json:"input"`
	Logs        string           `json:"logs"`
	Metrics     FluxMetrics      `json:"metrics"`
	// Output could be `string` or `[]string`
	Output    any       `json:"output"`
	StartedAt time.Time `json:"started_at"`
	Status    string    `json:"status"`
	URLs      FluxURLs  `json:"urls"`
	Version   string    `json:"version"`
}

func (r *ImageResponse) GetOutput() ([]string, error) {
	switch v := r.Output.(type) {
	case string:
		return []string{v}, nil
	case []string:
		return v, nil
	case nil:
		return nil, nil
	case []interface{}:
		// convert []interface{} to []string
		ret := make([]string, len(v))
		for idx, vv := range v {
			if vvv, ok := vv.(string); ok {
				ret[idx] = vvv
			} else {
				return nil, errors.Errorf("unknown output type: [%T]%v", vv, vv)
			}
		}

		return ret, nil
	default:
		return nil, errors.Errorf("unknown output type: [%T]%v", r.Output, r.Output)
	}
}

// FluxMetrics is metrics of ImageResponse
type FluxMetrics struct {
	ImageCount  int     `json:"image_count"`
	PredictTime float64 `json:"predict_time"`
	TotalTime   float64 `json:"total_time"`
}

// FluxURLs is urls of ImageResponse
type FluxURLs struct {
	Get    string `json:"get"`
	Cancel string `json:"cancel"`
}
