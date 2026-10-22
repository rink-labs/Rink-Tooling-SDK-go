# RINK Tooling Go SDK

RINK's fork of the [Avalanche Tooling Go SDK](https://github.com/ava-labs/avalanche-tooling-sdk-go), adapted for the RINK network.

- Module: `github.com/rink-labs/Rink-Tooling-SDK-go`
- Upstream project: `github.com/ava-labs/avalanche-tooling-sdk-go`
- This repository was previously mirrored as `Rinkuby-Labs/Rinkuby-Tooling-SDK-go`; the canonical module path is `github.com/rink-labs/Rink-Tooling-SDK-go`.

The SDK is experimental. Public APIs may change between releases.

## Installation

```bash
go get github.com/rink-labs/Rink-Tooling-SDK-go
```

Until the fork publishes a tagged release, use the branch explicitly:

```bash
go get github.com/rink-labs/Rink-Tooling-SDK-go@main
```

## RINK network configuration

All network IDs, address HRPs, and EVM chain IDs below are defined in `network/network.go` and were verified against the RINK node configuration:

| Network | Constructor | Avalanche network ID | Address HRP | C-Chain EVM chain ID | Default RPC endpoint |
| --- | --- | ---: | --- | ---: | --- |
| RINK mainnet | `network.RinkMainnetNetwork()` | `1` | `rink` | `90059` | none configured |
| Chennai testnet | `network.ChennaiTestnetNetwork()` | `5` | `chennai` | `2099` | none configured |
| Local RINK | `network.RinkLocalNetwork()` | `12345` | `local` | `43112` | `http://127.0.0.1:9650` |
| Avalanche Fuji (explicit) | `network.FujiNetwork()` | `5` | `fuji` | `43113` | `https://api.avax-test.network` |

Sources in the RINK node repository (`rink-labs/rinkgo`):

- `utils/constants/network_ids.go` for network IDs and HRPs.
- `genesis/genesis_rink.json` and `genesis/genesis_chennai.json` for genesis network IDs.
- `graft/coreth/params/config.go` for `RinkMainnetChainID`, `ChennaiChainID`, and `AvalancheLocalChainID`.

