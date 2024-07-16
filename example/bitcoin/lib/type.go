package lib

type KeyGenRequest struct {
	Network int `json:"network"`
}

type KeyGenResponse struct {
	Address   string `json:"address"`
	PublicKey string `json:"public_key"` // encoded base64
	Duration  int    `json:"duration"`
}

type SignRequest struct {
	Address   string `json:"address"`
	SecretKey string `json:"secret_key"` // encoded base64
	TxOrigin  string `json:"tx_origin"`  // encoded base64
}

type SignResponse struct {
	V        int    `json:"v"`
	R        string `json:"r"` // encoded base64
	S        string `json:"s"` // encoded base64
	Duration int    `json:"duration"`
}

type UTXOResponse struct {
	Chain         string  `json:"chain"`
	Address       string  `json:"address"`
	TxHash        string  `json:"txHash"`
	Index         int     `json:"index"`
	Value         float64 `json:"value"`
	ValueAsString string  `json:"valueAsString"`
}

type UTXOStatus struct {
	Confirmed   bool   `json:"confirmed"`
	BlockHeight int64  `json:"block_height"`
	BlockHash   string `json:"block_hash"`
	BlockTime   int64  `json:"block_time"`
}

type UTXO struct {
	TxID         string     `json:"txid"`
	Vout         uint32     `json:"vout"`
	Status       UTXOStatus `json:"status"`
	Value        int64      `json:"value"`
	ScriptPubKey []byte     `json:"scriptPubKey"`
	Address      string     `json:"address"` // Add this line

}

type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type RPCResponse struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
	ID     int         `json:"id"`
}
