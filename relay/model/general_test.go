package model

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGeneralOpenAIRequestDeepSeekThinkingPassthrough(t *testing.T) {
	Convey("GeneralOpenAIRequest should preserve DeepSeek thinking parameters", t, func() {
		raw := []byte(`{"model":"deepseek-v4-pro","messages":[],"thinking":{"type":"enabled"},"reasoning_effort":"high"}`)
		request := GeneralOpenAIRequest{}

		err := json.Unmarshal(raw, &request)
		So(err, ShouldBeNil)

		encoded, err := json.Marshal(request)
		So(err, ShouldBeNil)

		decoded := map[string]any{}
		err = json.Unmarshal(encoded, &decoded)
		So(err, ShouldBeNil)
		So(decoded["thinking"], ShouldResemble, map[string]any{"type": "enabled"})
		So(decoded["reasoning_effort"], ShouldEqual, "high")
	})
}
