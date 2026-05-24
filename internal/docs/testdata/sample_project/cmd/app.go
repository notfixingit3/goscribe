package cmd

import (
	"fmt"
	"os"
)

// RootCommand prints a welcome message and exits.
func RootCommand() {
	fmt.Fprintln(os.Stdout, "Welcome to the sample app")
}
