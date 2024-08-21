package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func fundUserWallet(addr common.Address, rpcURL string) error {

	reqbody := fmt.Sprintf(`
{
    "jsonrpc": "2.0",
    "method": "tenderly_setBalance",
    "params": [
      [
        "%s"
        ],
      "0xDE0B6B3A7640000"
      ]
}
	`, addr.Hex())

	req, err := http.NewRequest(http.MethodPost, rpcURL, strings.NewReader(reqbody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	client := &http.Client{
		Timeout: time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}
