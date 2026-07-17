// Command orbita is the CLI for the Orbita PaaS: provision a host, then deploy
// any containerisable app to it (a Dockerfile repo, a Compose file, a raw image,
// or anything Nixpacks can build). Grit apps get a zero-config fast path —
// Orbita detects grit.json and reuses the Dockerfiles Grit ships — but Grit is
// never a prerequisite.
//
// Orbita depends on nothing from the Grit framework or its CLI.
package main

import (
	"fmt"
	"os"

	"github.com/orbita-sh/orbita/cmd/orbita/cmd"
)

func main() {
	if err := cmd.Root().Execute(); err != nil {
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
}
