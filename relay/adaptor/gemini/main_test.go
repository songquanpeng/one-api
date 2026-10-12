package gemini

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/songquanpeng/one-api/relay/model"
)

func TestConvertRequestSystemInstruction(t *testing.T) {
	for _, name := range []string{"gemini-1.5-flash", "gemini-1.5-pro-002", "gemini-2.0-flash", "gemini-2.5-flash", "gemini-2.5-pro", "gemini-3.8-flash"} {
		t.Run(name, func(t *testing.T) {
			request := ConvertRequest(model.GeneralOpenAIRequest{
				Model: name,
				Messages: []model.Message{
					{Role: "system", Content: "You are a helpful assistant."},
					{Role: "user", Content: "hi"},
					{Role: "assistant", Content: "hello"},
					{Role: "user", Content: "continue"},
				},
			})
			want := []ChatContent{
				{Role: "user", Parts: []Part{{Text: "hi"}}},
				{Role: "model", Parts: []Part{{Text: "hello"}}},
				{Role: "user", Parts: []Part{{Text: "continue"}}},
			}
			if !reflect.DeepEqual(request.Contents, want) {
				t.Fatalf("conversation = %#v, want %#v", request.Contents, want)
			}
			if request.SystemInstruction == nil || request.SystemInstruction.Role != "" || !reflect.DeepEqual(request.SystemInstruction.Parts, []Part{{Text: "You are a helpful assistant."}}) {
				t.Fatalf("system instruction = %#v", request.SystemInstruction)
			}
			encoded, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			var body map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &body); err != nil {
				t.Fatal(err)
			}
			if _, ok := body["system_instruction"]; !ok {
				t.Fatalf("serialized request lacks system_instruction: %s", encoded)
			}
		})
	}
}

func TestConvertRequestLegacySystemPrompt(t *testing.T) {
	request := ConvertRequest(model.GeneralOpenAIRequest{
		Model: "gemini-pro",
		Messages: []model.Message{
			{Role: "system", Content: "instructions"},
			{Role: "user", Content: "hi"},
		},
	})
	if request.SystemInstruction != nil {
		t.Fatal("legacy model must not receive system_instruction")
	}
	want := []ChatContent{
		{Role: "user", Parts: []Part{{Text: "instructions"}}},
		{Role: "model", Parts: []Part{{Text: "Okay"}}},
		{Role: "user", Parts: []Part{{Text: "hi"}}},
	}
	if !reflect.DeepEqual(request.Contents, want) {
		t.Fatalf("legacy conversation = %#v, want %#v", request.Contents, want)
	}
}

func TestConvertRequestWithoutSystemPrompt(t *testing.T) {
	request := ConvertRequest(model.GeneralOpenAIRequest{
		Model:    "gemini-2.5-flash",
		Messages: []model.Message{{Role: "user", Content: "hi"}},
	})
	if request.SystemInstruction != nil || !reflect.DeepEqual(request.Contents, []ChatContent{{Role: "user", Parts: []Part{{Text: "hi"}}}}) {
		t.Fatalf("unexpected request: %#v", request)
	}
}
