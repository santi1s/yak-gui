package helm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santi1s/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

var whitelistFile = `external_helm_charts:
  - "https://foo.bar/cloudprem"
  - "oci://1234.dkr.ecr.eu-central-1.amazonaws.com/foo/cloudprem"
`

var registriesFile = `ecr:
- region: eu-central-1
  account: 580698825394
  profile: tooling
- region: eu-west-3
  account: 580698825394
  profile: tooling
`

type helmCheckE2eTest struct {
	TestFiles     []*testFile
	Args          []string
	ExpectedError bool
}

type testFile struct {
	Path    string
	Content string
}

func TestE2eHelmCheck(t *testing.T) {
	var testScenarios = map[string]helmCheckE2eTest{
		"argocd_internal_chart": {
			TestFiles: []*testFile{
				{
					Path: "configs/helm/dev/dev-aws-de-fra-1/helm-export.yaml",
					Content: `applications:
  teleport-agent:
    sources:
      - exportName: teleport-agent
        repoURL: oci://580698825394.dkr.ecr.eu-central-1.amazonaws.com/helm-teleport-agent
        targetRevision: 0.1.202408231203
        helm:
          valueFiles:
            - configs/helm/dev/dev-aws-de-fra-1/teleport-agent/values.yaml
    destination:
      namespace: teleport-agent
`},
			},
			Args:          []string{"check", "--file", "configs/helm/whitelist.yml"},
			ExpectedError: false,
		},
		"argocd_internal_chart_not_semver": {
			TestFiles: []*testFile{
				{
					Path: "configs/helm/dev/dev-aws-de-fra-1/helm-export.yaml",
					Content: `applications:
  teleport-agent:
    sources:
      - exportName: teleport-agent
        repoURL: oci://580698825394.dkr.ecr.eu-central-1.amazonaws.com/helm-teleport-agent
        targetRevision: latest
        helm:
          valueFiles:
            - configs/helm/dev/dev-aws-de-fra-1/teleport-agent/values.yaml
    destination:
      namespace: teleport-agent
`},
			},
			Args:          []string{"check", "--file", "configs/helm/whitelist.yml"},
			ExpectedError: true,
		},
		"argocd_internal_chart_prerelease": {
			TestFiles: []*testFile{
				{
					Path: "configs/helm/dev/dev-aws-de-fra-1/helm-export.yaml",
					Content: `applications:
  teleport-agent:
    sources:
      - exportName: teleport-agent
        repoURL: oci://580698825394.dkr.ecr.eu-central-1.amazonaws.com/helm-teleport-agent
        targetRevision: 0.0.0-pr1
        helm:
          valueFiles:
            - configs/helm/dev/dev-aws-de-fra-1/teleport-agent/values.yaml
    destination:
      namespace: teleport-agent
`},
			},
			Args:          []string{"check", "--file", "configs/helm/whitelist.yml"},
			ExpectedError: true,
		},
		"argocd_external_chart_not_whitelisted": {
			TestFiles: []*testFile{
				{
					Path: "configs/helm/dev/dev-aws-de-fra-1/helm-export.yaml",
					Content: `applications:
  teleport-agent:
    sources:
      - exportName: teleport-agent
        repoURL: oci://foo.bar.baz
        targetRevision: 0.1.202408231203
        helm:
          valueFiles:
            - configs/helm/dev/dev-aws-de-fra-1/teleport-agent/values.yaml
    destination:
      namespace: teleport-agent
`},
			},
			Args:          []string{"check", "--file", "configs/helm/whitelist.yml"},
			ExpectedError: true,
		},
		"argocd_external_chart_oci_whitelisted": {
			TestFiles: []*testFile{
				{
					Path: "configs/helm/dev/dev-aws-de-fra-1/helm-export.yaml",
					Content: `applications:
  teleport-agent:
    sources:
      - exportName: teleport-agent
        repoURL: oci://1234.dkr.ecr.eu-central-1.amazonaws.com/foo/cloudprem
        targetRevision: 0.1.202408231203
        helm:
          valueFiles:
            - configs/helm/dev/dev-aws-de-fra-1/teleport-agent/values.yaml
    destination:
      namespace: teleport-agent
`},
			},
			Args:          []string{"check", "--file", "configs/helm/whitelist.yml"},
			ExpectedError: false,
		},
		"argocd_external_chart_https_whitelisted": {
			TestFiles: []*testFile{
				{
					Path: "configs/helm/dev/dev-aws-de-fra-1/helm-export.yaml",
					Content: `applications:
  teleport-agent:
    sources:
      - exportName: teleport-agent
        repoURL: https://foo.bar/cloudprem
        targetRevision: 0.1.202408231203
        helm:
          valueFiles:
            - configs/helm/dev/dev-aws-de-fra-1/teleport-agent/values.yaml
    destination:
      namespace: teleport-agent
`},
			},
			Args:          []string{"check", "--file", "configs/helm/whitelist.yml"},
			ExpectedError: false,
		},
		"chart_oci_dependency_whitelisted": {
			TestFiles: []*testFile{
				{
					Path: "Chart.yaml",
					Content: `apiVersion: v2
name: formance
description: A Helm chart to deploy Formance stack

type: application

version: 0.1.0

appVersion: "2.0.0-rc.35"

dependencies:
- condition: cloudprem.enabled
  name: cloudprem
  repository: oci://1234.dkr.ecr.eu-central-1.amazonaws.com/foo
  version: v1.0.0
  `},
			},
			Args:          []string{"check", "--file", "configs/helm/whitelist.yml", "--chart"},
			ExpectedError: false,
		},
		"chart_https_dependency_whitelisted": {
			TestFiles: []*testFile{
				{
					Path: "Chart.yaml",
					Content: `apiVersion: v2
name: formance
description: A Helm chart to deploy Formance stack

type: application

version: 0.1.0

appVersion: "2.0.0-rc.35"

dependencies:
- condition: cloudprem.enabled
  name: cloudprem
  repository: https://foo.bar
  version: v1.0.0
  `},
			},
			Args:          []string{"check", "--file", "configs/helm/whitelist.yml", "--chart"},
			ExpectedError: false,
		},
		"chart_internal_chart_dependency": {
			TestFiles: []*testFile{
				{
					Path: "Chart.yaml",
					Content: `apiVersion: v2
name: formance
description: A Helm chart to deploy Formance stack

type: application

version: 0.1.0

appVersion: "2.0.0-rc.35"

dependencies:
- condition: cloudprem.enabled
  name: cloudprem
  repository: oci://580698825394.dkr.ecr.eu-central-1.amazonaws.com/foo
  version: v1.0.0
  `},
			},
			Args:          []string{"check", "--file", "configs/helm/whitelist.yml", "--chart"},
			ExpectedError: false,
		},
		"chart_internal_chart_dependency_prerelease": {
			TestFiles: []*testFile{
				{
					Path: "Chart.yaml",
					Content: `apiVersion: v2
name: formance
description: A Helm chart to deploy Formance stack

type: application

version: 0.1.0

appVersion: "2.0.0-rc.35"

dependencies:
- condition: cloudprem.enabled
  name: cloudprem
  repository: oci://580698825394.dkr.ecr.eu-central-1.amazonaws.com/foo
  version: v1.0.0-1
  `},
			},
			Args:          []string{"check", "--file", "configs/helm/whitelist.yml", "--chart"},
			ExpectedError: true,
		},
		"chart_not_whitelisted_dependency": {
			TestFiles: []*testFile{
				{
					Path: "Chart.yaml",
					Content: `apiVersion: v2
name: formance
description: A Helm chart to deploy Formance stack

type: application

version: 0.1.0

appVersion: "2.0.0-rc.35"

dependencies:
- condition: cloudprem.enabled
  name: cloudprem
  repository: oci://bar.baz
  version: v1.0.0
  `},
			},
			Args:          []string{"check", "--file", "configs/helm/whitelist.yml", "--chart"},
			ExpectedError: true,
		},
		"chart_no_dependency": {
			TestFiles: []*testFile{
				{
					Path: "Chart.yaml",
					Content: `apiVersion: v2
name: formance
description: A Helm chart to deploy Formance stack

type: application

version: 0.1.0

appVersion: "2.0.0-rc.35"
  `},
			},
			Args:          []string{"check", "--file", "configs/helm/whitelist.yml", "--chart"},
			ExpectedError: false,
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "TestE2eHelmCheck")
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println(dir)
			defer os.RemoveAll(dir)

			err = os.Chdir(dir)
			if err != nil {
				t.Fatal(err)
			}
			cwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			err = os.Setenv("YAK_HELM_WORKDIR", cwd)
			if err != nil {
				t.Fatal(err)
			}
			err = helper.CreateDirectory(filepath.Join(dir, "/configs/helm"))
			if err != nil {
				t.Fatal(err)
			}

			file, err := os.Create(filepath.Join(dir, "/configs/helm/whitelist.yml"))
			if err != nil {
				t.Fatal(err)
			}

			_, err = file.Write([]byte(whitelistFile))
			if err != nil {
				t.Fatal(err)
			}
			file.Close()

			file, err = os.Create(filepath.Join(dir, "/configs/helm/registries.yaml"))
			if err != nil {
				t.Fatal(err)
			}

			_, err = file.Write([]byte(registriesFile))
			if err != nil {
				t.Fatal(err)
			}
			file.Close()

			if v.TestFiles != nil {
				for _, ts := range v.TestFiles {
					err = helper.CreateDirectory(filepath.Join(dir, filepath.Dir(ts.Path)))
					if err != nil {
						t.Fatal(err)
					}
					file, err := os.Create(filepath.Join(dir, ts.Path))
					if err != nil {
						t.Fatal(err)
					}
					_, err = file.Write([]byte(ts.Content))
					if err != nil {
						t.Fatal(err)
					}
					file.Close()
				}
			}

			providedHelmFlags = HelmFlags{}
			_, _, err = helper.ExecuteCobraCommand(helmCmd, v.Args, strings.NewReader(""))
			if v.ExpectedError {
				assert.ErrorIs(t, err, errHelmChartsProblemFound, "err should be errHelmChartsProblemFound")
			} else {
				assert.NoError(t, err, "err should be nil")
			}
		})
	}
}

