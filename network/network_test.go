// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package network_test

import (
	"errors"
	"testing"

	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
	"github.com/stretchr/testify/require"

	"github.com/rink-labs/Rink-Tooling-SDK-go/network"
)

func TestRinkNetworkDefinitions(t *testing.T) {
	tests := []struct {
		name       string
		net        network.Network
		kind       network.NetworkKind
		id         uint32
		hrp        string
		evmChainID uint64
		endpoint   string
	}{
		{
			name:       "mainnet",
			net:        network.RinkMainnetNetwork(),
			kind:       network.RinkMainnet,
			id:         network.RinkMainnetID,
			hrp:        network.RinkMainnetHRP,
			evmChainID: network.RinkMainnetEVMChainID,
			endpoint:   "",
		},
		{
			name:       "chennai testnet",
			net:        network.ChennaiTestnetNetwork(),
			kind:       network.RinkTestnet,
			id:         network.ChennaiTestnetID,
			hrp:        network.ChennaiTestnetHRP,
			evmChainID: network.ChennaiTestnetEVMChainID,
			endpoint:   "",
		},
		{
			name:       "local",
			net:        network.RinkLocalNetwork(),
			kind:       network.RinkLocal,
			id:         network.RinkLocalID,
			hrp:        network.RinkLocalHRP,
			evmChainID: network.RinkLocalEVMChainID,
			endpoint:   network.RinkLocalAPIEndpoint,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.kind, test.net.Kind)
			require.Equal(t, test.id, test.net.ID)
			require.Equal(t, test.hrp, test.net.HRP)
			require.Equal(t, test.evmChainID, test.net.EVMChainID)
			require.Equal(t, test.endpoint, test.net.Endpoint)
			require.NoError(t, test.net.Validate())
		})
	}
}

func TestNetworkFromNetworkID(t *testing.T) {
	tests := []struct {
		name string
		id   uint32
		kind network.NetworkKind
		hrp  string
	}{
		{name: "rink mainnet", id: network.RinkMainnetID, kind: network.RinkMainnet, hrp: network.RinkMainnetHRP},
		{name: "chennai", id: network.ChennaiTestnetID, kind: network.RinkTestnet, hrp: network.ChennaiTestnetHRP},
		{name: "local", id: network.RinkLocalID, kind: network.RinkLocal, hrp: network.RinkLocalHRP},
		{name: "granite", id: network.GraniteNetworkID, kind: network.Granite},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			net := network.NetworkFromNetworkID(test.id)
			require.Equal(t, test.kind, net.Kind)
			require.Equal(t, test.id, net.ID)
			if test.hrp != "" {
				require.Equal(t, test.hrp, net.HRP)
			}
		})
	}

	// 64293 was proposed as a RINK network ID but does not appear in the RINK
	// genesis or network configuration. It must not silently select a network.
	unknown := network.NetworkFromNetworkID(64293)
	require.Equal(t, network.Undefined, unknown.Kind)
	require.ErrorIs(t, unknown.Validate(), network.ErrUndefinedNetwork)
}

func TestChennaiTakesPrecedenceOverUpstreamFujiID(t *testing.T) {
	// Upstream AvalancheGo still defines FujiID as 5, which RINK reassigns to
	// Chennai. NetworkFromNetworkID must use the RINK mapping.
	net := network.NetworkFromNetworkID(constants.FujiID)
	require.Equal(t, network.RinkTestnet, net.Kind)
	require.Equal(t, network.ChennaiTestnetHRP, net.HRP)
}

func TestFujiMustBeExplicit(t *testing.T) {
	net := network.FujiNetwork()
	require.Equal(t, network.Fuji, net.Kind)
	require.Equal(t, constants.FujiID, net.ID)
	require.Equal(t, constants.FujiHRP, net.HRP)
	require.Equal(t, network.FujiEVMChainID, net.EVMChainID)
	require.Equal(t, network.FujiAPIEndpoint, net.Endpoint)
	require.NoError(t, net.Validate())

	// It is not reachable by network ID; use FujiNetwork explicitly.
	require.NotEqual(t, network.Fuji, network.NetworkFromNetworkID(constants.FujiID).Kind)
}

func TestAddressHRPKindPrecedesNetworkID(t *testing.T) {
	// Fuji and Chennai share network ID 5. An explicit kind must decide the
	// HRP, otherwise Fuji addresses would be encoded as Chennai addresses.
	fuji := network.Network{Kind: network.Fuji, ID: constants.FujiID}
	hrp, err := fuji.AddressHRP()
	require.NoError(t, err)
	require.Equal(t, constants.FujiHRP, hrp)

	chennai := network.Network{Kind: network.RinkTestnet, ID: network.ChennaiTestnetID}
	hrp, err = chennai.AddressHRP()
	require.NoError(t, err)
	require.Equal(t, network.ChennaiTestnetHRP, hrp)

	// Legacy callers that only set the ID still get the verified RINK HRP.
	legacy := network.Network{ID: network.RinkMainnetID}
	hrp, err = legacy.AddressHRP()
	require.NoError(t, err)
	require.Equal(t, network.RinkMainnetHRP, hrp)

	// A custom kind must not silently inherit a RINK HRP from its ID.
	custom := network.Network{Kind: network.Devnet, ID: network.RinkMainnetID}
	_, err = custom.AddressHRP()
	require.ErrorIs(t, err, network.ErrUnknownNetworkID)
}

