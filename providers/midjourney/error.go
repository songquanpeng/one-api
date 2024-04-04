// Author: Calcium-Ion
// GitHub: https://github.com/Calcium-Ion/new-api
// Path: service/error.go
package midjourney

func MidjourneyErrorWithStatusCodeWrapper(code int, desc string, statusCode int) *MidjourneyResponseWithStatusCode {
	return &MidjourneyResponseWithStatusCode{
		StatusCode: statusCode,
		Response:   *MidjourneyErrorWrapper(code, desc),
	}
}

func MidjourneyErrorWrapper(code int, desc string) *MidjourneyResponse {
	return &MidjourneyResponse{
		Code:        code,
		Description: desc,
	}
}