The value `64293` was suggested as a possible RINK network ID, but it does not appear in the RINK genesis or network configuration. It is not used by this SDK. See [Unverified values and limitations](#unverified-values-and-limitations).

### Network ID is not EVM chain ID

These are separate identifiers:

- `Network.ID` is the Avalanche network ID used by P-Chain/X-Chain transactions and address HRP selection.
- `Network.EVMChainID` is the C-Chain EVM `chainId`.

For RINK mainnet they are `1` and `90059`; do not use one where the other is required.

### Endpoints are explicit

The SDK does not invent a public RINK RPC endpoint. RINK mainnet and Chennai testnet constructors leave `Endpoint` empty. Supply an endpoint before creating a wallet or making RPC calls:

```go
package main

import (
	"fmt"
	"os"

	"github.com/rink-labs/Rink-Tooling-SDK-go/network"
)

func main() {
	endpoint := os.Getenv("RINK_RPC_ENDPOINT")
	rink := network.RinkMainnetNetwork().WithEndpoint(endpoint)
	if err := rink.ValidateEndpoint(); err != nil {
		panic(fmt.Errorf("RINK network is not configured: %w", err))
	}

	fmt.Println("network ID:", rink.ID)
	fmt.Println("C-Chain EVM chain ID:", rink.EVMChainID)
}
```

For a local node, `network.RinkLocalNetwork()` already has the standard endpoint. To use Chennai or a custom deployment:

```go
chennai := network.ChennaiTestnetNetwork().WithEndpoint(os.Getenv("RINK_TESTNET_RPC_ENDPOINT"))
custom := network.NewNetwork(network.Devnet, 4242, "http://127.0.0.1:9650").WithHRP("custom")
```

`Validate()` checks network/HRP configuration. `ValidateEndpoint()` additionally requires a non-empty RPC endpoint.

## Addresses and HRPs

P/X chain addresses are bech32 strings with a network-specific human-readable part:

- RINK mainnet: `P-rink1...` / `X-rink1...`
- Chennai testnet: `P-chennai1...` / `X-chennai1...`
- Local RINK: `P-local1...` / `X-local1...`

The SDK uses the HRP on the `Network` object rather than blindly calling the upstream AvalancheGo `constants.GetHRP`. This matters because upstream maps network ID `1` to `avax`, while RINK mainnet uses `rink`.

Address parsing and validation are available on `Network`:

```go
rink := network.RinkMainnetNetwork()

// acct can be any account.Account implementation.
addr, err := acct.GetPChainAddress(rink)
if err != nil {
	panic(err)
}
if err := rink.ValidateAddress(addr); err != nil {
	panic(err) // includes network.ErrHRPMismatch for a different network
}
```

`ValidateAddress` and `ParseAddress` reject an address whose HRP does not match the network. P-Chain transaction builders validate supplied control keys, subnet auth keys, and validator owner addresses before building a transaction. Address encoding/checksum logic remains in AvalancheGo; this fork only supplies the correct HRP.

### Fuji is explicit, not a default

The fork removes implicit Fuji selection:

- `network.RinkMainnetNetwork()` is the RINK mainnet constructor.
- `NetworkFromNetworkID(1)` returns RINK mainnet.
- `NetworkFromNetworkID(5)` returns RINK Chennai, not Avalanche Fuji.
- `network.FujiNetwork()` remains available for explicit upstream compatibility, but it is not selected by default and cannot be recovered from network ID 5 because that ID is shared with Chennai.

The RinkGo node fork defines its own `FujiID` as `6` to avoid the Chennai collision. This SDK does not use that constant: `FujiNetwork()` targets the upstream Avalanche Fuji API and public archive, whose network ID is `5`. RINK deployments should use one of the RINK constructors instead.

If you retain Fuji code paths, select them explicitly and keep RINK network configuration separate. The public archive downloader remains Fuji-only because the upstream archive URL is Fuji-specific; it now requires `network.FujiNetwork()` rather than matching a raw network ID.

## Minimal example

```go
package main

import (
	"fmt"
	"os"

	"github.com/rink-labs/Rink-Tooling-SDK-go/account"
	"github.com/rink-labs/Rink-Tooling-SDK-go/network"
)

func main() {
	rink := network.RinkMainnetNetwork().WithEndpoint(os.Getenv("RINK_RPC_ENDPOINT"))
	if err := rink.ValidateEndpoint(); err != nil {
		panic(err)
	}

	acct, err := account.NewLocalAccount()
	if err != nil {
		panic(err)
	}

	pChainAddress, err := acct.GetPChainAddress(rink)
	if err != nil {
		panic(err)
	}

	fmt.Println(pChainAddress) // P-rink1...
}
```

More examples are in `example/`, `examples/`, and `interchain/`. Examples that connect to a node read `RINK_RPC_ENDPOINT`; no example silently connects to Fuji.

## Development and verification

Use the same toolchain as `go.mod`:

```bash
gofmt -w .
go mod tidy
go mod verify
go test ./...
go vet ./...
```

The public archive end-to-end test downloads a large external archive and is skipped unless explicitly enabled:

```bash
RUN_PUBLIC_ARCHIVE_E2E=1 go test ./publicarchive -run TestDownloader_EndToEnd
```

## Upgrading from upstream

This fork tracks the upstream Avalanche Tooling SDK. To update it:

1. Fetch the desired upstream release or commit.
2. Merge it into a feature branch without dropping RINK-specific changes.
3. Keep the module path as `github.com/rink-labs/Rink-Tooling-SDK-go`.
4. Replace only SDK self-imports; leave third-party `github.com/ava-labs/*` imports intact. In particular, this SDK currently depends on upstream AvalancheGo and does not use a `replace` directive for RinkGo.
5. Recheck `network/network.go`, `key/soft_key.go`, and `keychain/cubesigner/cubesigner_keychain.go` for RINK ID/HRP behavior.
6. Re-run `gofmt`, `go mod tidy`, `go mod verify`, `go test ./...`, and `go vet ./...`.
7. Revalidate RINK network values against the RINK node repository before changing constants.

## Unverified values and limitations

- **`64293` is not a verified RINK network ID.** The verified RINK mainnet network ID is `1`.
- **No public RINK RPC endpoint is configured.** Callers must provide one. Local examples use `http://127.0.0.1:9650`.
- **No RINK genesis or blockchain IDs are bundled.** Obtain them from the deployment/node or a trusted configuration source; do not hard-code unverified values.
- **Fuji collision:** upstream Avalanche Fuji and RINK Chennai both use network ID `5`. Explicit Fuji selection must not depend on `NetworkFromNetworkID(5)`.
- **CubeSigner and Fuji:** CubeSigner signing resolves the address HRP from the transaction's network ID alone. Network ID `5` is treated as Chennai, so CubeSigner cannot distinguish an explicit Avalanche Fuji transaction in this fork.
- **Upstream dependency:** this repository depends on upstream AvalancheGo modules. RinkGo is a separate node fork; compatibility of node-specific behavior must be validated against the target RINK node version.
- **Public archive:** only the Fuji archive URL is known; RINK archive support requires a verified URL.

## License

See [LICENSE](LICENSE). Upstream copyright and license notices are retained.