func TestNetworkIDIsNotEVMChainID(t *testing.T) {
	mainnet := network.RinkMainnetNetwork()
	require.Equal(t, uint32(1), mainnet.ID)
	require.Equal(t, uint64(90059), mainnet.EVMChainID)
	require.NotEqual(t, uint64(mainnet.ID), mainnet.EVMChainID)

	// Supplying the EVM chain ID where the Avalanche network ID belongs must
	// not silently validate.
	confused := network.Network{Kind: network.RinkMainnet, ID: uint32(network.RinkMainnetEVMChainID)}
	require.ErrorIs(t, confused.Validate(), network.ErrNetworkIDMismatch)

	chennai := network.ChennaiTestnetNetwork()
	require.Equal(t, uint32(5), chennai.ID)
	require.Equal(t, uint64(2099), chennai.EVMChainID)

	local := network.RinkLocalNetwork()
	require.Equal(t, uint32(12345), local.ID)
	require.Equal(t, uint64(43112), local.EVMChainID)
}

func TestMissingEndpoint(t *testing.T) {
	mainnet := network.RinkMainnetNetwork()
	require.Empty(t, mainnet.Endpoint)
	require.ErrorIs(t, mainnet.ValidateEndpoint(), network.ErrMissingEndpoint)
	require.NoError(t, mainnet.WithEndpoint("http://127.0.0.1:9650").ValidateEndpoint())

	require.ErrorIs(t, network.UndefinedNetwork.ValidateEndpoint(), network.ErrUndefinedNetwork)
}

func TestUnknownNetworkRequiresExplicitHRP(t *testing.T) {
	custom := network.Network{Kind: network.Devnet, ID: 4242}
	_, err := custom.AddressHRP()
	require.ErrorIs(t, err, network.ErrUnknownNetworkID)
	require.Error(t, custom.Validate())

	custom = custom.WithHRP("custom")
	hrp, err := custom.AddressHRP()
	require.NoError(t, err)
	require.Equal(t, "custom", hrp)
	require.NoError(t, custom.Validate())
}

func TestAddressHRPValidation(t *testing.T) {
	addrBytes := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	net := network.RinkMainnetNetwork()

	pAddr, err := address.Format("P", net.HRP, addrBytes)
	require.NoError(t, err)
	require.Contains(t, pAddr, "P-rink1")

	chainAlias, parsedBytes, err := net.ParseAddress(pAddr)
	require.NoError(t, err)
	require.Equal(t, "P", chainAlias)
	require.Equal(t, addrBytes, parsedBytes)
	require.NoError(t, net.ValidateAddress(pAddr))
	require.NoError(t, net.ValidateAddresses([]string{pAddr}))

	xAddr, err := address.Format("X", net.HRP, addrBytes)
	require.NoError(t, err)
	require.NoError(t, net.ValidateAddress(xAddr))

	mismatched, err := address.Format("P", "avax", addrBytes)
	require.NoError(t, err)
	require.ErrorIs(t, net.ValidateAddress(mismatched), network.ErrHRPMismatch)
	require.ErrorIs(t, net.ValidateAddresses([]string{pAddr, mismatched}), network.ErrHRPMismatch)

	require.Error(t, net.ValidateAddress("not-an-address"))
}

func TestAddressHRPDerivedFromIDWhenUnset(t *testing.T) {
	net := network.Network{Kind: network.RinkMainnet, ID: network.RinkMainnetID}
	require.NoError(t, net.Validate())
	hrp, err := net.AddressHRP()
	require.NoError(t, err)
	require.Equal(t, network.RinkMainnetHRP, hrp)
}

func TestNetworkKindString(t *testing.T) {
	require.Equal(t, "rink-mainnet", network.RinkMainnet.String())
	require.Equal(t, "rink-testnet", network.RinkTestnet.String())
	require.Equal(t, "rink-local", network.RinkLocal.String())
	require.Equal(t, "undefined", network.Undefined.String())
}

func TestErrorsAreDistinct(t *testing.T) {
	require.False(t, errors.Is(network.ErrMissingEndpoint, network.ErrUndefinedNetwork))
	require.False(t, errors.Is(network.ErrHRPMismatch, network.ErrUnknownNetworkID))
}
