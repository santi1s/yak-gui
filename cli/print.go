package cli

import (
	"encoding/json"
	"fmt"
	"time"

	logrus "github.com/sirupsen/logrus"

	"sigs.k8s.io/yaml"
)

// Printf formats according to a format specifier and writes to the current stdout.
// It returns the number of bytes written and any write error encountered.
func (c *Cli) Printf(s string, a ...any) (int, error) {
	return fmt.Fprintf(c.stdout, s, a...)
}

// Printf formats according to a format specifier and writes to the current stdout.
// It returns the number of bytes written and any write error encountered.
func Printf(s string, a ...any) (int, error) {
	return c.Printf(s, a...)
}

// Println formats using the default formats for its operands and writes to the current stdout.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func (c *Cli) Println(a ...any) (int, error) {
	return fmt.Fprintln(c.stdout, a...)
}

// Println formats using the default formats for its operands and writes to the current stdout.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func Println(a ...any) (int, error) {
	return c.Println(a...)
}

func LoglnAndSleep(seconds int, a ...any) {
	logrus.Println(a...)
	time.Sleep(time.Duration(seconds) * time.Second)
}

// Print formats using the default formats for its operands and writes to the current stdout.
// Spaces are added between operands when neither is a string.
// It returns the number of bytes written and any write error encountered.
func (c *Cli) Print(a ...any) (int, error) {
	return fmt.Fprint(c.stdout, a...)
}

// Print formats using the default formats for its operands and writes to the current stdout.
// Spaces are added between operands when neither is a string.
// It returns the number of bytes written and any write error encountered.
func Print(a ...any) (int, error) {
	return c.Print(a...)
}

// PrintfErr formats according to a format specifier and writes to the current stderr.
// It returns the number of bytes written and any write error encountered.
func (c *Cli) PrintfErr(s string, a ...any) (int, error) {
	return fmt.Fprintf(c.stderr, s, a...)
}

// PrintfErr formats according to a format specifier and writes to the current stderr.
// It returns the number of bytes written and any write error encountered.
func PrintfErr(s string, a ...any) (int, error) {
	return c.PrintfErr(s, a...)
}

// PrintlnErr formats using the default formats for its operands and writes to the current stderr.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func (c *Cli) PrintlnErr(a ...any) (int, error) {
	return fmt.Fprintln(c.stderr, a...)
}

// PrintlnErr formats using the default formats for its operands and writes to the current stderr.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func PrintlnErr(a ...any) (int, error) {
	return c.PrintlnErr(a...)
}

// PrintErr formats using the default formats for its operands and writes to the current stderr.
// Spaces are added between operands when neither is a string.
// It returns the number of bytes written and any write error encountered.
func (c *Cli) PrintErr(a ...any) (int, error) {
	return fmt.Fprint(c.stderr, a...)
}

// PrintErr formats using the default formats for its operands and writes to the current stderr.
// Spaces are added between operands when neither is a string.
// It returns the number of bytes written and any write error encountered.
func PrintErr(a ...any) (int, error) {
	return c.PrintErr(a...)
}

// PrintJSON marshals an interface to JSON and writes it to the current stdout.
// It returns any error encountered.
func (c *Cli) PrintJSON(v interface{}) error {
	str, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	c.Println(string(str))
	return nil
}

// PrintJSON marshals an interface to JSON and writes it to the current stdout.
// It returns any error encountered.
func PrintJSON(v interface{}) error {
	return c.PrintJSON(v)
}

// PrintYAML marshals an interface to YAML and writes it to the current stdout.
// It returns any error encountered.
func (c *Cli) PrintYAML(v interface{}) error {
	str, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	c.Println("---")
	c.Println(string(str))
	return nil
}

// PrintYAML marshals an interface to YAML and writes it to the current stdout.
// It returns any error encountered.
func PrintYAML(v interface{}) error {
	return c.PrintYAML(v)
}

// SprintJSON marshals an interface to JSON and returns it.
// Returns an empty string if an error is encountered.
func (c *Cli) SprintJSON(v interface{}) string {
	str, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		return string(str)
	}

	return ""
}

// SprintJSON marshals an interface to JSON and returns it.
// Returns an empty string if an error is encountered.
func SprintJSON(v interface{}) string {
	return c.SprintJSON(v)
}

// SprintYAML marshals an interface to YAML and returns it.
// Returns an empty string if an error is encountered.
func (c *Cli) SprintYAML(v interface{}) string {
	str, err := yaml.Marshal(v)
	if err == nil {
		return "---\n" + string(str)
	}

	return ""
}

// SprintYAML marshals an interface to YAML and returns it.
// Returns an empty string if an error is encountered.
func SprintYAML(v interface{}) string {
	return c.SprintYAML(v)
}
