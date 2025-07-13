package terraform

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestE2eVersionCheck(t *testing.T) {
	var testScenarios = map[string]e2eTest{
		"nominal": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
  required_version = ">= 1.0.0, < 2.0.0"
}`,
				},
			},
			args:     []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: nil,
		},
		"fixed_version": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
  required_version = "= 1.3.2"
}`,
				},
			},
			args: []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "main.tf", line: 2, message: "Fixed terraform version are not allowed in module ."},
			},
		},
		"not_defined": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
}`,
				},
			},
			args: []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "main.tf", line: 1, message: "Terraform version is not defined in module ."},
			},
		},
		"no_version": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
  required_version = ""
}`,
				},
			},
			args: []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "main.tf", line: 2, message: "Version constraint empty or malformed for terraform version in module .: improper constraint: "},
			},
		},
		"only_exclusions": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
  required_version = "!= 1.3.2"
}`,
				},
			},
			args: []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "main.tf", line: 2, message: "Version constraints cannot contain only exclusions for terraform version in module ."},
			},
		},
		"malformed_version": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
  required_version = ">= 1.a.0, < 1.3.2"
}`,
				},
			},
			args: []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "main.tf", line: 2, message: "Version constraint empty or malformed for terraform version in module .: improper constraint: >= 1.a.0"},
			},
		},
		"match_only_deprecated": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
  required_version = ">= 0.14.11, < 1.0.0"
}`,
				},
			},
			args: []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: []githubAnnotation{
				{level: "warning", path: "main.tf", line: 2, message: "Terraform version defined only matches a deprecated version in module ."},
			},
		},
		"no_match": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
  required_version = "> 0.1.0, < 0.2.0"
}`,
				},
			},
			args: []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: []githubAnnotation{
				{level: "warning", path: "main.tf", line: 2, message: "Terraform version defined does not match any allowed version in module ."},
			},
		},
		"only_ge_bound": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
  required_version = ">= 1.0.0"
}`,
				},
			},
			args: []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "main.tf", line: 2, message: "Version range must be defined with two bounds for terraform version in module ."},
			},
		},
		"only_gt_bound": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
  required_version = "> 1.0.0"
}`,
				},
			},
			args: []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "main.tf", line: 2, message: "Version range must be defined with two bounds for terraform version in module ."},
			},
		},
		"only_le_bound": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
  required_version = "<= 2.0.0"
}`,
				},
			},
			args: []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "main.tf", line: 2, message: "Version range must be defined with two bounds for terraform version in module ."},
			},
		},
		"only_lt_bound": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
  required_version = "< 2.0.0"
}`,
				},
			},
			args: []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "main.tf", line: 2, message: "Version range must be defined with two bounds for terraform version in module ."},
			},
		},
		"more_than_one_minimum_bound": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
  required_version = ">= 1.0.0, > 1.2.0, < 3.0.0"
}`,
				},
			},
			args: []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: []githubAnnotation{
				{level: "warning", path: "main.tf", line: 2, message: "Version range should define only one minimum bound for terraform version in module ."},
			},
		},
		"more_than_one_maximum_bound": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
  required_version = ">= 1.0.0, <= 2.0.0, < 3.0.0"
}`,
				},
			},
			args: []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: []githubAnnotation{
				{level: "warning", path: "main.tf", line: 2, message: "Version range should define only one maximum bound for terraform version in module ."},
			},
		},
		"more_than_two_major_wide": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
  required_version = ">= 1.0.0, < 4.0.0"
}`,
				},
			},
			args: []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "main.tf", line: 2, message: "Maximum bound cannot be more than two Major versions above minimum bound for terraform version in module ."},
			},
		},
		"too_wide": {
			initial: []testFile{
				{
					path: "main.tf",
					content: `terraform {
  required_version = ">= 1.0.0, <= 3.0.0"
}`,
				},
			},
			args: []string{"version", "check", "--file", "../terraform_versions.yml"},
			expected: []githubAnnotation{
				{level: "error", path: "main.tf", line: 2, message: "Version constraint can only be defined to two next Major version - excluded - for terraform version in module ."},
			},
		},
	}

	dir, err := os.MkdirTemp("", "TestTerraformVersionCheck")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	var versionsRef = `---
deprecated: 0.14.11
stable: 1.3.2
next: 1.4.4`

	err = helper.WriteFile(versionsRef, dir+"/"+"/terraform_versions.yml")
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
					assert.ErrorContains(t, err, fmt.Sprintf("found %d errors for terraform version definition, please fix them", numberAnnotations["error"]))
				}
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}
