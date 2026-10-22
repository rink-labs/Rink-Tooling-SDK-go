// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package network defines the RINK networks supported by this SDK.
//
// A Network intentionally keeps the Avalanche network ID separate from the
// EVM chain ID. RINK mainnet, for example, uses Avalanche network ID 1 but its
// C-Chain uses EVM chain ID 90059.
package network

import (
	"errors"
	"fmt"

	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
)

type NetworkKind int64

const (
	Undefined NetworkKind = iota
	// Mainnet identifies RINK mainnet. The name is retained for compatibility
	// with the upstream SDK; this fork does not select Avalanche mainnet.
	Mainnet
	// Fuji identifies Avalanche Fuji. It is only reachable through the explicit
	// FujiNetwork constructor because network ID 5 is used by RINK Chennai.
	Fuji
	Devnet
	Granite
	// RinkTestnet identifies the RINK Chennai testnet.
	RinkTestnet
	// RinkLocal identifies a local RINK network.
	RinkLocal
)

// RinkMainnet is an explicit alias for the RINK mainnet kind.
const RinkMainnet NetworkKind = Mainnet

func (k NetworkKind) String() string {
	switch k {
	case Mainnet:
		return "rink-mainnet"
	case Fuji:
		return "fuji"
	case Devnet:
		return "devnet"
	case Granite:
		return "granite"
	case RinkTestnet:
		return "rink-testnet"
	case RinkLocal:
		return "rink-local"
	default:
		return "undefined"
	}
}

const (
	// RinkMainnetID is the Avalanche network ID in the RINK mainnet genesis.
	RinkMainnetID uint32 = 1
	// ChennaiTestnetID is the Avalanche network ID in the Chennai testnet genesis.
	ChennaiTestnetID uint32 = 5
	// RinkLocalID is the Avalanche network ID used by a local RINK network.
	RinkLocalID uint32 = 12345

	GraniteNetworkID uint32 = 76
)

const (
	// RinkMainnetHRP is the bech32 human-readable part used by RINK mainnet P/X
	// chain addresses.
	RinkMainnetHRP = "rink"
	// ChennaiTestnetHRP is the bech32 human-readable part used by Chennai P/X
	// chain addresses.
	ChennaiTestnetHRP = "chennai"
	// RinkLocalHRP is the bech32 human-readable part used by local RINK P/X
	// chain addresses.
	RinkLocalHRP = "local"
)

const (
	// RinkMainnetEVMChainID is the C-Chain EVM chain ID in genesis_rink.json.
	RinkMainnetEVMChainID uint64 = 90059
	// ChennaiTestnetEVMChainID is the Chennai C-Chain EVM chain ID.
	ChennaiTestnetEVMChainID uint64 = 2099
	// RinkLocalEVMChainID is the local C-Chain EVM chain ID.
	RinkLocalEVMChainID uint64 = 43112
	// FujiEVMChainID is Avalanche Fuji's C-Chain EVM chain ID. It is unrelated
	// to the Avalanche Fuji network ID.
	FujiEVMChainID uint64 = 43113
)

const (
	FujiAPIEndpoint = "https://api.avax-test.network"
	// MainnetAPIEndpoint is intentionally empty: no public RINK mainnet RPC
	// endpoint has been verified. Configure one with Network.WithEndpoint.
	//
	// Deprecated: use RinkMainnetNetwork().WithEndpoint(...).
	MainnetAPIEndpoint = ""
	GraniteEndpoint    = "https://granite.avax-dev.network"
	// RinkLocalAPIEndpoint is the default endpoint of a local RINK node.
	RinkLocalAPIEndpoint = "http://127.0.0.1:9650"
)

var (
	// ErrUndefinedNetwork is returned when an operation requires a network but
	// the zero value was supplied.
	ErrUndefinedNetwork = errors.New("network is undefined")
	// ErrUnknownNetworkID is returned when an HRP cannot be derived for a
	// network ID because it is not a known RINK network.
	ErrUnknownNetworkID = errors.New("unknown network ID")
	// ErrMissingEndpoint is returned when an operation needs an RPC endpoint but
	// the network does not have one configured.
	ErrMissingEndpoint = errors.New("network endpoint is not configured")
	// ErrNetworkIDMismatch is returned when a network kind and network ID
	// disagree, for example when an EVM chain ID is supplied as a network ID.
	ErrNetworkIDMismatch = errors.New("network ID does not match network kind")
	// ErrHRPMismatch is returned when an address belongs to a different network
	// than the one it is being validated against.
	ErrHRPMismatch = errors.New("address HRP does not match network")
)

