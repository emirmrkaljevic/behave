package main

import (
	"fmt"
	"os"
)

const version = "0.0.1-dev"

const usage = `behave - record how a CLI behaves, prove a rewrite behaves the same

Usage:
  behave probe <cmd>              explore a command and propose cases
  behave record <cmd> [--out dir] record golden cases
  behave run <spec-dir>           replay cases against the original
  behave diff <spec-dir> --impl <bin>  replay cases against another implementation
  behave mcp                      start MCP server (stdio)
  behave version
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "version", "--version":
		fmt.Println("behave", version)
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "behave: %q not implemented yet\n", os.Args[1])
		os.Exit(1)
	}
}
