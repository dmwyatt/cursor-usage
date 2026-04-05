package main

import (
	"os"

	"github.com/dmwyatt/cursor-usage/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
