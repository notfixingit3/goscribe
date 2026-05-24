package main

import (
	"fmt"
)

func main() {
	fmt.Println(greet("World"))
}

// greet returns a friendly greeting.
func greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}
