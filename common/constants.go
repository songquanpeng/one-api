package common

import "time"

var StartTime = time.Now().Unix() // unit: second
var Version = "v0.0.0"            // this hard coding will be replaced automatically when building, no need to manually change

var (
	// CtxKeyChannel is the key to store the channel in the context
	CtxKeyChannel          string = "channel_docu"
	CtxKeyRequestModel     string = "request_model"
	CtxKeyRawRequest       string = "raw_request"
	CtxKeyConvertedRequest string = "converted_request"
	CtxKeyOriginModel      string = "origin_model"
)
