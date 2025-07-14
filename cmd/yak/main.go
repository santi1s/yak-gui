package main

import (
	"os"

	"github.com/santi1s/yak/internal/cmd/root"
)

// entrypoint of yak binary
func main() {
	yakCmd := root.GetRootCmd()

	if err := yakCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
