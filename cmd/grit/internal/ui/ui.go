// Package ui styles CLI output per the Grit Cloud design guide §9: indigo for
// headers/prompts, cyan for live/streaming, green/amber/red for status, and a
// step timeline (✔ ⠹ ✖) mirroring the dashboard deploy screen. Colors are
// disabled automatically when stdout isn't a terminal or NO_COLOR is set.
package ui

import (
	"fmt"
	"os"
	"strings"
)

var enabled = colorEnabled()

func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	// Best-effort: honor FORCE_COLOR; otherwise assume a capable terminal.
	return true
}

// ANSI codes (256-color approximations of the brand palette).
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	indigo = "\033[38;5;99m"  // Orbita indigo
	cyan   = "\033[38;5;51m"  // accent / streaming
	green  = "\033[38;5;42m"  // success
	amber  = "\033[38;5;214m" // warning / in-progress
	red    = "\033[38;5;203m" // error
	slate  = "\033[38;5;245m" // neutral
)

func c(code, s string) string {
	if !enabled {
		return s
	}
	return code + s + reset
}

// Header prints an indigo banner line.
func Header(s string) {
	fmt.Println()
	fmt.Println(c(indigo+bold, "▸ "+s))
}

// Step prints a completed step (green check).
func Step(s string) { fmt.Println(c(green, "  ✔ ") + s) }

// StepActive prints an in-progress step (amber spinner glyph).
func StepActive(s string) { fmt.Println(c(amber, "  ⠹ ") + s) }

// StepFail prints a failed step (red cross).
func StepFail(s string) { fmt.Println(c(red, "  ✖ ") + s) }

// Info prints a neutral line.
func Info(s string) { fmt.Println("  " + s) }

// Live prints a cyan streaming line.
func Live(s string) { fmt.Println(c(cyan, "  ● ") + s) }

// Value renders a monospace-style value (bold) — IDs, URLs, tokens.
func Value(s string) string { return c(bold, s) }

// URL renders a link in cyan.
func URL(s string) string { return c(cyan, s) }

// Success prints a green banner.
func Success(s string) {
	fmt.Println()
	fmt.Println(c(green+bold, "✔ "+s))
}

// ErrorLine prints the two-line error pattern from the design guide: one red
// cause line, one dim fix line.
func ErrorLine(cause, fix string) {
	fmt.Fprintln(os.Stderr, c(red+bold, "✖ "+cause))
	if fix != "" {
		fmt.Fprintln(os.Stderr, c(dim, "  → "+fix))
	}
}

// Field prints an aligned "label: value" line.
func Field(label, value string) {
	fmt.Printf("  %s %s\n", c(slate, pad(label+":", 16)), value)
}

// Status renders a colored status pill (● label).
func Status(status string) string {
	dot := "●"
	switch strings.ToLower(status) {
	case "ok", "running", "healthy", "live", "success":
		return c(green, dot+" "+status)
	case "deploying", "pending", "creating", "updating", "warning":
		return c(amber, dot+" "+status)
	case "failed", "crashed", "error", "unhealthy":
		return c(red, dot+" "+status)
	default:
		return c(slate, dot+" "+status)
	}
}

func pad(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
