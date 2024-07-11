package test_util

type SignResponse struct {
	Success bool   `json:"success"`
	V       uint64 `json:"v"`
	R       []byte `json:"r"`
	S       []byte `json:"s"`
}
