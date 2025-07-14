package terraform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/santi1s/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

type e2eTest struct {
	initial        []testFile
	args           []string
	expectedStdout string
	expectedStderr string
	expected       interface{}
}

type testFile struct {
	path    string
	content string
}

type githubAnnotation struct {
	level   string
	path    string
	line    int
	message string
}

type expectedError struct {
	name     string
	error    error
	contains string
}

func initE2ETest() {
	helper.InitViper("")

	providedFlags = terraformFlags{}
	moduleCheckFlags = ModuleCheckFlags{}
}

func TestGetTerraformFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestGetTerraformFiles")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testFiles := []testFile{
		{path: dir + "/foo.tf"},
		{path: dir + "/other_dir/bar.tf"},
		{path: dir + "/.terraform/baz.tf"},
		{path: dir + "terraform/ci/foo.tf"},
		{path: dir + "/.git/baz.tf"},
		{path: dir + "/baz.txt"},
	}

	expected := []string{
		dir + "/foo.tf",
		dir + "/other_dir/bar.tf",
	}

	for _, tfFile := range testFiles {
		dir := filepath.Dir(tfFile.path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = helper.CreateDirectory(dir)
			if err != nil {
				t.Fatal(err)
			}
		}
		file, err := os.Create(tfFile.path)
		if err != nil {
			t.Fatal(err)
		}
		file.Close()
	}

	files, err := getTerraformFiles(dir)
	assert.NoError(t, err)
	assert.Equal(t, files, expected, "content should be the same")
}

func TestGetTerraformDirs(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestGetTerraformDirs")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	testFiles := []testFile{
		{path: dir + "/foo.tf"},
		{path: dir + "/.terraform/bar.tf"},
		{path: dir + "/module1/main.tf"},
		{path: dir + "/module2/variables.tf"},
		{path: dir + "/terraform/ci/ignored.tf"},
	}

	expected := map[string]bool{
		dir:              true,
		dir + "/module1": true,
		dir + "/module2": true,
	}

	for _, tfFile := range testFiles {
		dir := filepath.Dir(tfFile.path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = helper.CreateDirectory(dir)
			if err != nil {
				t.Fatal(err)
			}
		}
		file, err := os.Create(tfFile.path)
		if err != nil {
			t.Fatal(err)
		}
		file.Close()
	}

	dirs, err := GetTerraformDirs(dir)
	assert.NoError(t, err)
	assert.Equal(t, expected, dirs, "directories should be the same")
}
