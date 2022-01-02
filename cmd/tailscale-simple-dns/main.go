package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	rootCommand, _ := NewRootCmd(os.Stdout)

	if err := rootCommand.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
