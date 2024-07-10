package test_util

type SignResponse struct {
	Success bool   `json:"success"`
	V       string `json:"v"`
	R       string `json:"r"`
	S       string `json:"s"`
}
