package openai

import "github.com/songquanpeng/one-api/relay/model"

func ErrorWrapper(err error, code string, statusCode int) *model.ErrorWithStatusCode {
	Error := model.Error{
		Message: err.Error(),
		Type:    "one_api_error",
		Code:    code,
	}
	return &model.ErrorWithStatusCode{
		Error:      Error,
		StatusCode: statusCode,
	}
}
