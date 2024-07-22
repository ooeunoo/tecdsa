package main

import (
	"context"
	"ethereum_example/lib"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("이더리움 테스트", func() {
	var (
		client     *ethclient.Client
		infuraKey  string
		privateKey string
	)

	BeforeEach(func() {
		infuraKey, privateKey = loadENV()
		var err error
		client, err = connectToEthereum(infuraKey)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("환경 설정", func() {
		It("한경 변수 로드", func() {
			Expect(infuraKey).NotTo(BeEmpty())
			Expect(privateKey).NotTo(BeEmpty())
		})

		It("이더리움 네트워크 로드", func() {
			Expect(client).NotTo(BeNil())
		})
	})

	Describe("키 생성", func() {
		It("새로운 키 생성 시 주소를 리턴", func() {
			keyGenResp, err := handleKeyGeneration()
			Expect(err).NotTo(HaveOccurred())
			Expect(keyGenResp.Address).NotTo(BeEmpty())
		})
	})

	Describe("테스트 이더 주입", func() {
		It("-", func() {
			keyGenResp, _ := handleKeyGeneration()
			err := injectTestEther(client, privateKey, keyGenResp.Address)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("트랜잭션 전송", func() {
		It("서명 값(V, R, S)와 원본 Tx를 결합하여 네트워크 전송", func() {
			chainID, _ := client.NetworkID(context.Background())
			keyGenResp, _ := handleKeyGeneration()
			signer := types.NewEIP155Signer(chainID)
			amount := big.NewInt(33333333333333333) // 0.033333333333333333 Ether
			tx, rlpEncodedTx := lib.GenerateRlpEncodedTx(
				*client,
				signer,
				common.HexToAddress(keyGenResp.Address),
				common.HexToAddress("0xFDcBF476B286796706e273F86aC51163DA737FA8"),
				amount,
			)
			signResp, _ := signTransaction(keyGenResp.Address, rlpEncodedTx)
			fmt.Printf("signResp: V: 0x%x, R: 0x%s, S: 0x%s\n", signResp.V, signResp.R, signResp.S)
			signedTx, _, _ := lib.CombineETHUnsignedTxWithSignature(tx, chainID, *signResp)
			_, err := lib.SendSignedTransaction(client, signedTx, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(signedTx).NotTo(BeNil())

		})
	})
})
