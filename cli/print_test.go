package cli_test

import (
	"bytes"
	"testing"

	"github.com/santi1s/yak/cli"
	"github.com/stretchr/testify/assert"
)

// PrintTest represent the test scenario for Print* functions.
type PrintTest struct {
	wanted   string
	function interface{}
}

// initOutput is a helper function to initialize stdin and stderr in a bytes.Buffer
// It returns the buffer, in order, for stdout and stderr.
func initOutput() (*bytes.Buffer, *bytes.Buffer) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cli.SetOut(stdout)
	cli.SetErr(stderr)

	return stdout, stderr
}

// TestPrintln is a unit test for Println function.
func TestPrintln(t *testing.T) {
	stdout, stderr := initOutput()

	n, err := cli.Println("test")
	assert.Equal(t, "test\n", stdout.String(), "should be equal")
	assert.Empty(t, stderr, "should be empty")
	assert.NoError(t, err, "err should be nil")
	assert.Equal(t, 5, n, "should be equal")
}

// TestPrint is a unit test for Print function.
func TestPrint(t *testing.T) {
	stdout, stderr := initOutput()

	n, err := cli.Print("test")
	assert.Equal(t, "test", stdout.String(), "should be equal")
	assert.Empty(t, stderr, "should be empty")
	assert.NoError(t, err, "err should be nil")
	assert.Equal(t, 4, n, "should be equal")
}

// TestPrintf is a unit test for Printf function.
func TestPrintf(t *testing.T) {
	stdout, stderr := initOutput()

	n, err := cli.Printf("test")
	assert.Equal(t, "test", stdout.String(), "should be equal")
	assert.Empty(t, stderr, "should be empty")
	assert.NoError(t, err, "err should be nil")
	assert.Equal(t, 4, n, "should be equal")
}

// TestPrintlnErr is a unit test for PrintlnErr function.
func TestPrintlnErr(t *testing.T) {
	stdout, stderr := initOutput()

	n, err := cli.PrintlnErr("test")
	assert.Equal(t, "test\n", stderr.String(), "should be equal")
	assert.Empty(t, stdout, "should be empty")
	assert.NoError(t, err, "err should be nil")
	assert.Equal(t, 5, n, "should be equal")
}

// TestPrintErr is a unit test for PrintErr function.
func TestPrintErr(t *testing.T) {
	stdout, stderr := initOutput()

	n, err := cli.PrintErr("test")
	assert.Equal(t, "test", stderr.String(), "should be equal")
	assert.Empty(t, stdout, "should be empty")
	assert.NoError(t, err, "err should be nil")
	assert.Equal(t, 4, n, "should be equal")
}

// TestPrintfErr is a unit test for PrintfErr function.
func TestPrintfErr(t *testing.T) {
	stdout, stderr := initOutput()

	n, err := cli.PrintfErr("test")
	assert.Equal(t, "test", stderr.String(), "should be equal")
	assert.Empty(t, stdout, "should be empty")
	assert.NoError(t, err, "err should be nil")
	assert.Equal(t, 4, n, "should be equal")
}

// TestPrintJSON is a unit test for PrintJSON function.
func TestPrintJSON(t *testing.T) {
	stdout, stderr := initOutput()
	wanted := `{
  "baz": 1,
  "foo": "bar"
}`
	err := cli.PrintJSON(map[string]interface{}{"foo": "bar", "baz": 1})
	assert.NoError(t, err, "should be nil")
	assert.Empty(t, stderr, "should be empty")
	assert.Contains(t, stdout.String(), wanted, "should be equal")
}

// TestPrintYAML is a unit test for PrintYAML function.
func TestPrintYAML(t *testing.T) {
	stdout, stderr := initOutput()
	wanted := `---
baz: 1
foo: bar`
	err := cli.PrintYAML(map[string]interface{}{"foo": "bar", "baz": 1})
	assert.NoError(t, err, "should be nil")
	assert.Empty(t, stderr, "should be empty")
	assert.Contains(t, stdout.String(), wanted, "should be equal")
}

// TestSprintJSON is a unit test for SprintJSON function.
func TestSprintJSON(t *testing.T) {
	wanted := `{
  "baz": 1,
  "foo": "bar"
}`
	result := cli.SprintJSON(map[string]interface{}{"foo": "bar", "baz": 1})
	assert.Contains(t, result, wanted, "should be equal")
}

// TestSprintYAML is a unit test for SprintYAML function.
func TestSprintYAML(t *testing.T) {
	wanted := `---
baz: 1
foo: bar`
	result := cli.SprintYAML(map[string]interface{}{"foo": "bar", "baz": 1})
	assert.Contains(t, result, wanted, "should be equal")
}
