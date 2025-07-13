// Package cli provides a convenient way to interact with stdin, stderr and stdout.
package cli

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

// Cli is a user input/output interaction interface.
// It's mainly used to print output to and get input
// from a user in a CLI environment in a configurable way.
type Cli struct {
	// skipConfirm is a bool defined by the user to disable confirmation prompts from some functions
	skipConfirmation bool
	// stderr is a writer defined by the user that replaces stderr
	stderr io.Writer
	// stdout is a writer defined by the user that replaces stdout
	stdout io.Writer
	// stdin is a reader defined by the user that replaces stdin
	stdin io.Reader
	// passwordReader is a PasswordReader defined by the user that will be used to read passwords securely from input
	passwordReader PasswordReader
}

// c is a default initialized Cli instance that can be
// directly used in most cases.
var c *Cli = &Cli{
	skipConfirmation: false,
	stderr:           os.Stderr,
	stdout:           os.Stdout,
	stdin:            os.Stdin,
	passwordReader:   &TermPasswordReader{},
}

// GetOut returns the current io.Writer used as stdout.
func (c *Cli) GetOut() io.Writer {
	if c.stdout != nil {
		return c.stdout
	}
	return nil
}

// GetOut returns the current io.Writer used as stdout.
func GetOut() io.Writer {
	return c.GetOut()
}

// SetOut set the io.Writer used as stdout if the provided io.Writer is not nil.
func (c *Cli) SetOut(w io.Writer) {
	if w != nil {
		c.stdout = w
	}
}

// SetOut set the io.Writer used as stdout if the provided io.Writer is not nil.
func SetOut(w io.Writer) {
	c.SetOut(w)
}

// GetErr returns the current io.Writer used as stderr.
func (c *Cli) GetErr() io.Writer {
	if c.stderr != nil {
		return c.stderr
	}
	return nil
}

// GetErr returns the current io.Writer used as stderr.
func GetErr() io.Writer {
	return c.GetErr()
}

// SetErr set the io.Writer used as stderr if the provided io.Writer is not nil.
func (c *Cli) SetErr(w io.Writer) {
	if w != nil {
		c.stderr = w
	}
}

// SetErr set the io.Writer used as stderr if the provided io.Writer is not nil.
func SetErr(w io.Writer) {
	c.SetErr(w)
}

// GetIn returns the current io.Reader used as stdin.
func (c *Cli) GetIn() io.Reader {
	if c.stdin != nil {
		return c.stdin
	}
	return nil
}

// GetIn returns the current io.Reader used as stdin.
func GetIn() io.Reader {
	return c.GetIn()
}

// SetIn set the io.Reader used as stdin if the provided io.Reader is not nil.
func (c *Cli) SetIn(r io.Reader) {
	if r != nil {
		c.stdin = r
	}
}

// SetIn set the io.Reader used as stdin if the provided io.Reader is not nil.
func SetIn(r io.Reader) {
	c.SetIn(r)
}

// SetPasswordReader set the PasswordReader used if the provided PasswordReader is not nil.
func (c *Cli) SetPasswordReader(pr PasswordReader) {
	if pr != nil {
		c.passwordReader = pr
	}
}

// SetPasswordReader set the PasswordReader used if the provided PasswordReader is not nil.
func SetPasswordReader(pr PasswordReader) {
	c.SetPasswordReader(pr)
}

// SetCobraCmdOut set the given *cobra.Command stdout to the same stdout currently used.
func (c *Cli) SetCobraCmdOut(cmd *cobra.Command) {
	cmd.SetOut(c.stdout)
}

// SetCobraCmdOut set the given *cobra.Command stdout to the same stdout currently used.
func SetCobraCmdOut(cmd *cobra.Command) {
	c.SetCobraCmdOut(cmd)
}

// SetCobraCmdErr set the given *cobra.Command stderr to the same stderr currently used.
func (c *Cli) SetCobraCmdErr(cmd *cobra.Command) {
	cmd.SetOut(c.stderr)
}

// SetCobraCmdErr set the given *cobra.Command stderr to the same stderr currently used.
func SetCobraCmdErr(cmd *cobra.Command) {
	c.SetCobraCmdErr(cmd)
}

// SetCobraCmdOutErr set the given *cobra.Command stdout and stderr to the same stdout and stderr currently used.
func (c *Cli) SetCobraCmdOutErr(cmd *cobra.Command) {
	c.SetCobraCmdOut(cmd)
	c.SetCobraCmdErr(cmd)
}

// SetCobraCmdOutErr set the given *cobra.Command stdout and stderr to the same stdout and stderr currently used.
func SetCobraCmdOutErr(cmd *cobra.Command) {
	c.SetCobraCmdOutErr(cmd)
}

// SetSkipConfirmation enable or disable the confirmation prompt of some commands.
func (c *Cli) SetSkipConfirmation(b bool) {
	c.skipConfirmation = b
}

// SetSkipConfirmation enable or disable the confirmation prompt of some commands.
func SetSkipConfirmation(b bool) {
	c.SetSkipConfirmation(b)
}

// GetSkipConfirmation returns if the confirmation prompt of some commands is enabled or not.
func (c *Cli) GetSkipConfirmation() bool {
	return c.skipConfirmation
}

// GetSkipConfirmation returns if the confirmation prompt of some commands is enabled or not.
func GetSkipConfirmation() bool {
	return c.GetSkipConfirmation()
}
