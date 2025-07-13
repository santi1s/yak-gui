package helm

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type HelmPostRenderer struct {
	ExecPath string   `yaml:"execPath"`
	Args     []string `yaml:"args"`
}
type ApplicationSource struct {
	Path           string                 `yaml:"path"`
	Chart          string                 `yaml:"chart"`
	RepoURL        string                 `yaml:"repoURL"`
	TargetRevision string                 `yaml:"targetRevision"`
	Helm           *ApplicationSourceHelm `yaml:"helm"`
	ExportName     string                 `yaml:"exportName"`
}
type ApplicationSourceHelm struct {
	ValueFiles    []string               `yaml:"valueFiles"`
	ValuesObject  map[string]interface{} `yaml:"valuesObject"`
	SkipCrds      bool                   `yaml:"skipCrds"`
	IncludeCrds   bool                   `yaml:"includeCrds"`
	DontSkipTests bool                   `yaml:"dontSkipTests"`
	PostRenderer  *HelmPostRenderer      `yaml:"postRenderer"`
}

type ApplicationValues struct {
	Destination struct {
		Namespace string `yaml:"namespace"`
	} `yaml:"destination"`
	Sources []*ApplicationSource `yaml:"sources"`
}

type ApplicationsValues struct {
	Applications map[string]*ApplicationValues `yaml:"applications"`
	Filepath     string
	ExportDir    string
}

func (a *ApplicationSource) Validate() error {
	if a.ExportName == "" {
		return fmt.Errorf("exportName is required")
	}
	if a.Chart == "" && !a.IsHelmOCI() {
		return fmt.Errorf("chart is required")
	}
	if a.Chart != "" && a.IsHelmOCI() {
		return fmt.Errorf("do not define chart when oci repo is used in repoURL")
	}
	if a.Chart != "." && a.Path != "" {
		return fmt.Errorf("chart name should be . if path is defined")
	}
	if a.Chart == "." && a.Path == "" {
		return fmt.Errorf("path should be defined if using local chart")
	}
	return nil
}

func (a *ApplicationSource) IsHelmOCI() bool {
	if u, err := url.Parse(a.RepoURL); err == nil {
		return u.Scheme == "oci"
	}
	return false
}

func (a *ApplicationSource) IsHelmRepository() bool {
	if u, err := url.Parse(a.RepoURL); err == nil {
		return u.Scheme == "https" || u.Scheme == "http"
	}
	return false
}

func (a *ApplicationSource) IsLocalChart() bool {
	if a.Chart == "." && a.Path != "" {
		return true
	}
	return false
}

func (a *ApplicationSource) IsExternalChart(registries *HelmRegistries) bool {
	if (!a.IsHelmOCI() && !a.IsHelmRepository()) || a.IsLocalChart() {
		return false
	}

	if a.IsHelmOCI() && IsInternalOciRepository(a.RepoURL, registries) {
		return false
	}

	return true
}

func NewExportConfig(f string) (*ApplicationsValues, error) {
	data, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	var applications *ApplicationsValues
	err = yaml.Unmarshal(data, &applications)

	if err != nil {
		return nil, err
	}
	applications.Filepath = f
	return applications, nil
}

func (applications *ApplicationsValues) ExtractEnv() string {
	return filepath.Base(filepath.Dir(applications.Filepath))
}

func (a *ApplicationValues) Export(appName string) []*ExportChart {
	var exports []*ExportChart
	for _, source := range a.Sources {
		exports = append(exports, &ExportChart{
			appName:   appName,
			namespace: a.Destination.Namespace,
			Source:    source,
		})
	}
	return exports
}
