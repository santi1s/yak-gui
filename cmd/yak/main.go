package main

import (
	"os"

	"github.com/doctolib/yak/internal/cmd/root"
)

// entrypoint of yak binary
func main() {
	yakCmd := root.GetRootCmd()

	if err := yakCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
