package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

var stdin = bufio.NewReader(os.Stdin)

// Ask prompts for a line of input, showing a default in brackets. Empty input
// returns the default.
func Ask(label, def string) string {
	if def != "" {
		fmt.Printf("%s %s ", c(indigo, "?")+" "+label, c(dim, "["+def+"]"))
	} else {
		fmt.Printf("%s ", c(indigo, "?")+" "+label)
	}
	line, _ := stdin.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

// AskRequired prompts until a non-empty value is given.
func AskRequired(label string) string {
	for {
		v := Ask(label, "")
		if strings.TrimSpace(v) != "" {
			return v
		}
		fmt.Println(c(dim, "  (required)"))
	}
}

// Confirm asks a yes/no question. def is the default when the user just hits
// Enter.
func Confirm(label string, def bool) bool {
	suffix := "[y/N]"
	if def {
		suffix = "[Y/n]"
	}
	for {
		fmt.Printf("%s %s ", c(indigo, "?")+" "+label, c(dim, suffix))
		line, _ := stdin.ReadString('\n')
		line = strings.ToLower(strings.TrimSpace(line))
		switch line {
		case "":
			return def
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println(c(dim, "  please answer y or n"))
		}
	}
}

// Secret prompts for hidden input (passwords). Falls back to visible input if
// stdin isn't a terminal.
func Secret(label string) string {
	fmt.Printf("%s ", c(indigo, "?")+" "+label)
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		fmt.Println()
		if err == nil {
			return strings.TrimSpace(string(b))
		}
	}
	line, _ := stdin.ReadString('\n')
	return strings.TrimSpace(line)
}

// Select presents a numbered menu and returns the chosen index (0-based).
func Select(label string, options []string, def int) int {
	fmt.Println(c(indigo, "?") + " " + label)
	for i, o := range options {
		marker := " "
		if i == def {
			marker = c(indigo, "›")
		}
		fmt.Printf("  %s %s %s\n", marker, c(dim, strconv.Itoa(i+1)+")"), o)
	}
	for {
		fmt.Printf("  choose %s ", c(dim, fmt.Sprintf("[1-%d, default %d]", len(options), def+1)))
		line, _ := stdin.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			return def
		}
		n, err := strconv.Atoi(line)
		if err == nil && n >= 1 && n <= len(options) {
			return n - 1
		}
		fmt.Println(c(dim, "  enter a number from the list"))
	}
}
