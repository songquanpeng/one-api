package openai

func ErrorWrapper(err error, code string, statusCode int) *ErrorWithStatusCode {
	Error := Error{
		Message: err.Error(),
		Type:    "one_api_error",
		Code:    code,
	}
	return &ErrorWithStatusCode{
		Error:      Error,
		StatusCode: statusCode,
	}
}
