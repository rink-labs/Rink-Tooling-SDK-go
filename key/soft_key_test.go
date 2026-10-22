// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package key_test

import (
	"testing"

	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/stretchr/testify/require"

	"github.com/rink-labs/Rink-Tooling-SDK-go/key"
	"github.com/rink-labs/Rink-Tooling-SDK-go/network"
)

func TestSoftKeyRinkAddresses(t *testing.T) {
	softKey, err := key.NewSoft()
	require.NoError(t, err)

	net := network.RinkMainnetNetwork()
	pAddrs, err := softKey.GetNetworkChainAddress(net, "P")
	require.NoError(t, err)
	require.Len(t, pAddrs, 1)

	chainAlias, hrp, addrBytes, err := address.Parse(pAddrs[0])
	require.NoError(t, err)
	require.Equal(t, "P", chainAlias)
	require.Equal(t, network.RinkMainnetHRP, hrp)
	require.Equal(t, softKey.Addresses()[0].Bytes(), addrBytes)
	require.NoError(t, net.ValidateAddress(pAddrs[0]))

	xAddrs, err := softKey.GetNetworkChainAddress(net, "X")
	require.NoError(t, err)
	require.Len(t, xAddrs, 1)
	require.NoError(t, net.ValidateAddress(xAddrs[0]))

	mismatched, err := address.Format("P", constants.MainnetHRP, addrBytes)
	require.NoError(t, err)
	require.ErrorIs(t, net.ValidateAddress(mismatched), network.ErrHRPMismatch)
}

func TestSoftKeyExplicitNetworkHRPs(t *testing.T) {
	softKey, err := key.NewSoft()
	require.NoError(t, err)

	tests := []struct {
		name string
		net  network.Network
		hrp  string
	}{
		{name: "chennai", net: network.ChennaiTestnetNetwork(), hrp: network.ChennaiTestnetHRP},
		{name: "local", net: network.RinkLocalNetwork(), hrp: network.RinkLocalHRP},
		{name: "fuji", net: network.FujiNetwork(), hrp: constants.FujiHRP},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			addrs, err := softKey.GetNetworkChainAddress(test.net, "P")
			require.NoError(t, err)
			require.Len(t, addrs, 1)

			_, hrp, _, err := address.Parse(addrs[0])
			require.NoError(t, err)
			require.Equal(t, test.hrp, hrp)
			require.NoError(t, test.net.ValidateAddress(addrs[0]))
		})
	}
}

func TestSoftKeyAddressHRPFromNetworkObject(t *testing.T) {
	softKey, err := key.NewSoft()
	require.NoError(t, err)

	// Callers can override the HRP for a custom network.
	net := network.NewNetwork(network.Devnet, 4242, "http://127.0.0.1:9650").WithHRP("custom")
	addrs, err := softKey.GetNetworkChainAddress(net, "P")
	require.NoError(t, err)
	require.Len(t, addrs, 1)

	_, hrp, _, err := address.Parse(addrs[0])
	require.NoError(t, err)
	require.Equal(t, "custom", hrp)
	require.NoError(t, net.ValidateAddress(addrs[0]))
}

func TestSoftKeyRejectsChainWithoutHRP(t *testing.T) {
	softKey, err := key.NewSoft()
	require.NoError(t, err)

	net := network.Network{Kind: network.Devnet, ID: 4242}
	_, err = softKey.GetNetworkChainAddress(net, "P")
	require.ErrorIs(t, err, network.ErrUnknownNetworkID)
}

func TestSoftKeyRejectsUnknownChainAlias(t *testing.T) {
	softKey, err := key.NewSoft()
	require.NoError(t, err)

	_, err = softKey.GetNetworkChainAddress(network.RinkMainnetNetwork(), "C")
	require.Error(t, err)
}

func TestSoftKeyDirectAddressFormatting(t *testing.T) {
	softKey, err := key.NewSoft()
	require.NoError(t, err)

	pAddr, err := softKey.P(network.RinkMainnetHRP)
	require.NoError(t, err)
	require.NoError(t, network.RinkMainnetNetwork().ValidateAddress(pAddr))

	xAddr, err := softKey.X(network.RinkMainnetHRP)
	require.NoError(t, err)
	require.NoError(t, network.RinkMainnetNetwork().ValidateAddress(xAddr))
}