// Network describes an Avalanche-compatible network used by the SDK.
type Network struct {
	Kind NetworkKind
	// ID is the Avalanche network ID. It is not the EVM chain ID.
	ID uint32
	// Endpoint is the node/API endpoint. It may be empty for networks whose
	// public endpoint has not been configured by the caller.
	Endpoint string
	// HRP is the bech32 human-readable part for P/X chain addresses.
	HRP string
	// EVMChainID is the C-Chain EVM chain ID. Zero means unknown.
	EVMChainID uint64
}

var UndefinedNetwork = Network{}

// NetworkFromNetworkID maps an Avalanche network ID to a RINK network.
//
// Network ID 1 is RINK mainnet and network ID 5 is the Chennai testnet. The
// upstream Avalanche Fuji network also uses ID 5; explicit Fuji selection must
// therefore use FujiNetwork rather than this function.
func NetworkFromNetworkID(networkID uint32) Network {
	switch networkID {
	case RinkMainnetID:
		return RinkMainnetNetwork()
	case ChennaiTestnetID:
		return ChennaiTestnetNetwork()
	case RinkLocalID:
		return RinkLocalNetwork()
	case GraniteNetworkID:
		return GraniteNetwork()
	}
	return UndefinedNetwork
}

// NewNetwork creates a network and fills in the verified RINK HRP and EVM chain
// ID for the supplied kind. Unknown kinds keep an empty HRP and chain ID 0.
func NewNetwork(kind NetworkKind, id uint32, endpoint string) Network {
	return Network{
		Kind:       kind,
		ID:         id,
		Endpoint:   endpoint,
		HRP:        defaultHRP(kind),
		EVMChainID: defaultEVMChainID(kind),
	}
}

// WithEndpoint returns a copy of the network with the supplied endpoint.
func (n Network) WithEndpoint(endpoint string) Network {
	n.Endpoint = endpoint
	return n
}

// WithHRP returns a copy of the network with the supplied address HRP. It is
// intended for custom/dev networks; the RINK constructors already set the
// verified HRPs.
func (n Network) WithHRP(hrp string) Network {
	n.HRP = hrp
	return n
}

// RinkMainnetNetwork returns RINK mainnet. No public RPC endpoint has been
// verified, so Endpoint is empty until the caller supplies one with
// WithEndpoint.
func RinkMainnetNetwork() Network {
	return NewNetwork(Mainnet, RinkMainnetID, "")
}

// MainnetNetwork is retained as an alias for RinkMainnetNetwork.
//
// Deprecated: use RinkMainnetNetwork.
func MainnetNetwork() Network {
	return RinkMainnetNetwork()
}

// ChennaiTestnetNetwork returns the RINK Chennai testnet. No public RPC
// endpoint has been verified, so Endpoint is empty until the caller supplies
// one with WithEndpoint.
func ChennaiTestnetNetwork() Network {
	return NewNetwork(RinkTestnet, ChennaiTestnetID, "")
}

// RinkTestnetNetwork is an alias for ChennaiTestnetNetwork.
func RinkTestnetNetwork() Network {
	return ChennaiTestnetNetwork()
}

// RinkLocalNetwork returns a local RINK network using the standard local node
// endpoint.
func RinkLocalNetwork() Network {
	return NewNetwork(RinkLocal, RinkLocalID, RinkLocalAPIEndpoint)
}

// FujiNetwork returns Avalanche Fuji explicitly. This fork does not use Fuji as
// a default. Because RINK Chennai uses network ID 5, Fuji is only selected
// through this constructor and not through NetworkFromNetworkID.
func FujiNetwork() Network {
	return NewNetwork(Fuji, constants.FujiID, FujiAPIEndpoint)
}

// GraniteNetwork returns the upstream Granite devnet configuration.
func GraniteNetwork() Network {
	return NewNetwork(Granite, GraniteNetworkID, GraniteEndpoint)
}

// AddressHRP returns the configured address HRP, or derives one from the
// network kind. For legacy callers that only set ID, it falls back to the
// verified RINK ID mapping. The kind takes precedence because RINK Chennai and
// Avalanche Fuji share network ID 5 but use different HRPs.
func (n Network) AddressHRP() (string, error) {
	if n.HRP != "" {
		return n.HRP, nil
	}
	if hrp := defaultHRP(n.Kind); hrp != "" {
		return hrp, nil
	}
	if n.Kind == Undefined {
		if hrp, ok := HRPForNetworkID(n.ID); ok {
			return hrp, nil
		}
	}
	return "", fmt.Errorf("%w: %d", ErrUnknownNetworkID, n.ID)
}

