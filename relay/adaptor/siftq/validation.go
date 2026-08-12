package siftq

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var validRatios = map[string]bool{
	"adaptive": true, "21:9": true, "16:9": true, "4:3": true,
	"1:1": true, "3:4": true, "9:16": true,
}

var validRoles = map[string]bool{
	"first_frame": true, "last_frame": true, "reference_image": true,
	"reference_video": true, "reference_audio": true,
}

func ValidateCreateRequest(request *CreateRequest, contextIR bool) error {
	if request == nil {
		return errors.New("request is required")
	}
	if request.Model != ModelName {
		return fmt.Errorf("model must be %s", ModelName)
	}
	if request.Duration == nil || *request.Duration < 4 || *request.Duration > 15 {
		return errors.New("duration must be an integer from 4 through 15")
	}
	if contextIR {
		if request.Resolution != nil {
			return errors.New("resolution is not supported for h3_context_ir")
		}
	} else if request.Resolution == nil || (*request.Resolution != "768P" && *request.Resolution != "2K") {
		return errors.New("resolution must be 768P or 2K")
	}
	if request.Ratio != nil && !validRatios[*request.Ratio] {
		return errors.New("ratio is invalid")
	}
	if request.CallbackURL != "" {
		parsed, err := url.ParseRequestURI(request.CallbackURL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return errors.New("callback_url must be an absolute HTTP or HTTPS URL")
		}
	}
	if len(request.Content) == 0 {
		return errors.New("content is required")
	}

	var promptFound, firstFrame, lastFrame, hasReferences bool
	imageCount, referenceImageCount, firstFrameCount, lastFrameCount := 0, 0, 0, 0
	videoCount, audioCount := 0, 0
	unroledImage := -1
	for index := range request.Content {
		item := &request.Content[index]
		if item.Role != "" && !validRoles[item.Role] {
			return fmt.Errorf("content[%d].role is invalid", index)
		}
		switch item.Type {
		case "text":
			if strings.TrimSpace(item.Text) == "" {
				return fmt.Errorf("content[%d].text must not be empty", index)
			}
			if len([]rune(item.Text)) > 7000 {
				return fmt.Errorf("content[%d].text exceeds 7000 characters", index)
			}
			if item.Role != "" || item.ImageURL != nil || item.VideoURL != nil || item.AudioURL != nil {
				return fmt.Errorf("content[%d] has fields that do not match type text", index)
			}
			promptFound = true
		case "image_url":
			if item.ImageURL == nil || item.VideoURL != nil || item.AudioURL != nil || item.Text != "" {
				return fmt.Errorf("content[%d] must contain only image_url", index)
			}
			if err := validateMediaURL(item.ImageURL.URL, "image", 30<<20); err != nil {
				return fmt.Errorf("content[%d]: %w", index, err)
			}
			imageCount++
			switch item.Role {
			case "":
				unroledImage = index
			case "first_frame":
				firstFrame = true
				firstFrameCount++
			case "last_frame":
				lastFrame = true
				lastFrameCount++
			case "reference_image":
				hasReferences = true
				referenceImageCount++
			default:
				return fmt.Errorf("content[%d].role does not match image_url", index)
			}
		case "video_url":
			if item.VideoURL == nil || item.ImageURL != nil || item.AudioURL != nil || item.Text != "" || item.Role != "reference_video" {
				return fmt.Errorf("content[%d] must be a reference_video video_url", index)
			}
			if err := validateMediaURL(item.VideoURL.URL, "video", 50<<20); err != nil {
				return fmt.Errorf("content[%d]: %w", index, err)
			}
			videoCount++
			hasReferences = true
		case "audio_url":
			if item.AudioURL == nil || item.ImageURL != nil || item.VideoURL != nil || item.Text != "" || item.Role != "reference_audio" {
				return fmt.Errorf("content[%d] must be a reference_audio audio_url", index)
			}
			if err := validateMediaURL(item.AudioURL.URL, "audio", 15<<20); err != nil {
				return fmt.Errorf("content[%d]: %w", index, err)
			}
			audioCount++
			hasReferences = true
		default:
			return fmt.Errorf("content[%d].type is invalid", index)
		}
	}
	if !promptFound {
		return errors.New("content must include a non-empty text prompt")
	}
	if unroledImage >= 0 {
		if imageCount != 1 || firstFrame || lastFrame || hasReferences {
			return errors.New("an image without role is allowed only as the single first frame")
		}
		request.Content[unroledImage].Role = "first_frame"
		firstFrame = true
		firstFrameCount++
	}
	if firstFrameCount > 1 || lastFrameCount > 1 {
		return errors.New("first_frame and last_frame each allow at most one image")
	}
	if lastFrame && !firstFrame {
		return errors.New("last_frame requires first_frame")
	}
	if (firstFrame || lastFrame) && hasReferences {
		return errors.New("first/last frame roles cannot be combined with reference roles")
	}
	if referenceImageCount > 9 || videoCount > 3 || audioCount > 3 {
		return errors.New("reference media count exceeds provider limits")
	}
	if firstFrame || lastFrame {
		if request.Ratio == nil {
			ratio := "adaptive"
			request.Ratio = &ratio
		} else if *request.Ratio != "adaptive" {
			return errors.New("first/last-frame requests must use adaptive ratio")
		}
	} else if !hasReferences {
		if request.Ratio == nil || *request.Ratio == "adaptive" {
			return errors.New("text-only requests require a concrete ratio")
		}
	} else if request.Ratio == nil {
		ratio := "adaptive"
		request.Ratio = &ratio
	}
	return nil
}

