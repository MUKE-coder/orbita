// Command grit is the Grit Cloud CLI. In production these `cloud` and `deploy`
// subcommands are part of the existing Grit framework CLI; here they are built
// as a standalone binary against the Orbita control plane. The Grit-awareness
// logic (grit.yaml/grit.json parsing, detection, build planning) is shared with
// Orbita via the internal/grit package.
package main

import (
	"fmt"
	"os"

	"github.com/orbita-sh/orbita/cmd/grit/cmd"
)

func main() {
	if err := cmd.Root().Execute(); err != nil {
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
}
