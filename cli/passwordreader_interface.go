package cli

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// PasswordReader is an interface representing what must be implemented to be a PasswordReader
type PasswordReader interface {
	ReadPassword() (string, error)
}

// TermPasswordReader is an implementation of PasswordReader that reads a password securely from stdin
type TermPasswordReader struct {
}

// ReadPassword reads a password securely from stdin.
// It returns the string given by the user on stdin and any error encountered.
func (pr *TermPasswordReader) ReadPassword() (string, error) {
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	return string(password), err
}

// IoReaderPasswordReader is an implementation of PasswordReader thas reads a password from a standard io.Reader
type IoReaderPasswordReader struct {
	Reader io.Reader
}

// ReadPassword reads a password from the Reader io.Reader of IoReaderPasswordReader.
// It returns the string given by the user from the Reader io.Reader and any error encountered.
func (pr *IoReaderPasswordReader) ReadPassword() (string, error) {
	var password string
	_, err := fmt.Fscanln(pr.Reader, &password)
	if err != nil {
		return "", err
	}
	return password, nil
}
