package terraform

import (
	"os"
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestE2eProviderGenerate(t *testing.T) {
	var testScenarios = map[string]e2eTest{
		"nominal": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = ">= 0.14.11, < 2.0.0"

  required_providers {
    aws = {
      source  = "any/source"
      version = ">= 4.0.0, < 5.0.0"
    }
    tls = {
      source  = "any/source"
      version = ">= 3.0.0, < 4.0.0"
    }
  }
}`,
				},
			},
			args: []string{"provider", "generate"},
			expected: []testFile{
				{
					path: "provider_aws.tf",
					content: `provider "aws" {}
`,
				},
				{
					path: "provider_tls.tf",
					content: `provider "tls" {}
`,
				},
			},
		},
		"nominal_aliases": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = ">= 0.14.11, < 2.0.0"

  required_providers {
    aws = {
      source  = "any/source"
      version = ">= 4.0.0, < 5.0.0"
	  configuration_aliases = [ aws.alias1, aws.alias2 ]
    }
  }
}`,
				},
			},
			args: []string{"provider", "generate"},
			expected: []testFile{
				{
					path: "provider_aws.tf",
					content: `provider "aws" {
  alias = "alias1"
}
provider "aws" {
  alias = "alias2"
}
`,
				},
			},
		},
	}

	dir, err := os.MkdirTemp("", "TestTerraformProviderCheck")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			initE2ETest()

			err = helper.CreateDirectory(dir + "/" + k)
			if err != nil {
				t.Fatal(err)
			}

			err := os.Chdir(dir + "/" + k)
			if err != nil {
				t.Fatal(err)
			}

			for _, initial := range v.initial {
				err = helper.WriteFile(initial.content, dir+"/"+k+"/"+initial.path)
				if err != nil {
					t.Fatal(t)
				}
			}

			_, stderr, err := helper.ExecuteCobraCommand(terraformCmd, v.args)
			if v.expected == nil {
				assert.NoError(t, err)
				assert.Empty(t, stderr, "stderr should be empty")
			} else if expectedFiles, ok := v.expected.([]testFile); ok {
				for _, expectedFile := range expectedFiles {
					assert.NoError(t, err, "error should be nil")
					content, err := os.ReadFile(dir + "/" + k + "/" + expectedFile.path)
					if err != nil {
						t.Fatal(err)
					}
					assert.Equal(t, expectedFile.content, string(content), "should be equal")
				}
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}
