// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package pchain

import (
	"testing"

	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/stretchr/testify/require"

	"github.com/rink-labs/Rink-Tooling-SDK-go/network"
	pchainTxs "github.com/rink-labs/Rink-Tooling-SDK-go/wallet/txs/p-chain"
)

func TestBuildCreateSubnetTxRejectsMismatchedHRP(t *testing.T) {
	mismatched, err := address.Format("P", "avax", make([]byte, 20))
	require.NoError(t, err)

	// The HRP check runs before the wallet is used, so a nil wallet is safe
	// here and keeps this a focused unit test.
	_, err = buildCreateSubnetTx(nil, network.RinkMainnetNetwork(), &pchainTxs.CreateSubnetTxParams{
		ControlKeys: []string{mismatched},
		Threshold:   1,
	})
	require.ErrorIs(t, err, network.ErrHRPMismatch)
}

func TestConvertSubnetAuthKeysRejectsMismatchedHRP(t *testing.T) {
	mismatched, err := address.Format("P", "avax", make([]byte, 20))
	require.NoError(t, err)

	_, err = convertSubnetAuthKeys(network.RinkMainnetNetwork(), []string{mismatched})
	require.ErrorIs(t, err, network.ErrHRPMismatch)
}
