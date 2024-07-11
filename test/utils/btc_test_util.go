package test_util

// import (
// 	"bytes"
// 	"encoding/base64"
// 	"encoding/hex"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"math/big"
// 	"net/http"

// 	"github.com/btcsuite/btcd/btcec"

// 	"github.com/btcsuite/btcd/chaincfg"
// 	"github.com/btcsuite/btcd/chaincfg/chainhash"
// 	"github.com/btcsuite/btcd/txscript"
// 	"github.com/btcsuite/btcd/wire"
// 	"github.com/btcsuite/btcutil"
// )

// type UnspentUTXO struct {
// 	TxID   string
// 	Vout   uint32
// 	Amount int64 // in satoshis
// 	Script string
// }

// // GenerateTxOrigin generates a Bitcoin transaction for testing purposes.
// func GenerateBTCTxOrigin(toAddress string) (*wire.MsgTx, []byte, error) {
// 	// 테스트용 값 설정
// 	amount, err := btcutil.NewAmount(0.0001) // 0.0001 BTC
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("failed to create amount: %v", err)
// 	}
// 	toAddr, err := btcutil.DecodeAddress(toAddress, &chaincfg.MainNetParams)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("failed to decode address: %v", err)
// 	}

// 	// 트랜잭션 입력 설정
// 	prevTxHash, err := chainhash.NewHashFromStr("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("failed to create prevTxHash: %v", err)
// 	}
// 	prevOut := wire.NewOutPoint(prevTxHash, 0)
// 	txIn := wire.NewTxIn(prevOut, nil, nil)

// 	// 트랜잭션 출력 설정
// 	pkScript, err := txscript.PayToAddrScript(toAddr)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("failed to create pkScript: %v", err)
// 	}
// 	txOut := wire.NewTxOut(int64(amount), pkScript)

// 	// 트랜잭션 생성
// 	tx := wire.NewMsgTx(wire.TxVersion)
// 	tx.AddTxIn(txIn)
// 	tx.AddTxOut(txOut)

// 	// 트랜잭션 직렬화
// 	var buf bytes.Buffer
// 	if err := tx.Serialize(&buf); err != nil {
// 		return nil, nil, fmt.Errorf("failed to serialize transaction: %v", err)
// 	}
// 	txBytes := buf.Bytes()

// 	return tx, txBytes, nil
// }
// func CombineBTCUnsignedTxWithSignature(tx *wire.MsgTx, response []byte, publicKey []byte) (*wire.MsgTx, error) {
// 	var signResponse SignResponse
// 	err := json.Unmarshal(response, &signResponse)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal sign response: %v", err)
// 	}

// 	if !signResponse.Success {
// 		return nil, fmt.Errorf("signing was not successful")
// 	}

// 	// R과 S를 big.Int로 변환
// 	rBytes, err := base64.StdEncoding.DecodeString(signResponse.R)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to decode 'r': %v", err)
// 	}
// 	r := new(big.Int).SetBytes(rBytes)

// 	sBytes, err := base64.StdEncoding.DecodeString(signResponse.S)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to decode 's': %v", err)
// 	}
// 	s := new(big.Int).SetBytes(sBytes)

// 	// 서명 생성
// 	signature := &btcec.Signature{
// 		R: r,
// 		S: s,
// 	}

// 	// DER 인코딩된 서명 생성
// 	signatureDER := signature.Serialize()

// 	// 서명 스크립트 생성 (P2PKH 트랜잭션 가정)
// 	builder := txscript.NewScriptBuilder()
// 	builder.AddData(signatureDER)
// 	builder.AddData(publicKey)
// 	signatureScript, err := builder.Script()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to build signature script: %v", err)
// 	}

// 	// 서명을 트랜잭션에 추가
// 	tx.TxIn[0].SignatureScript = signatureScript

// 	// 서명 검증
// 	// flags := txscript.StandardVerifyFlags
// 	// vm, err := txscript.NewEngine(prevOutputScript, tx, 0, flags, nil, nil, amount)
// 	// if err != nil {
// 	// 	return nil, fmt.Errorf("failed to create script engine: %v", err)
// 	// }
// 	// if err := vm.Execute(); err != nil {
// 	// 	return nil, fmt.Errorf("script execution failed: %v", err)
// 	// }

// 	return tx, nil
// }

// func InjectTestBTC(utxo UnspentUTXO, privateKeyWIF string, toAddress string, amount int64) (string, error) {
// 	// 1. 개인키 디코딩
// 	wif, err := btcutil.DecodeWIF(privateKeyWIF)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to decode WIF: %v", err)
// 	}
// 	privateKey := wif.PrivKey

