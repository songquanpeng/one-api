package controller

import (
	"testing"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/adaptor/siftq"
)

func TestConvertVideoCreateRequest(t *testing.T) {
	request := &videoCreateRequest{Prompt: "a paper boat on a river", Model: siftq.ModelName, Seconds: "8", Size: "1280x720"}
	converted, seconds, size, err := convertVideoCreateRequest(request, "")
	if err != nil {
		t.Fatal(err)
	}
	if seconds != 8 || size != "1280x720" || *converted.Resolution != "768P" || *converted.Ratio != "16:9" {
		t.Fatalf("unexpected conversion: %#v, seconds=%d size=%s", converted, seconds, size)
	}
	if converted.Content[0].Type != "text" || converted.Content[0].Text != request.Prompt {
		t.Fatalf("prompt was not converted: %#v", converted.Content)
	}
}

func TestConvertVideoCreateRequestWithReference(t *testing.T) {
	request := &videoCreateRequest{Prompt: "camera slowly moves forward", Model: siftq.ModelName, Seconds: float64(5), Size: "1792x1024"}
	converted, _, _, err := convertVideoCreateRequest(request, "https://example.com/first.png")
	if err != nil {
		t.Fatal(err)
	}
	if len(converted.Content) != 2 || converted.Content[1].Role != "first_frame" || *converted.Ratio != "adaptive" {
		t.Fatalf("reference image was not converted as first frame: %#v", converted)
	}
}

func TestVideoRequestValidation(t *testing.T) {
	tests := []videoCreateRequest{
		{Prompt: "", Model: siftq.ModelName},
		{Prompt: "ok", Model: "other-model"},
		{Prompt: "ok", Model: siftq.ModelName, Seconds: "3"},
		{Prompt: "ok", Model: siftq.ModelName, Size: "1920x1080"},
	}
	for index := range tests {
		if _, _, _, err := convertVideoCreateRequest(&tests[index], ""); err == nil {
			t.Fatalf("case %d should fail", index)
		}
	}
}

func TestVideoStatusConversion(t *testing.T) {
	binding := &model.SiftQTaskBinding{TaskID: "video-1", Model: siftq.ModelName, Prompt: "prompt", Seconds: 8, Size: "1280x720", CreatedAt: 100}
	tests := []struct {
		provider string
		want     string
		progress int
	}{
		{"queued", "queued", 0},
		{"running", "in_progress", 50},
		{"succeeded", "completed", 100},
		{"failed", "failed", 100},
		{"cancelled", "failed", 100},
	}
	for _, test := range tests {
		got := videoFromTask(siftq.VideoTask{ID: "video-1", Status: test.provider}, binding)
		if got.Status != test.want || got.Progress != test.progress {
			t.Fatalf("status %s mapped to %s/%d", test.provider, got.Status, got.Progress)
		}
	}
}
