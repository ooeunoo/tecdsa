package response

import "net/http"

// Error codes
const (
	ErrCodeBadRequest          = "BAD_REQUEST"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeInternalServerError = "INTERNAL_SERVER_ERROR"
	ErrCodeKeyGeneration       = "KEY_GENERATION_ERROR"
	ErrCodeSigning             = "SIGNING_ERROR"

	// Add more error codes as needed
)

// Error code to HTTP status code mapping
var ErrorCodeToStatusCode = map[string]int{
	ErrCodeBadRequest:          http.StatusBadRequest,
	ErrCodeUnauthorized:        http.StatusUnauthorized,
	ErrCodeForbidden:           http.StatusForbidden,
	ErrCodeNotFound:            http.StatusNotFound,
	ErrCodeInternalServerError: http.StatusInternalServerError,
	ErrCodeKeyGeneration:       http.StatusInternalServerError,
	ErrCodeSigning:             http.StatusInternalServerError,
}

// Error code to message mapping
var ErrorCodeToMessage = map[string]string{
	ErrCodeBadRequest:          "잘못된 요청입니다",
	ErrCodeUnauthorized:        "인증되지 않은 요청입니다",
	ErrCodeForbidden:           "접근이 금지되었습니다",
	ErrCodeNotFound:            "요청한 리소스를 찾을 수 없습니다",
	ErrCodeInternalServerError: "내부 서버 오류가 발생했습니다",
	ErrCodeKeyGeneration:       "키 생성 중 알 수 없는 오류가 발생했습니다",
	ErrCodeSigning:             "서명 중 오류가 발생했습니다",
}
