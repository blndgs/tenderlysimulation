package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stackup-wallet/stackup-bundler/pkg/entrypoint"
	"github.com/stackup-wallet/stackup-bundler/pkg/userop"
)

type tenderlySimulationRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Id      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type simulationRequest struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
	Data string `json:"data,omitempty"`
}

type simulationResponse struct {
	ID      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Status bool `json:"status"`
	} `json:"result"`
}

func doSimulateUserop(userop *userop.UserOperation,
	entrypointAddr common.Address,
	rpcURL string) (simulationResponse, error) {

	client := &http.Client{
		Timeout: time.Minute,
	}

	parsed, err := abi.JSON(strings.NewReader(entrypoint.EntrypointABI))
	if err != nil {
		return simulationResponse{}, err
	}

	calldata, err := parsed.Pack("handleOps", []entrypoint.UserOperation{entrypoint.UserOperation(*userop)}, common.HexToAddress("0xa4bfe126d3ad137f972695dddb1780a29065e556"))
	if err != nil {
		return simulationResponse{}, err
	}

	var data = simulationRequest{
		Data: "0x" + hex.EncodeToString(calldata),
		From: userop.Sender.Hex(),
		To:   entrypointAddr.Hex(),
	}

	r := tenderlySimulationRequest{
		Params: []interface {
		}{data, "latest"},
		Method:  "tenderly_simulateTransaction",
		Jsonrpc: "2.0",
		Id:      1,
	}

	var b = new(bytes.Buffer)

	if err := json.NewEncoder(b).Encode(r); err != nil {
		return simulationResponse{}, err
	}

	req, err := http.NewRequest(http.MethodPost, rpcURL, b)
	if err != nil {
		return simulationResponse{}, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {

		err = errors.New("could not simulate transaction. error making http request")
		return simulationResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode > http.StatusOK {
		var errResp struct {
			Error struct {
				ID      string `json:"id"`
				Slug    string `json:"slug"`
				Message string `json:"message"`
			} `json:"error"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return simulationResponse{}, err
		}

		err := errors.New(errResp.Error.Message)

		return simulationResponse{}, err
	}

	var simulatedResponse simulationResponse

	if err := json.NewDecoder(resp.Body).Decode(&simulatedResponse); err != nil {
		return simulationResponse{}, err
	}

	if !simulatedResponse.Result.Status {
		// all failures have an "execution reverted" message
		err = errors.New("could not simulate transaction. execution reverted")

		return simulationResponse{}, err
	}

	return simulatedResponse, nil
}