func validateMediaURL(raw, mediaType string, maxDecodedBytes int) error {
	if strings.HasPrefix(raw, "mm_file://") {
		if len(strings.TrimPrefix(raw, "mm_file://")) == 0 {
			return errors.New("mm_file reference is missing a file id")
		}
		return nil
	}
	if strings.HasPrefix(raw, "data:") {
		parts := strings.SplitN(raw, ",", 2)
		if len(parts) != 2 || !strings.HasSuffix(parts[0], ";base64") {
			return errors.New("media data URI must be base64 encoded")
		}
		mime := strings.TrimPrefix(strings.TrimSuffix(parts[0], ";base64"), "data:")
		if mime != strings.ToLower(mime) || !strings.HasPrefix(mime, mediaType+"/") {
			return fmt.Errorf("data URI MIME must be lowercase %s/*", mediaType)
		}
		allowedMIME := map[string]map[string]bool{
			"image": {"image/jpeg": true, "image/jpg": true, "image/png": true, "image/webp": true, "image/heic": true, "image/heif": true},
			"video": {"video/mp4": true, "video/quicktime": true, "video/mov": true},
			"audio": {"audio/wav": true, "audio/wave": true, "audio/x-wav": true, "audio/mp3": true, "audio/mpeg": true},
		}
		if !allowedMIME[mediaType][mime] {
			return fmt.Errorf("unsupported %s data URI format", mediaType)
		}
		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return errors.New("media data URI has invalid base64")
		}
		if len(decoded) > maxDecodedBytes {
			return fmt.Errorf("%s exceeds the per-file size limit", mediaType)
		}
		return nil
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("media URL must be public HTTP(S), mm_file, or a data URI")
	}
	return nil
}

func ValidateTask(task *VideoTask) error {
	if task == nil || task.ID == "" {
		return errors.New("task.id is required")
	}
	switch task.Status {
	case "queued", "running", "cancelled":
		return nil
	case "failed":
		if task.Error == nil || task.Error.Message == "" {
			return errors.New("failed task is missing error details")
		}
		return nil
	case "succeeded":
		if task.Content == nil {
			return errors.New("succeeded task is missing content")
		}
		if task.TaskType == "h3_context_ir" || task.Modality == "text" {
			if strings.TrimSpace(task.Content.Prompt) == "" {
				return errors.New("succeeded context task is missing content.prompt")
			}
		} else if strings.TrimSpace(task.Content.URL) == "" {
			return errors.New("succeeded video task is missing content.url")
		}
		return nil
	default:
		return fmt.Errorf("unexpected task status %q", task.Status)
	}
}

func ValidateCreateResponse(response *CreateResponse) error {
	if response == nil || strings.TrimSpace(response.TaskID) == "" {
		return errors.New("create response is missing task_id")
	}
	return nil
}

func ValidateDeleteResponse(response *DeleteResponse, taskID string) error {
	if response == nil || response.TaskID != taskID {
		return errors.New("delete response task_id does not match")
	}
	if response.Action != "cancelled" && response.Action != "deleted" {
		return errors.New("delete response action is invalid")
	}
	if response.Status != response.Action {
		return errors.New("delete response status does not match action")
	}
	return nil
}
