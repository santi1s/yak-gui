package cli_test

import (
	"io"
	"strings"
	"testing"

	"github.com/doctolib/yak/cli"
	"github.com/stretchr/testify/assert"
)

// TestReadPassword is a unit test for ReadPassword function.
func TestReadPassword(t *testing.T) {
	reader := strings.NewReader("password_value")
	cli.SetPasswordReader(&cli.IoReaderPasswordReader{Reader: reader})
	password, err := cli.ReadPassword()

	assert.NoError(t, err, "should be nil")
	assert.Equal(t, "password_value", password, "should be equal")
}

// TestRead is a unit test for Read function.
func TestRead(t *testing.T) {
	reader := strings.NewReader("value")
	cli.SetIn(reader)
	value, err := cli.Read()

	assert.NoError(t, err, "should be nil")
	assert.Equal(t, "value", value, "should be equal")
}

// TestReadLine is a unit test for Read function.
func TestReadLine(t *testing.T) {
	// test with a string containing spaces
	tests := []string{
		"foo bar baz",
		"value",
		"with a space at the end ",
	}

	for _, expected := range tests {
		reader := strings.NewReader(expected)
		cli.SetIn(reader)
		value, err := cli.ReadLine()

		assert.NoError(t, err, "should be nil")
		assert.Equal(t, expected, value, "should be equal")
	}
}

type AskConfirmationTest struct {
	Input       string
	SkipConfirm bool
	Expected    bool
}

// TestAskConfirmation is a unit test for AskConfirmation function.
func TestAskConfirmation(t *testing.T) {
	var testScenariosAskConfirmation = map[string]AskConfirmationTest{
		"skip confirm": {
			Input:       "",
			SkipConfirm: true,
			Expected:    true,
		},
		"input y": {
			Input:       "y",
			SkipConfirm: false,
			Expected:    true,
		},
		"input Y": {
			Input:       "Y",
			SkipConfirm: false,
			Expected:    true,
		},
		"input yes": {
			Input:       "yes",
			SkipConfirm: false,
			Expected:    true,
		},
		"input YES": {
			Input:       "YES",
			SkipConfirm: false,
			Expected:    true,
		},
		"input n": {
			Input:       "n",
			SkipConfirm: false,
			Expected:    false,
		},
		"input N": {
			Input:       "N",
			SkipConfirm: false,
			Expected:    false,
		},
		"input no": {
			Input:       "no",
			SkipConfirm: false,
			Expected:    false,
		},
		"input NO": {
			Input:       "NO",
			SkipConfirm: false,
			Expected:    false,
		},
		"invalid input then yes": {
			Input:       "test\nyes",
			SkipConfirm: false,
			Expected:    true,
		},
		"invalid input then no": {
			Input:       "test\nno",
			SkipConfirm: false,
			Expected:    false,
		},
	}

	for k, v := range testScenariosAskConfirmation {
		t.Run(k, func(t *testing.T) {
			reader := strings.NewReader(v.Input)
			cli.SetIn(reader)
			cli.SetOut(io.Discard)
			cli.SetSkipConfirmation(v.SkipConfirm)
			result := cli.AskConfirmation("test?")
			assert.Equal(t, v.Expected, result, "should be equal")
		})
	}
}
