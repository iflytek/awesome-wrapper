// Package build contains build-related variables set at compile time.
package build

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/pyroscope-io/pyroscope/pkg/agent/spy"
)

var (
	Version = "N/A"
	ID      = "N/A"
	Time    = "N/A"

	GitSHA      = "N/A"
	GitDirtyStr = "-1"
	GitDirty    int

	UseEmbeddedAssetsStr = "false"
	UseEmbeddedAssets    bool
)

func init() {
	GitDirty, _ = strconv.Atoi(GitDirtyStr)
	UseEmbeddedAssets = UseEmbeddedAssetsStr == "true"
}

const tmplt = `
GENERAL
  GOARCH:             %s
  GOOS:               %s
  Version:            %s
  Build ID:           %s
  Build Time:         %s
  Git SHA:            %s
  Git Dirty Files:    %d
  Embedded Assets:    %t

AGENT
  Supported Spies:    %q
`

func Summary() string {
	return fmt.Sprintf(strings.TrimSpace(tmplt),
		runtime.GOARCH,
		runtime.GOOS,
		Version,
		ID,
		Time,
		GitSHA,
		GitDirty,
		UseEmbeddedAssets,
		spy.SupportedSpies,
	)
}