// 	// 2. 목적지 주소 디코딩
// 	destAddr, err := btcutil.DecodeAddress(toAddress, &chaincfg.TestNet3Params)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to decode destination address: %v", err)
// 	}

// 	// 3. 트랜잭션 생성
// 	tx := wire.NewMsgTx(wire.TxVersion)

// 	// 4. 입력 추가
// 	prevTxHash, err := chainhash.NewHashFromStr(utxo.TxID)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to parse transaction hash: %v", err)
// 	}
// 	outpoint := wire.NewOutPoint(prevTxHash, utxo.Vout)
// 	txIn := wire.NewTxIn(outpoint, nil, nil)
// 	tx.AddTxIn(txIn)

// 	// 5. 출력 추가 (목적지 주소로)
// 	pkScript, err := txscript.PayToAddrScript(destAddr)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to create pkScript: %v", err)
// 	}
// 	txOut := wire.NewTxOut(amount, pkScript)
// 	tx.AddTxOut(txOut)

// 	// 6. 거스름돈 출력 추가 (필요한 경우)
// 	changeAmount := utxo.Amount - amount - 1000 // 1000 satoshis for fee
// 	if changeAmount > 0 {
// 		changeAddr, err := btcutil.NewAddressPubKey(privateKey.PubKey().SerializeCompressed(), &chaincfg.TestNet3Params)
// 		if err != nil {
// 			return "", fmt.Errorf("failed to create change address: %v", err)
// 		}
// 		changePkScript, err := txscript.PayToAddrScript(changeAddr)
// 		if err != nil {
// 			return "", fmt.Errorf("failed to create change pkScript: %v", err)
// 		}
// 		changeOutput := wire.NewTxOut(changeAmount, changePkScript)
// 		tx.AddTxOut(changeOutput)
// 	}

// 	// 7. 서명
// 	inputScript, err := hex.DecodeString(utxo.Script)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to decode input script: %v", err)
// 	}
// 	sigScript, err := txscript.SignTxOutput(&chaincfg.TestNet3Params, tx, 0, inputScript, txscript.SigHashAll, txscript.KeyClosure(func(a btcutil.Address) (*btcec.PrivateKey, bool, error) {
// 		return privateKey, true, nil
// 	}), nil, nil)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to sign transaction: %v", err)
// 	}
// 	tx.TxIn[0].SignatureScript = sigScript

// 	// 8. 트랜잭션 직렬화
// 	var signedTx bytes.Buffer
// 	if err := tx.Serialize(&signedTx); err != nil {
// 		return "", fmt.Errorf("failed to serialize transaction: %v", err)
// 	}

// 	return hex.EncodeToString(signedTx.Bytes()), nil
// }

// // BroadcastTestBTCTransaction broadcasts a signed transaction to the Bitcoin testnet
// func BroadcastTestBTCTransaction(signedTxHex string) (string, error) {
// 	// Bitcoin testnet API URL (예: BlockCypher API)
// 	apiURL := "https://api.blockcypher.com/v1/btc/test3/txs/push"

// 	// 요청 본문 생성
// 	requestBody, err := json.Marshal(map[string]string{
// 		"tx": signedTxHex,
// 	})
// 	if err != nil {
// 		return "", fmt.Errorf("failed to marshal request body: %v", err)
// 	}

// 	// HTTP POST 요청 생성
// 	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(requestBody))
// 	if err != nil {
// 		return "", fmt.Errorf("failed to send transaction: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	// 응답 읽기
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to read response: %v", err)
// 	}

// 	// 응답 파싱
// 	var response struct {
// 		Tx struct {
// 			Hash string `json:"hash"`
// 		} `json:"tx"`
// 	}
// 	if err := json.Unmarshal(body, &response); err != nil {
// 		return "", fmt.Errorf("failed to parse response: %v", err)
// 	}

// 	// 트랜잭션 해시 반환
// 	return response.Tx.Hash, nil
// }

// // PrintSignedTxAsJSON prints the signed Bitcoin transaction as JSON.
// func PrintBTCSignedTxAsJSON(signedTx *wire.MsgTx) {
// 	var buf bytes.Buffer
// 	if err := signedTx.Serialize(&buf); err != nil {
// 		log.Fatalf("Failed to serialize transaction: %v", err)
// 	}
// 	txHex := hex.EncodeToString(buf.Bytes())

// 	txJSON := struct {
// 		TxHex string `json:"txHex"`
// 	}{
// 		TxHex: txHex,
// 	}

// 	jsonData, err := json.MarshalIndent(txJSON, "", "  ")
// 	if err != nil {
// 		log.Fatalf("Failed to marshal transaction to JSON: %v", err)
// 	}
// 	fmt.Printf("Signed Transaction as JSON:\n%s\n", string(jsonData))
// }
