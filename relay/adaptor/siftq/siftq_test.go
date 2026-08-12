package siftq

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/meta"
)

func strptr(value string) *string { return &value }
func intptr(value int) *int       { return &value }

func validTextRequest() *CreateRequest {
	return &CreateRequest{
		Model:      ModelName,
		Content:    []ContentItem{{Type: "text", Text: "A cinematic sunrise."}},
		Resolution: strptr("2K"), Duration: intptr(5), Ratio: strptr("16:9"),
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct{ base, path, want string }{
		{"", CreateVideoPath, "https://siftq.com/api/minimax/v2/video_generation"},
		{DefaultBaseURL, QueryVideoPath, "https://siftq.com/api/minimax/v2/query/video_generation"},
		{"https://gateway.example/siftq", CreateContextPath, "https://gateway.example/siftq/v2/h3_context_ir"},
		{"https://gateway.example/siftq/", CreateVideoPath + "/task", "https://gateway.example/siftq/v2/video_generation/task"},
	}
	for _, test := range tests {
		got, err := BuildURL(test.base, test.path)
		if err != nil || got != test.want {
			t.Fatalf("BuildURL(%q, %q) = %q, %v; want %q", test.base, test.path, got, err, test.want)
		}
	}
	for _, invalid := range []string{"not-a-url", "ftp://example.com", "https://user:pass@example.com", "https://example.com?key=secret"} {
		if _, err := BuildURL(invalid, CreateVideoPath); err == nil {
			t.Errorf("BuildURL accepted invalid base %q", invalid)
		}
	}
}

func TestBuildListURLUsesRepeatedDottedFilters(t *testing.T) {
	raw, err := BuildListURL(DefaultBaseURL, ListFilters{
		PageNum: 2, PageSize: 10, Status: "succeeded",
		TaskIDs: []string{"task-1", "task-2"}, Model: ModelName, TaskType: "generation",
	})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Path != "/api/minimax/v2/query/video_generation" {
		t.Fatalf("unexpected path %q", parsed.Path)
	}
	query := parsed.Query()
	if got := query["filter.task_ids"]; len(got) != 2 || got[0] != "task-1" || got[1] != "task-2" {
		t.Fatalf("filter.task_ids = %#v", got)
	}
	if query.Get("filter.status") != "succeeded" || query.Get("filter.model") != ModelName || query.Get("filter.task_type") != "generation" {
		t.Fatalf("missing dotted filters: %s", parsed.RawQuery)
	}
}

func TestNewUpstreamRequestHeadersAndCancellation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	c.Request, _ = http.NewRequest(http.MethodPost, "http://one-api.test/v2/video_generation", nil)
	providerMeta := &meta.Meta{APIKey: "runtime-secret"}
	request, err := NewUpstreamRequest(c, providerMeta, http.MethodPost, "https://example.com/v2/video_generation", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	if request.Header.Get("Authorization") != "Bearer runtime-secret" {
		t.Fatal("missing bearer authorization")
	}
	if request.Header.Get("Content-Type") != "application/json" {
		t.Fatal("missing JSON content type")
	}
	if _, err := NewUpstreamRequest(c, &meta.Meta{}, http.MethodGet, "https://example.com", nil); err == nil {
		t.Fatal("missing API key was accepted")
	}
	requestContext, cancel := context.WithCancel(c.Request.Context())
	c.Request = c.Request.WithContext(requestContext)
	cancel()
	cancelledRequest, err := NewUpstreamRequest(c, providerMeta, http.MethodGet, "https://example.com/v2/query/video_generation/task", nil)
	if err != nil {
		t.Fatal(err)
	}
	if cancelledRequest.Context().Err() != context.Canceled {
		t.Fatal("request cancellation was not propagated")
	}
}

func TestValidateCreateRequestModes(t *testing.T) {
	if err := ValidateCreateRequest(validTextRequest(), false); err != nil {
		t.Fatal(err)
	}

	image := validTextRequest()
	image.Content = append(image.Content, ContentItem{Type: "image_url", ImageURL: &URLValue{URL: "https://example.com/frame.png"}})
	image.Ratio = nil
	if err := ValidateCreateRequest(image, false); err != nil {
		t.Fatal(err)
	}
	if image.Content[1].Role != "first_frame" || image.Ratio == nil || *image.Ratio != "adaptive" {
		t.Fatalf("single image was not normalized: %#v", image)
	}

	frames := validTextRequest()
	frames.Ratio = strptr("adaptive")
	frames.Content = append(frames.Content,
		ContentItem{Type: "image_url", ImageURL: &URLValue{URL: "mm_file://first"}, Role: "first_frame"},
		ContentItem{Type: "image_url", ImageURL: &URLValue{URL: "mm_file://last"}, Role: "last_frame"})
	if err := ValidateCreateRequest(frames, false); err != nil {
		t.Fatal(err)
	}

	references := validTextRequest()
	references.Ratio = nil
	references.Content = append(references.Content,
		ContentItem{Type: "image_url", ImageURL: &URLValue{URL: "https://example.com/ref.webp"}, Role: "reference_image"},
		ContentItem{Type: "video_url", VideoURL: &URLValue{URL: "https://example.com/ref.mp4"}, Role: "reference_video"},
		ContentItem{Type: "audio_url", AudioURL: &URLValue{URL: "data:audio/mp3;base64,YQ=="}, Role: "reference_audio"})
	if err := ValidateCreateRequest(references, false); err != nil {
		t.Fatal(err)
	}
	if references.Ratio == nil || *references.Ratio != "adaptive" {
		t.Fatal("reference ratio was not defaulted")
	}

	contextIR := validTextRequest()
	contextIR.Resolution = nil
	if err := ValidateCreateRequest(contextIR, true); err != nil {
		t.Fatal(err)
	}
}

func TestValidateCreateRequestRejectsBoundaries(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CreateRequest)
	}{
		{"wrong model", func(r *CreateRequest) { r.Model = "video-01" }},
		{"missing duration", func(r *CreateRequest) { r.Duration = nil }},
		{"zero duration", func(r *CreateRequest) { r.Duration = intptr(0) }},
		{"duration too low", func(r *CreateRequest) { r.Duration = intptr(3) }},
		{"duration too high", func(r *CreateRequest) { r.Duration = intptr(16) }},
		{"bad resolution", func(r *CreateRequest) { r.Resolution = strptr("1080P") }},
		{"adaptive text", func(r *CreateRequest) { r.Ratio = strptr("adaptive") }},
		{"empty prompt", func(r *CreateRequest) { r.Content[0].Text = " " }},
		{"long prompt", func(r *CreateRequest) { r.Content[0].Text = strings.Repeat("字", 7001) }},
		{"bad callback", func(r *CreateRequest) { r.CallbackURL = "javascript:alert(1)" }},
		{"mismatched media", func(r *CreateRequest) {
			r.Content = append(r.Content, ContentItem{Type: "audio_url", ImageURL: &URLValue{URL: "https://example.com/x"}, Role: "reference_audio"})
		}},
		{"last without first", func(r *CreateRequest) {
			r.Ratio = strptr("adaptive")
			r.Content = append(r.Content, ContentItem{Type: "image_url", ImageURL: &URLValue{URL: "https://example.com/last.png"}, Role: "last_frame"})
		}},
		{"mixed frame and reference", func(r *CreateRequest) {
			r.Ratio = strptr("adaptive")
			r.Content = append(r.Content, ContentItem{Type: "image_url", ImageURL: &URLValue{URL: "https://example.com/first.png"}, Role: "first_frame"}, ContentItem{Type: "audio_url", AudioURL: &URLValue{URL: "https://example.com/ref.mp3"}, Role: "reference_audio"})
		}},
		{"bad data MIME", func(r *CreateRequest) {
			r.Ratio = strptr("adaptive")
			r.Content = append(r.Content, ContentItem{Type: "image_url", ImageURL: &URLValue{URL: "data:Image/PNG;base64,YQ=="}, Role: "reference_image"})
		}},
		{"too many reference images", func(r *CreateRequest) {
			r.Ratio = strptr("adaptive")
			for index := 0; index < 10; index++ {
				r.Content = append(r.Content, ContentItem{Type: "image_url", ImageURL: &URLValue{URL: "https://example.com/ref.png"}, Role: "reference_image"})
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := validTextRequest()
			test.mutate(request)
			if err := ValidateCreateRequest(request, false); err == nil {
				t.Fatal("invalid request was accepted")
			}
		})
	}

	contextIR := validTextRequest()
	if err := ValidateCreateRequest(contextIR, true); err == nil {
		t.Fatal("Context-IR accepted resolution")
	}
}

