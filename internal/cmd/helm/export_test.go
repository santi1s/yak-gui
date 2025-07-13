package helm

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestFullExport(t *testing.T) {
	workDir := func() string {
		path, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		return path
	}()
	config, err := NewExportConfig("testdata/helm-export.yaml")
	assert.NoError(t, err)

	exportDir, err := os.MkdirTemp("", "yak_helm_export_test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(exportDir)
	errors, err := config.Export(workDir, exportDir)
	assert.NoError(t, err)
	if errors != 0 {
		t.Fatalf("expected 0 errors, got %d", errors)
	}
	manifest, err := os.ReadFile(filepath.Join(exportDir, "envs", "testdata", "helm", "myapp", "export_myapp1.yml"))
	assert.NoError(t, err)
	output := map[string]string{}
	err = yaml.Unmarshal(manifest, &output)
	assert.NoError(t, err)

	expected := map[string]string{
		"a": "foo",
		"b": "monday",
		"c": "wednesday",
	}

	if !reflect.DeepEqual(output, expected) {
		assert.Equal(t, expected, output)
	}
}
