package controller

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/songquanpeng/one-api/relay/adaptor/openai"
)

func TestBuildMinimaxSpeechRequest(t *testing.T) {
	body, format, err := buildMinimaxSpeechRequest(openai.TextToSpeechRequest{
		Model:          "speech-2.8-hd",
		Input:          "hello",
		Voice:          "female-shaonv",
		ResponseFormat: "wav",
	})
	if err != nil {
		t.Fatal(err)
	}
	if format != "wav" {
		t.Fatalf("format = %q", format)
	}
	var request minimaxSpeechRequest
	if err := json.Unmarshal(body, &request); err != nil {
		t.Fatal(err)
	}
	if request.Model != "speech-2.8-hd" || request.Text != "hello" || request.VoiceSetting.VoiceID != "female-shaonv" || request.VoiceSetting.Speed != 1 || request.AudioSetting.Format != "wav" || request.OutputFormat != "hex" {
		t.Fatalf("unexpected request: %+v", request)
	}
}

func TestDecodeMinimaxSpeechResponse(t *testing.T) {
	hexAudio := hex.EncodeToString([]byte("audio"))
	decoded, err := decodeMinimaxSpeechResponse([]byte(`{"data":{"audio":"` + hexAudio + `"},"base_resp":{"status_code":0}}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(decoded) != "audio" {
		t.Fatalf("decoded = %q", decoded)
	}

	base64Audio := base64.StdEncoding.EncodeToString([]byte("audio"))
	decoded, err = decodeMinimaxSpeechResponse([]byte(`{"data":{"audio":"` + base64Audio + `"},"base_resp":{"status_code":0}}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(decoded) != "audio" {
		t.Fatalf("decoded = %q", decoded)
	}
}

func TestMinimaxSpeechURL(t *testing.T) {
	if got := minimaxSpeechURL("https://api.minimax.chat"); got != "https://api.minimax.io/v1/t2a_v2" {
		t.Fatalf("url = %q", got)
	}
	if got := minimaxSpeechURL("https://api.minimax.io/v1/"); got != "https://api.minimax.io/v1/t2a_v2" {
		t.Fatalf("url = %q", got)
	}
}
