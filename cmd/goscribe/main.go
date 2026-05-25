// Package main is the goscribe CLI entry point.
package main

import (
	"fmt"
	"os"

	"github.com/house/goscribe/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if exitErr, ok := err.(*cmd.ExitCodeError); ok {
			os.Exit(exitErr.Code)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
