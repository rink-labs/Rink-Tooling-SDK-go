// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rink-labs/Rink-Tooling-SDK-go/network"
	"github.com/rink-labs/Rink-Tooling-SDK-go/utils"
	"github.com/rink-labs/Rink-Tooling-SDK-go/wallet/local"
	"github.com/rink-labs/Rink-Tooling-SDK-go/wallet/types"

	avagoTxs "github.com/ava-labs/avalanchego/vms/platformvm/txs"
	pchainTxs "github.com/rink-labs/Rink-Tooling-SDK-go/wallet/txs/p-chain"
)

func CreateSubnet() error {
	ctx, cancel := utils.GetTimedContext(120 * time.Second)
	defer cancel()

	rpcEndpoint := os.Getenv("RINK_RPC_ENDPOINT")
	if rpcEndpoint == "" {
		return fmt.Errorf("RINK_RPC_ENDPOINT must be set")
	}
	network := network.RinkMainnetNetwork().WithEndpoint(rpcEndpoint)

	localWallet, err := local.NewLocalWallet()
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	existingAccount, err := localWallet.ImportAccount("EXISTING_KEY_PATH")
	if err != nil {
		return fmt.Errorf("failed to ImportAccount: %w", err)
	}

	createSubnetParams := &pchainTxs.CreateSubnetTxParams{
		ControlKeys: []string{"P-rink1..."},
		Threshold:   1,
	}
	buildTxParams := types.BuildTxParams{
		Account:      *existingAccount,
		Network:      network,
		BuildTxInput: createSubnetParams,
	}
	buildTxResult, err := localWallet.BuildTx(ctx, buildTxParams)
	if err != nil {
		return fmt.Errorf("failed to BuildTx: %w", err)
	}

	signTxParams := types.SignTxParams{
		Account:       *existingAccount,
		Network:       network,
		BuildTxResult: &buildTxResult,
	}
	signTxResult, err := localWallet.SignTx(ctx, signTxParams)
	if err != nil {
		return fmt.Errorf("failed to signTx: %w", err)
	}

	sendTxParams := types.SendTxParams{
		Account:      *existingAccount,
		Network:      network,
		SignTxResult: &signTxResult,
	}
	sendTxResult, err := localWallet.SendTx(ctx, sendTxParams)
	if err != nil {
		return fmt.Errorf("failed to sendTx: %w", err)
	}
	if tx := sendTxResult.GetTx(); tx != nil {
		if pChainTx, ok := tx.(*avagoTxs.Tx); ok {
			fmt.Printf("sendTxResult %s \n", pChainTx.ID())
		} else {
			fmt.Printf("sendTxResult %s transaction \n", sendTxResult.GetChainType())
		}
	}
	return nil
}

func main() {
	if err := CreateSubnet(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
