package response

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse struct {
	StatusCode int         `json:"-"` // HTTP 상태 코드 (JSON에는 포함되지 않음)
	Data       interface{} `json:"data"`
}

type ErrorResponse struct {
	StatusCode int    `json:"status_code"`
	ErrorCode  string `json:"error_code"`
	Message    string `json:"message"`
}

func (e *ErrorResponse) Error() string {
	return e.Message
}

func NewSuccessResponse(statusCode int, data interface{}) *SuccessResponse {
	return &SuccessResponse{
		StatusCode: statusCode,
		Data:       data,
	}
}

func NewErrorResponse(errorCode string, customMessage ...string) *ErrorResponse {
	statusCode, ok := ErrorCodeToStatusCode[errorCode]
	if !ok {
		statusCode = http.StatusInternalServerError
	}

	message := ErrorCodeToMessage[errorCode]
	if len(customMessage) > 0 {
		message = customMessage[0]
	}

	return &ErrorResponse{
		StatusCode: statusCode,
		ErrorCode:  errorCode,
		Message:    message,
	}
}

func SendResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")

	var statusCode int
	switch r := response.(type) {
	case *SuccessResponse:
		statusCode = r.StatusCode
	case *ErrorResponse:
		statusCode = r.StatusCode
	default:
		statusCode = http.StatusOK
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
