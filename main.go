package main

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	"github.com/stackup-wallet/stackup-bundler/pkg/userop"
	"golang.org/x/crypto/sha3"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	chainID, ok := big.NewInt(0).SetString(os.Getenv("CHAIN_ID"), 10)
	if !ok {
		log.Fatal("chain id is invalid..")
	}

	addr, key := generateWallet()
	defer zeroKey(key)

	userOperation := &userop.UserOperation{
		Sender:               addr,
		Nonce:                big.NewInt(0),
		CallData:             []byte(""),
		InitCode:             getInitCode(addr),
		CallGasLimit:         big.NewInt(1000),
		PreVerificationGas:   big.NewInt(1000),
		VerificationGasLimit: big.NewInt(1000),
		MaxFeePerGas:         big.NewInt(1000),
		MaxPriorityFeePerGas: big.NewInt(1000),
	}

	userOperation, err := sign(chainID, key, userOperation)
	if err != nil {
		log.Fatalf("could not sign userop...%v", err)
	}

	rpcURL := os.Getenv("ETH_RPC_URL")
	if rpcURL == "" {
		panic("please provide a valid RPC url")
	}

	if err := fundUserWallet(addr, rpcURL); err != nil {
		log.Fatalf("could not fund wallet with ETH... %v", err)
	}

	resp, err := doSimulateUserop(userOperation, common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789"), rpcURL)
	if err != nil {
		log.Fatal(err.Error())
	}

	_ = resp
}

func sign(chainID *big.Int, privateKey *ecdsa.PrivateKey, userOp *userop.UserOperation) (*userop.UserOperation, error) {

	entryPointAddr := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")

	signature, err := getSignature(chainID, privateKey, entryPointAddr, userOp)
	if err != nil {
		return &userop.UserOperation{}, err
	}
	userOp.Signature = signature
	return userOp, nil
}

func getSignature(chainID *big.Int, privateKey *ecdsa.PrivateKey, entryPointAddr common.Address,
	userOp *userop.UserOperation) ([]byte, error) {
	userOpHashObj := userOp.GetUserOpHash(entryPointAddr, chainID)

	userOpHash := userOpHashObj.Bytes()
	prefixedHash := crypto.Keccak256Hash(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(userOpHash), userOpHash)),
	)

	signature, err := crypto.Sign(prefixedHash.Bytes(), privateKey)
	if err != nil {
		return nil, err
	}

	signature[64] += 27
	return signature, nil
}

func getInitCode(addr common.Address) []byte {

	s := fmt.Sprintf(`0x42E60c23aCe33c23e0945a07f6e2c1E53843a1d55fbfb9cf000000000000000000000000%s0000000000000000000000000000000000000000000000000000000000000000`,
		strings.TrimPrefix(addr.Hex(), "0x"))

	return []byte(s)
}

func generateWallet() (common.Address, *ecdsa.PrivateKey) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	hash := sha3.NewLegacyKeccak256()
	_, err = hash.Write(publicKeyBytes[1:])
	if err != nil {
		log.Fatalf("could not write hash... %v", err)
	}

	return common.HexToAddress(address), privateKey
}

func zeroKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	for i := range b {
		b[i] = 0
	}
}
