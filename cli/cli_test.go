// Package cli_test provides the needed unit tests for cli package.
package cli_test

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/santi1s/yak/cli"
	"github.com/stretchr/testify/assert"
)

// TestOut is a unit test for GetOut and SetOut functions.
func TestOut(t *testing.T) {
	out := cli.GetOut()
	assert.Equal(t, os.Stdout, out, "should be equal")
	cli.SetOut(io.Discard)
	out = cli.GetOut()

	assert.Equal(t, io.Discard, out, "should be equal")
}

// TestErr is a unit test for GetErr and SetErr functions.
func TestErr(t *testing.T) {
	out := cli.GetErr()
	assert.Equal(t, os.Stderr, out, "should be equal")
	cli.SetErr(io.Discard)
	out = cli.GetErr()

	assert.Equal(t, io.Discard, out, "should be equal")
}

// TestIn is a unit test for GetIn and SetIn functions.
func TestIn(t *testing.T) {
	out := cli.GetIn()
	assert.Equal(t, os.Stdin, out, "should be equal")
	cli.SetIn(strings.NewReader(""))
	out = cli.GetIn()

	assert.Equal(t, strings.NewReader(""), out, "should be equal")
}

// TestSkipConfirmation is a unit test for SetSkipConfirmation and GetSkipConfirmation functions.
func TestSkipConfirmation(t *testing.T) {
	cli.SetSkipConfirmation(true)
	out := cli.GetSkipConfirmation()
	assert.Equal(t, true, out, "should be equal")

	cli.SetSkipConfirmation(false)
	out = cli.GetSkipConfirmation()
	assert.Equal(t, false, out, "should be equal")
}
