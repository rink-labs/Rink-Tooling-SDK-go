# Public Archive Downloader SDK

This Go package provides a utility to download and extract tar archives from public URLs. It currently supports the upstream Avalanche Fuji public archive only. No verified RINK archive URL is shipped with this fork, so RINK archive support must be added explicitly once a trusted URL is available.


## Features

* Downloads files from predefined URLs.
* Tracks download progress and logs status updates.
* Safely unpacks .tar archives to a target directory.
* Includes security checks to prevent path traversal and manage large files.

## Usage example

```go
// Copyright (C) 2025, Ava Labs, Inc. All rights reserved
// See the file LICENSE for licensing terms.

package main

import (
	"fmt"
	"os"

	"github.com/ava-labs/avalanchego/utils/logging"

	"github.com/rink-labs/Rink-Tooling-SDK-go/network"
	"github.com/rink-labs/Rink-Tooling-SDK-go/publicarchive"
)

func main() {
	// Initialize the downloader
	downloader, err := publicarchive.NewDownloader(network.FujiNetwork(), logging.NoLog{})
	if err != nil {
		fmt.Printf("Failed to create downloader: %v\n", err)
		os.Exit(1)
	}

	// Start downloading
	if err := downloader.Download(); err != nil {
		fmt.Printf("Download failed: %v\n", err)
		os.Exit(1)
	}

	// Specify the target directory for unpacking
	targetDir := "./extracted_files"
	if err := downloader.UnpackTo(targetDir); err != nil {
		fmt.Printf("Failed to unpack archive: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Files successfully unpacked to %s\n", targetDir)
}
```