// Validate checks that the network has enough information for address
// operations. An RPC endpoint is not required by Validate; call
// ValidateEndpoint for operations that connect to a node.
func (n Network) Validate() error {
	if n.Kind == Undefined {
		return ErrUndefinedNetwork
	}
	if expectedID, ok := networkIDForKind(n.Kind); ok && n.ID != expectedID {
		return fmt.Errorf("%w: %s requires %d, got %d", ErrNetworkIDMismatch, n.Kind, expectedID, n.ID)
	}
	if _, err := n.AddressHRP(); err != nil {
		return err
	}
	return nil
}

// ValidateEndpoint checks the network and ensures that an RPC endpoint is set.
func (n Network) ValidateEndpoint() error {
	if err := n.Validate(); err != nil {
		return err
	}
	if n.Endpoint == "" {
		return fmt.Errorf("%w for %s", ErrMissingEndpoint, n.name())
	}
	return nil
}

// ValidateAddress checks that addr is a valid P/X chain address whose HRP
// matches this network.
func (n Network) ValidateAddress(addr string) error {
	_, _, err := n.ParseAddress(addr)
	return err
}

// ValidateAddresses checks each address in addrs against this network.
func (n Network) ValidateAddresses(addrs []string) error {
	for _, addr := range addrs {
		if err := n.ValidateAddress(addr); err != nil {
			return err
		}
	}
	return nil
}

// ParseAddress parses a P/X chain address and verifies its HRP against this
// network. It returns the chain alias ("P" or "X") and the address bytes.
func (n Network) ParseAddress(addr string) (string, []byte, error) {
	chainAlias, hrp, addrBytes, err := address.Parse(addr)
	if err != nil {
		return "", nil, err
	}
	expectedHRP, err := n.AddressHRP()
	if err != nil {
		return "", nil, err
	}
	if hrp != expectedHRP {
		return "", nil, fmt.Errorf("%w: expected %q, got %q", ErrHRPMismatch, expectedHRP, hrp)
	}
	return chainAlias, addrBytes, nil
}

// HRPForNetworkID returns the verified RINK HRP for a network ID. It returns
// false for IDs that are not one of the RINK networks supported by this SDK.
func HRPForNetworkID(networkID uint32) (string, bool) {
	switch networkID {
	case RinkMainnetID:
		return RinkMainnetHRP, true
	case ChennaiTestnetID:
		return ChennaiTestnetHRP, true
	case RinkLocalID:
		return RinkLocalHRP, true
	default:
		return "", false
	}
}

func networkIDForKind(kind NetworkKind) (uint32, bool) {
	switch kind {
	case Mainnet:
		return RinkMainnetID, true
	case RinkTestnet:
		return ChennaiTestnetID, true
	case RinkLocal:
		return RinkLocalID, true
	case Fuji:
		return constants.FujiID, true
	case Granite:
		return GraniteNetworkID, true
	default:
		return 0, false
	}
}

func defaultHRP(kind NetworkKind) string {
	switch kind {
	case Mainnet:
		return RinkMainnetHRP
	case RinkTestnet:
		return ChennaiTestnetHRP
	case RinkLocal:
		return RinkLocalHRP
	case Fuji:
		return constants.FujiHRP
	case Granite:
		// The Granite HRP is not part of RINK configuration; retain upstream
		// behavior only for this explicit, non-RINK network.
		return constants.GetHRP(GraniteNetworkID)
	default:
		return ""
	}
}

func defaultEVMChainID(kind NetworkKind) uint64 {
	switch kind {
	case Mainnet:
		return RinkMainnetEVMChainID
	case RinkTestnet:
		return ChennaiTestnetEVMChainID
	case RinkLocal:
		return RinkLocalEVMChainID
	case Fuji:
		return FujiEVMChainID
	default:
		return 0
	}
}

func (n Network) name() string {
	switch n.Kind {
	case Mainnet:
		return "RINK mainnet"
	case RinkTestnet:
		return "RINK Chennai testnet"
	case RinkLocal:
		return "RINK local network"
	case Fuji:
		return "Avalanche Fuji"
	case Granite:
		return "Granite"
	default:
		return fmt.Sprintf("network ID %d", n.ID)
	}
}
