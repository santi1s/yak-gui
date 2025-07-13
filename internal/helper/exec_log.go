package helper

import (
	"bytes"
	"errors"
	"fmt"

	"os/exec"

	log "github.com/sirupsen/logrus"
)

// ExecLogger executes a command and prints logs to the adequate log level
// optionally log stdout if command is successful
func ExecLogger(cmd *exec.Cmd, logStdout bool) (bytes.Buffer, error) {
	var stdout, stderr bytes.Buffer
	var exitError *exec.ExitError
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	log.Debugf("%s> %s", cmd.Dir, cmd.String())
	if err := cmd.Run(); err != nil {
		if stdout.Len() > 0 {
			log.Error(stdout.String())
		}
		if stderr.Len() > 0 {
			log.Error(stderr.String())
		}
		if errors.As(err, &exitError) {
			return stdout, fmt.Errorf("command failed with exit code %d", exitError.ExitCode())
		} else {
			return stdout, err
		}
	}

	if logStdout && stdout.Len() > 0 {
		log.Info(stdout.String())
	}
	log.Info(stderr.String())
	return stdout, nil
}