func TestIsInternalOciRepositoryAllowed(t *testing.T) {
	registries := &HelmRegistries{
		ECR: []*ECRRegistry{
			{
				Account: "580698825394",
				Region:  "eu-central-1",
				Profile: "tooling",
			},
		},
	}
	testScenarios := map[string]interface{}{"allowed": map[string]interface{}{"expected": true, "url": "oci://580698825394.dkr.ecr.eu-central-1.amazonaws.com/foo"},
		"forbidden": map[string]interface{}{"expected": false, "url": "oci://1234567890.dkr.ecr.eu-central-1.amazonaws.com/foo"}}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			result := IsInternalOciRepository(v.(map[string]interface{})["url"].(string), registries)
			assert.Equal(t, v.(map[string]interface{})["expected"].(bool), result)
		})
	}
}

func TestIsExternalChartAllowed(t *testing.T) {
	whitelist := []string{"https://foo.bar/baz", "oci://580698825394.dkr.ecr.eu-central-1.amazonaws.com/foo", "https://foo.test/a"}
	testScenarios := map[string]interface{}{"allowed": map[string]interface{}{"expected": true, "url": "oci://580698825394.dkr.ecr.eu-central-1.amazonaws.com/foo"},
		"forbidden": map[string]interface{}{"expected": false, "url": "https://ba.bar/baz"}}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			result := IsExternalChartAllowed(v.(map[string]interface{})["url"].(string), whitelist)
			assert.Equal(t, v.(map[string]interface{})["expected"].(bool), result)
		})
	}
}
