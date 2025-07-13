package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsHelmOCI(t *testing.T) {
	tests := map[string]struct {
		app      *ApplicationSource
		expected bool
	}{
		"https": {
			app: &ApplicationSource{
				RepoURL: "https://test",
			},
			expected: false,
		},
		"local": {
			app: &ApplicationSource{
				Chart: ".",
				Path:  "foo/bar",
			},
			expected: false,
		},
		"oci": {
			app: &ApplicationSource{
				RepoURL: "oci://test",
			},
			expected: true,
		},
	}

	for k, v := range tests {
		t.Run(k, func(t *testing.T) {
			output := v.app.IsHelmOCI()
			assert.Equal(t, v.expected, output)
		})
	}
}

func TestIsHelmRepository(t *testing.T) {
	tests := map[string]struct {
		app      *ApplicationSource
		expected bool
	}{
		"https": {
			app: &ApplicationSource{
				RepoURL: "https://test",
			},
			expected: true,
		},
		"local": {
			app: &ApplicationSource{
				Chart: ".",
				Path:  "foo/bar",
			},
			expected: false,
		},
		"oci": {
			app: &ApplicationSource{
				RepoURL: "oci://test",
			},
			expected: false,
		},
	}

	for k, v := range tests {
		t.Run(k, func(t *testing.T) {
			output := v.app.IsHelmRepository()
			assert.Equal(t, v.expected, output)
		})
	}
}

func TestIsLocalChart(t *testing.T) {
	tests := map[string]struct {
		app      *ApplicationSource
		expected bool
	}{
		"https": {
			app: &ApplicationSource{
				RepoURL: "https://test",
			},
			expected: false,
		},
		"local": {
			app: &ApplicationSource{
				Chart: ".",
				Path:  "foo/bar",
			},
			expected: true,
		},
		"oci": {
			app: &ApplicationSource{
				RepoURL: "oci://test",
			},
			expected: false,
		},
	}

	for k, v := range tests {
		t.Run(k, func(t *testing.T) {
			output := v.app.IsLocalChart()
			assert.Equal(t, v.expected, output)
		})
	}
}

func TestIsExternalChart(t *testing.T) {
	registries := &HelmRegistries{
		ECR: []*ECRRegistry{
			{
				Account: "580698825394",
				Region:  "eu-central-1",
				Profile: "tooling",
			},
		},
	}
	tests := map[string]struct {
		app      *ApplicationSource
		expected bool
	}{
		"https": {
			app: &ApplicationSource{
				RepoURL: "https://test",
			},
			expected: true,
		},
		"http": {
			app: &ApplicationSource{
				RepoURL: "http://test",
			},
			expected: true,
		},
		"local": {
			app: &ApplicationSource{
				Chart: ".",
				Path:  "foo/bar",
			},
			expected: false,
		},
		"oci_internal": {
			app: &ApplicationSource{
				RepoURL: "oci://580698825394.dkr.ecr.eu-central-1.amazonaws.com/test",
			},
			expected: false,
		},
		"oci_external": {
			app: &ApplicationSource{
				RepoURL: "oci://foo.bar.baz",
			},
			expected: true,
		},
	}

	for k, v := range tests {
		t.Run(k, func(t *testing.T) {
			output := v.app.IsExternalChart(registries)
			assert.Equal(t, v.expected, output)
		})
	}
}

func TestExtractEnvFromHelmExportConfigFile(t *testing.T) {
	apps := &ApplicationsValues{
		Filepath: "/kube/configs/helm/preprod/preprod/helm-export.yaml",
	}
	output := apps.ExtractEnv()
	expected := "preprod"
	assert.Equal(t, expected, output)
}

func TestValidateApplicationSourceErrors(t *testing.T) {
	tests := map[string]struct {
		app *ApplicationSource
	}{
		"no_exportName": {
			app: &ApplicationSource{},
		},
		"no_chart": {
			app: &ApplicationSource{
				ExportName: "test",
			},
		},
		"oci_repo": {
			app: &ApplicationSource{
				ExportName: "test",
				RepoURL:    "oci://test",
				Chart:      "test",
			},
		},
		"chart_path": {
			app: &ApplicationSource{
				ExportName: "test",
				Chart:      ".",
			},
		},
	}

	for k, v := range tests {
		t.Run(k, func(t *testing.T) {
			err := v.app.Validate()
			assert.Error(t, err)
		})
	}
}
