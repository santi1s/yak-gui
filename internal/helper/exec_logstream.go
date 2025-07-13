package helper

import (
	"bufio"
	"bytes"
	"fmt"
	"sync"

	"os/exec"

	log "github.com/sirupsen/logrus"
)

// ExecStreamLogger executes a command and streams stdout/stderr to the logger
// optionally captures stdout in a buffer and returns it instead of logging it
func ExecStreamLogger(cmd *exec.Cmd, logStdout bool) (bytes.Buffer, error) {
	var buf bytes.Buffer
	var wg sync.WaitGroup
	if logStdout {
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			return buf, err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				log.Info(scanner.Text())
			}
		}()
	} else {
		cmd.Stdout = &buf
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return buf, err
	}
	log.Debugf("%s> %s", cmd.Dir, cmd.String())
	if err := cmd.Start(); err != nil {
		return buf, err
	}

	scanner := bufio.NewScanner(stderrPipe)
	for scanner.Scan() {
		log.Error(scanner.Text())
	}
	wg.Wait()
	if err := cmd.Wait(); err != nil {
		return buf, err
	}
	if !cmd.ProcessState.Success() {
		return buf, fmt.Errorf("command failed with exit code %d", cmd.ProcessState.ExitCode())
	}
	return buf, nil
}
