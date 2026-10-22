// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package rinktoolingsdk_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const modulePath = "github.com/rink-labs/Rink-Tooling-SDK-go"

func TestModulePath(t *testing.T) {
	goMod, err := os.ReadFile("go.mod")
	require.NoError(t, err)
	require.Contains(t, string(goMod), "module "+modulePath)
}

func TestGoFilesUseRinkModulePath(t *testing.T) {
	// Construct the old path so this test file does not match itself.
	oldModulePath := strings.Join([]string{"github.com", "ava-labs", "avalanche-tooling-sdk-go"}, "/")

	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		require.NotContains(t, string(contents), oldModulePath, "legacy module import in %s", path)
		return nil
	})
	require.NoError(t, err)
}
