package lib

type KeyGenRequest struct {
	Network int `json:"network"`
}

type KeyGenResponse struct {
	Address  string `json:"address"`
	Duration int    `json:"duration"`
}

type SignRequest struct {
	Address  string `json:"address"`
	TxOrigin string `json:"tx_origin"` // encoded base64
}

type SignResponse struct {
	V        int    `json:"v"`
	R        string `json:"r"` // encoded base64
	S        string `json:"s"` // encoded base64
	Duration int    `json:"duration"`
}
