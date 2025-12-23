package main

import (
	"os"

	"tunnel-monitor/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
