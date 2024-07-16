package lib

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

func GetUnspentTxs(address string, network string) ([]UTXO, error) {
	var url string
	switch network {
	case "mainnet":
		url = fmt.Sprintf("https://mempool.space/api/address/%s/utxo", address)
	case "testnet":
		url = fmt.Sprintf("https://mempool.space/testnet/api/address/%s/utxo", address)
	case "regtest":
		return getRegtestUnspentTxs(address)
	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get UTXOs: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var utxos []UTXO
	err = json.Unmarshal(body, &utxos)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal UTXOs: %v", err)
	}

	return utxos, nil
}

func getRegtestUnspentTxs(address string) ([]UTXO, error) {
	rpcURL := "http://localhost:18443"
	rpcUser := "myuser"
	rpcPassword := "SomeDecentp4ssw0rd"

	payload := []byte(fmt.Sprintf(`{
        "jsonrpc": "1.0",
        "id": "curltest",
        "method": "listunspent",
        "params": [1, 9999999, ["%s"], true, {
            "minimumAmount": 0,
            "maximumCount": 1000
        }]
    }`, address))

	req, err := http.NewRequest("POST", rpcURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth(rpcUser, rpcPassword)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var rpcResponse struct {
		Result []struct {
			TxID          string  `json:"txid"`
			Vout          uint32  `json:"vout"`
			Address       string  `json:"address"`
			Amount        float64 `json:"amount"`
			Confirmations int64   `json:"confirmations"`
			ScriptPubKey  string  `json:"scriptPubKey"`
			Spendable     bool    `json:"spendable"`
			Solvable      bool    `json:"solvable"`
			Safe          bool    `json:"safe"`
		} `json:"result"`
	}

	err = json.Unmarshal(body, &rpcResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	var utxos []UTXO
	for _, unspent := range rpcResponse.Result {
		scriptPubKey, err := hex.DecodeString(unspent.ScriptPubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode scriptPubKey: %v", err)
		}
		utxos = append(utxos, UTXO{
			TxID:         unspent.TxID,
			Vout:         unspent.Vout,
			Address:      unspent.Address,
			Value:        int64(unspent.Amount * 1e8), // BTC to satoshis
			ScriptPubKey: scriptPubKey,
			Status: UTXOStatus{
				Confirmed:   unspent.Confirmations > 0,
				BlockHeight: 0,  // This information is not provided by listunspent
				BlockHash:   "", // This information is not provided by listunspent
				BlockTime:   0,  // This information is not provided by listunspent
			},
		})
	}

	return utxos, nil
}

func GetAddressType(address string, params *chaincfg.Params) AddressType {
	addr, err := btcutil.DecodeAddress(address, params)
	if err != nil {
		return Unknown
	}

	switch addr.(type) {
	case *btcutil.AddressPubKeyHash:
		return P2PKH
	case *btcutil.AddressScriptHash:
		return P2SH
	case *btcutil.AddressWitnessPubKeyHash:
		return P2WPKH
	case *btcutil.AddressWitnessScriptHash:
		return P2WSH
	default:
		// Additional check for Bech32 addresses
		if strings.HasPrefix(address, "bc1") || strings.HasPrefix(address, "tb1") {
			if len(address) == 42 {
				return P2WPKH
			} else if len(address) == 62 {
				return P2WSH
			}
		}
		return Unknown
	}
}

func main() {
	// Example usage
	testnet := &chaincfg.TestNet3Params
	mainnet := &chaincfg.MainNetParams

	addresses := []string{
		"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2",                             // P2PKH
		"3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy",                             // P2SH
		"bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4",                     // P2WPKH
		"bc1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3qccfmv3", // P2WSH
		"tb1qw508d6qejxtdg4y5r3zarvary0c5xw7kxpjzsx",                     // P2WPKH (testnet)
	}

	for _, addr := range addresses {
		var addrType AddressType
		if strings.HasPrefix(addr, "bc1") || strings.HasPrefix(addr, "1") || strings.HasPrefix(addr, "3") {
			addrType = GetAddressType(addr, mainnet)
		} else {
			addrType = GetAddressType(addr, testnet)
		}
		println(addr, ":", string(addrType))
	}
}
