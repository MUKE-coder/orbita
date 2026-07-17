// Package assets embeds files the CLI ships (the vendored VPS hardening script).
package assets

import _ "embed"

// VPSHarden is the vendored hardening script from github.com/MUKE-coder/vps-harden.
// orbita init uploads and runs it with --no-dokploy (Orbita replaces Dokploy).
//
//go:embed vps-harden.sh
var VPSHarden []byte