func TestValidateCreateAndDeleteResponses(t *testing.T) {
	if err := ValidateCreateResponse(&CreateResponse{TaskID: "task-1"}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateCreateResponse(&CreateResponse{}); err == nil {
		t.Fatal("missing task_id was accepted")
	}
	for _, action := range []string{"cancelled", "deleted"} {
		if err := ValidateDeleteResponse(&DeleteResponse{TaskID: "task-1", Action: action, Status: action}, "task-1"); err != nil {
			t.Fatal(err)
		}
	}
	invalid := []DeleteResponse{
		{TaskID: "other", Action: "deleted", Status: "deleted"},
		{TaskID: "task-1", Action: "running", Status: "running"},
		{TaskID: "task-1", Action: "cancelled", Status: "deleted"},
	}
	for index := range invalid {
		if err := ValidateDeleteResponse(&invalid[index], "task-1"); err == nil {
			t.Errorf("invalid delete response %d accepted", index)
		}
	}
}

func TestValidateTaskStates(t *testing.T) {
	valid := []VideoTask{
		{ID: "1", Status: "queued"},
		{ID: "1", Status: "running"},
		{ID: "1", Status: "cancelled"},
		{ID: "1", Status: "failed", Error: &TaskError{Code: "policy", Message: "rejected"}},
		{ID: "1", Status: "succeeded", TaskType: "generation", Modality: "video", Content: &TaskContent{URL: "https://example.com/out.mp4"}},
		{ID: "1", Status: "succeeded", TaskType: "h3_context_ir", Modality: "text", Content: &TaskContent{Prompt: "enhanced"}, Usage: &TaskUsage{TotalTokens: 12}},
	}
	for index := range valid {
		if err := ValidateTask(&valid[index]); err != nil {
			t.Errorf("valid task %d: %v", index, err)
		}
	}
	invalid := []VideoTask{
		{Status: "queued"},
		{ID: "1", Status: "FAILED"},
		{ID: "1", Status: "failed"},
		{ID: "1", Status: "succeeded", Content: &TaskContent{}},
		{ID: "1", Status: "succeeded", TaskType: "h3_context_ir", Modality: "text", Content: &TaskContent{}},
	}
	for index := range invalid {
		if err := ValidateTask(&invalid[index]); err == nil {
			t.Errorf("invalid task %d was accepted", index)
		}
	}
}

func TestCreateJSONUsesH3V2Fields(t *testing.T) {
	request := validTextRequest()
	request.CallbackURL = "https://example.com/callback"
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, required := range []string{`"model":"MiniMax-H3"`, `"content"`, `"resolution":"2K"`, `"duration":5`, `"ratio":"16:9"`, `"callback_url":"https://example.com/callback"`} {
		if !strings.Contains(text, required) {
			t.Errorf("missing %s in %s", required, text)
		}
	}
	for _, legacy := range []string{"base_resp", "file_id", "/v1/video_generation"} {
		if strings.Contains(text, legacy) {
			t.Errorf("legacy V1 field %q found", legacy)
		}
	}
}
