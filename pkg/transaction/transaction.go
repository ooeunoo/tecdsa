package transaction

type UnsignedTransaction struct {
	NetworkID       int32       `json:"network_id"`
	UnSignedTxEncodedBase64 string      `json:"unsigned_tx_encoded_base64"`
	Extra           interface{} `json:"extra,omitempty"`
}
