package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Read reads a value from the current stdin until it finds a newline or a space.
// It returns the value and any encountered error.
func (c *Cli) Read() (string, error) {
	var value string
	_, err := fmt.Fscanln(c.stdin, &value)
	if err != nil {
		return "", err
	}
	return value, nil
}

// Read reads a value from the current stdin until it finds a newline or a space.
// It returns the value and any encountered error.
func Read() (string, error) {
	return c.Read()
}

// ReadLine reads a value from the current stdin until it finds a newline.
// It returns the value and any encountered error.
func (c *Cli) ReadLine() (string, error) {
	reader := bufio.NewReader(c.stdin)
	s := bufio.NewScanner(reader)
	s.Scan()
	return strings.TrimSuffix(s.Text(), "\n"), s.Err()
}

// ReadLine reads a value from the current stdin until it finds a newline.
// It returns the value and any encountered error.
func ReadLine() (string, error) {
	return c.ReadLine()
}

// ReadAll reads a value from the current stdin.
// It returns the value and any encountered error.
func (c *Cli) ReadAll() (string, error) {
	reader := bufio.NewReader(c.stdin)

	if i, ok := c.stdin.(interface{ Stat() (os.FileInfo, error) }); ok {
		fi, err := i.Stat()
		if err != nil {
			panic(err)
		}
		// check if nothing is piped to the command
		if fi.Mode()&os.ModeNamedPipe == 0 {
			return "", nil
		}
	}

	bytes, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ReadAll reads a value from the current stdin.
// It returns the value and any encountered error.
func ReadAll() (string, error) {
	return c.ReadAll()
}

// ReadPassword reads a password securely from the current stdin.
// It returns the value and any encountered error.
func (c *Cli) ReadPassword() (string, error) {
	return c.passwordReader.ReadPassword()
}

// ReadPassword reads a password securely from the current stdin.
// It returns the value and any encountered error.
func ReadPassword() (string, error) {
	return c.ReadPassword()
}

// AskConfirmation prompts the given message and ask for a yes/no confirmation on the current stdin.
// If an invalid input is given, then the confirmation is prompted again until a valid input.
// It returns a bool value depending of the yes/no answer.
func (c *Cli) AskConfirmation(message string) bool {
	if !c.skipConfirmation {
		var input string

		for {
			c.Printf("%s [y/n] ", message)
			if _, err := fmt.Fscanln(c.stdin, &input); err != nil {
				c.Println(err)
				continue
			}

			switch input {
			case "y", "Y", "yes", "YES":
				return true
			case "n", "N", "no", "NO":
				return false
			}
		}
	}
	return true
}

// AskConfirmation prompts the given message and ask for a yes/no confirmation on the current stdin.
// If an invalid input is given, then the confirmation is prompted again until a valid input.
// It returns a bool value depending of the yes/no answer.
func AskConfirmation(message string) bool {
	return c.AskConfirmation(message)
}
