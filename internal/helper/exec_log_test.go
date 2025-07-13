package helper

import (
	"os/exec"
	"testing"
)

func TestExecLogger(t *testing.T) {
	cmd := exec.Command("/bin/sh", "-c", "ls /ee; echo foobar")
	buf, err := ExecStreamLogger(cmd, false)
	if err != nil {
		t.Fatal(err)
	}
	expected := "foobar\n"
	output := buf.String()
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

func TestExecLoggerFailure(t *testing.T) {
	cmd := exec.Command("blahblah")
	_, err := ExecStreamLogger(cmd, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
