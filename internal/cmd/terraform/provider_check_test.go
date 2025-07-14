package terraform

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/santi1s/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestE2eProviderCheck(t *testing.T) {
	var testScenarios = map[string]e2eTest{
		"nominal": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    provider1 = {
      source  = "editor/trusted1"
      version = "= 1.0.0"
    }
	provider2 = {
      source  = "editor/trusted2"
      version = ">= 1.0.0, != 1.2.3, < 2.0.0"
    }
	provider3 = {
      source  = "editor/untrusted1"
      version = "~> 1.0.0"
    }
	provider4 = {
      source  = "editor/untrusted2"
      version = ">=1.0.0,!=1.2.3,<1.8.0"
    }
  }
}`,
				},
			},
			args:     []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: nil,
		},
		"no_source": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    provider1 = {
      source  = ""
      version = "= 1.0.0"
    }
	provider2 = {
      version = "= 1.0.0"
    }
  }
}`,
				},
			},
			args: []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "versions.tf", line: 5, message: "No source defined for provider provider1 in module ."},
				{level: "error", path: "versions.tf", line: 9, message: "No source defined for provider provider2 in module ."},
			},
		},
		"forbidden_source": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    provider1 = {
      source  = "editor/not_in_the_list"
      version = "= 1.0.0"
    }
  }
}`,
				},
			},
			args: []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "versions.tf", line: 6, message: "Forbidden source editor/not_in_the_list for provider provider1 in module ."},
			},
		},
		"no_version": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    provider1 = {
      source  = "editor/trusted1"
      version = ""
    }
	provider2 = {
      source  = "editor/trusted1"
    }
  }
}`,
				},
			},
			args: []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "versions.tf", line: 5, message: "No version defined for provider provider1 in module ."},
				{level: "error", path: "versions.tf", line: 9, message: "No version defined for provider provider2 in module ."},
			},
		},
		"only_exclusions": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    provider1 = {
      source  = "editor/trusted1"
      version = "!= 1.0.0"
    }
  }
}`,
				},
			},
			args: []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "versions.tf", line: 5, message: "Version constraints cannot contain only exclusions for provider provider1 in module ."},
			},
		},
		"malformed_version": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    provider1 = {
      source  = "editor/trusted1"
      version = ">= 1.a.0, < 2.0.0"
    }
    provider2 = {
      source  = "editor/trusted2"
      version = "= 1.0.0.0"
    }
    provider3 = {
      source  = "editor/untrusted1"
      version = "= "
    }
  }
}`,
				},
			},
			args: []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "versions.tf", line: 5, message: "Version constraint empty or malformed for provider provider1 in module .: improper constraint: >= 1.a.0"},
				{level: "error", path: "versions.tf", line: 9, message: "Version constraint empty or malformed for provider provider2 in module .: improper constraint: = 1.0.0.0"},
				{level: "error", path: "versions.tf", line: 13, message: "Version constraint empty or malformed for provider provider3 in module .: improper constraint: = "},
			},
		},
		"exclusion_or_range_with_specific": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    provider1 = {
      source  = "editor/trusted1"
      version = ">= 1.0.0, = 1.2.3, < 2.0.0"
    }
    provider2 = {
      source  = "editor/trusted2"
      version = "= 1.0.0, != 1.2.3"
    }
  }
}`,
				},
			},
			args: []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{
				{level: "warning", path: "versions.tf", line: 5, message: "Version ranges and exclusions should not be defined alongside specific version for provider provider1 in module ."},
				{level: "warning", path: "versions.tf", line: 9, message: "Version ranges and exclusions should not be defined alongside specific version for provider provider2 in module ."},
			},
		},
		"version_one_bound": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    provider1 = {
      source  = "editor/trusted1"
      version = ">= 1.0.0"
    }
    provider2 = {
      source  = "editor/trusted2"
      version = "> 1.0.0"
    }
    provider3 = {
      source  = "editor/untrusted1"
      version = "<= 2.0.0"
    }
    provider4 = {
      source  = "editor/untrusted2"
      version = "< 2.0.0"
    }
  }
}`,
				},
			},
			args: []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "versions.tf", line: 5, message: "Version range must be defined with two bounds for provider provider1 in module ."},
				{level: "error", path: "versions.tf", line: 9, message: "Version range must be defined with two bounds for provider provider2 in module ."},
				{level: "error", path: "versions.tf", line: 13, message: "Version range must be defined with two bounds for provider provider3 in module ."},
				{level: "error", path: "versions.tf", line: 17, message: "Version range must be defined with two bounds for provider provider4 in module ."},
			},
		},
		"more_than_one_minimum_bound": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    provider1 = {
      source  = "editor/trusted1"
      version = ">= 1.0.0, > 2.0.0, < 3.0.0"
    }
  }
}`,
				},
			},
			args: []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{
				{level: "warning", path: "versions.tf", line: 5, message: "Version range should define only one minimum bound for provider provider1 in module ."},
			},
		},
		"more_than_one_maximum_bound": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    provider1 = {
      source  = "editor/trusted1"
      version = ">= 1.0.0, <= 2.0.0, < 3.0.0"
    }
  }
}`,
				},
			},
			args: []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{
				{level: "warning", path: "versions.tf", line: 5, message: "Version range should define only one maximum bound for provider provider1 in module ."},
			},
		},
		"only_exclusion": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    provider1 = {
      source  = "editor/trusted1"
      version = "!= 1.0.0"
    }
  }
}`,
				},
			},
			args: []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "versions.tf", line: 5, message: "Version constraints cannot contain only exclusions for provider provider1 in module ."},
			},
		},
		"more_than_two_major_wide": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    provider1 = {
      source  = "editor/trusted1"
      version = ">= 1.0.0, < 4.0.0"
    }
  }
}`,
				},
			},
			args: []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "versions.tf", line: 5, message: "Maximum bound cannot be more than two Major versions above minimum bound for provider provider1 in module ."},
			},
		},
		"too_wide_for_trusted": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    provider1 = {
      source  = "editor/trusted1"
      version = ">= 1.0.0, <= 3.0.0"
    }
  }
}`,
				},
			},
			args: []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "versions.tf", line: 5, message: "Version constraint can only be defined to two next Major version - excluded - for trusted provider provider1 in module ."},
			},
		},
		"too_wide_for_untrusted": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    provider1 = {
      source  = "editor/untrusted1"
      version = ">= 1.0.0, <= 2.0.0"
    }
    provider2 = {
      source  = "editor/untrusted2"
      version = ">= 1.0.0, < 2.0.0"
    }
  }
}`,
				},
			},
			args: []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "versions.tf", line: 5, message: "Version constraint must use the same Major digit for untrusted provider provider1 in module ."},
				{level: "error", path: "versions.tf", line: 9, message: "Version constraint must use the same Major digit for untrusted provider provider2 in module ."},
			},
		},
		"required_provider_outside_versions": {
			initial: []testFile{
				{
					path: "not_versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    provider1 = {
      source  = "editor/trusted1"
      version = "1.0.0"
    }
  }
}`,
				},
			},
			args: []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "not_versions.tf", line: 4, message: "required provider declared in 'not_versions.tf' must be declared in 'versions.tf'."},
			},
		},
		"provider_wrong_file": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    it = {
      source  = "editor/trusted1"
      version = "1.0.0"
    }
  }
}`,
				},
				{
					path: "provider_not.tf",
					content: `
provider "it" {
}
					`,
				},
			},
			args: []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "provider_not.tf", line: 2, message: "provider 'it' in 'provider_not.tf' should be declared in 'provider_it.tf'"},
			},
		},
		"provider_correct_file": {
			initial: []testFile{
				{
					path: "versions.tf",
					content: `terraform {
  required_version = "= 0.14.11"

  required_providers {
    it = {
      source  = "editor/trusted1"
      version = "1.0.0"
    }
  }
}`,
				},
				{
					path: "provider_it.tf",
					content: `
provider "it" {
}
					`,
				},
				{
					path: "provider_it.autogenerated.tf",
					content: `
provider "it" {
}
					`,
				},
			},
			args:     []string{"provider", "check", "--file", "../terraform_providers.yml"},
			expected: []githubAnnotation{},
		},
	}

	dir, err := os.MkdirTemp("", "TestTerraformProviderCheck")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	var providersRef = `terraform_providers:
  editor/trusted1:
    trusted: true
  editor/trusted2:
    trusted: true
  editor/untrusted1:
    trusted: false
  editor/untrusted2:
    trusted: false`

	err = helper.WriteFile(providersRef, dir+"/"+"/terraform_providers.yml")
	if err != nil {
		t.Fatal(err)
	}

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

			stdout, stderr, err := helper.ExecuteCobraCommand(terraformCmd, v.args)
			if v.expected == nil {
				assert.NoError(t, err)
				assert.Empty(t, stderr, "stderr should be empty")
			} else if expectedAnnotations, ok := v.expected.([]githubAnnotation); ok {
				// If we expect an annotation to be returned
				for _, expectedAnnotation := range expectedAnnotations {
					// Check expected annotations are found
					assert.Contains(t, stdout, fmt.Sprintf("::%s file=%s,line=%d::%s\n", expectedAnnotation.level, expectedAnnotation.path, expectedAnnotation.line, expectedAnnotation.message))
				}
				// Count each kind of annotations we expect
				numberAnnotations := map[string]int{"error": 0, "warning": 0, "notice": 0}
				for _, a := range expectedAnnotations {
					switch a.level {
					case "error":
						numberAnnotations["error"] = numberAnnotations["error"] + 1
					case "warning":
						numberAnnotations["warning"]++
					case "notice":
						numberAnnotations["notice"]++
					}
				}
				// Assert there is the expected number of occurrences for each (especially if 0)
				for lvl, occurrences := range numberAnnotations {
					assert.Equalf(t, occurrences, strings.Count(stdout, fmt.Sprintf("::%s file", lvl)), "Number of occurrence of %s annotation is unexpected", lvl)
				}
				// Check we have a summary if we expect at least one error
				if numberAnnotations["error"] > 0 {
					assert.ErrorContains(t, err, fmt.Sprintf("found %d errors in providers, please fix them", numberAnnotations["error"]))
				}
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}
