package terraform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santi1s/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestProviderBump(t *testing.T) {
	var testScenarios = map[string]e2eTest{
		"nothing_to_do": {
			args:     []string{"provider", "bump", "-n", "hashicorp/aws", "-v", "1.2.3"},
			expected: []string{},
		},
		"invalid_semver": {
			args:     []string{"provider", "bump", "-n", "hashicorp/aws", "-v", ">= 3.0.a"},
			expected: expectedError{contains: "could not parse version constraint: improper constraint: >= 3.0.a"},
		},
		"nominal": {
			initial: []testFile{
				{
					path: "foo/bar/versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
	  # any comment
      version = "= 3.63.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "= 3.0.0"
    }
  }
}`,
				},
			},
			args:           []string{"provider", "bump", "-n", "hashicorp/aws", "-v", "= 3.64.0"},
			expectedStdout: "provider bumped in foo/bar/versions.tf",
			expected: []testFile{
				{
					path: "foo/bar/versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
	  # any comment
      version = "= 3.64.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "= 3.0.0"
    }
  }
}`,
				},
			},
		},
		"nominal_version_before_source": {
			initial: []testFile{
				{
					path: "foo/bar/versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    aws = {
      version = "= 3.63.0"
	  # any comment
      source  = "hashicorp/aws"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "= 3.0.0"
    }
  }
}`,
				},
			},
			args:           []string{"provider", "bump", "-n", "hashicorp/aws", "-v", "= 3.64.0"},
			expectedStdout: "provider bumped in foo/bar/versions.tf",
			expected: []testFile{
				{
					path: "foo/bar/versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    aws = {
      version = "= 3.64.0"
	  # any comment
      source  = "hashicorp/aws"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "= 3.0.0"
    }
  }
}`,
				},
			},
		},
	}

	testRoot, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testRoot)

	for k, v := range testScenarios {
		initE2ETest()

		err = helper.CreateDirectory(testRoot + "/" + k)
		if err != nil {
			t.Fatal(err)
		}

		err = os.Chdir(testRoot + "/" + k)
		if err != nil {
			panic(err)
		}

		if v.initial != nil {
			for _, testFile := range v.initial {
				err = helper.CreateDirectory(testRoot + "/" + k + "/" + filepath.Dir(testFile.path))
				if err != nil {
					t.Fatal(err)
				}
				err = helper.WriteFile(testFile.content, testRoot+"/"+k+"/"+testFile.path)
				if err != nil {
					t.Fatal(err)
				}
			}
		}

		t.Run(k, func(t *testing.T) {
			stdout, stderr, err := helper.ExecuteCobraCommand(terraformCmd, v.args, strings.NewReader(""))
			assert.Contains(t, stdout, v.expectedStdout, "stdout does not contain expected output")

			if expectedTestFiles, ok := v.expected.([]testFile); ok {
				assert.NoError(t, err, "error should be nil")
				assert.Empty(t, stderr, "stderr should be empty")
				for _, expectedTestFile := range expectedTestFiles {
					content, err := os.ReadFile(expectedTestFile.path)
					if err != nil {
						t.Fatal(err)
					}
					assert.Equal(t, expectedTestFile.content, string(content), "file has unexpected content")
				}
			} else if expectedError, ok := v.expected.(expectedError); ok {
				if expectedError.error != nil && expectedError.name != "" {
					// if we know which error we expect
					assert.ErrorIsf(t, err, expectedError.error, "error should be %s", expectedError.name)
				} else if expectedError.contains != "" {
					assert.ErrorContains(t, err, expectedError.contains)
				}
			} else if v.expected == nil && v.expectedStdout == "" {
				t.Fatal("no expectations defined for this test")
			}
		})
	}
}
