package main

import (
	"os"

	"github.com/doctolib/yak/internal/cmd/secret"
)

// entrypoint of yak-secret binary
func main() {
	yakSecretCmd := secret.GetRootCmd()
	yakSecretCmd.Use = "yak-secret"

	if err := yakSecretCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
