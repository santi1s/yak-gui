package terraform

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestE2eDeclarationCheck(t *testing.T) {
	var testScenarios = map[string]e2eTest{
		"cloud_outside_backend": {
			initial: []testFile{
				{
					path: "ver.tf",
					content: `terraform {
	cloud {
	}
}`,
				},
			},
			args: []string{"declaration", "check"},
			expected: []githubAnnotation{
				{level: "error", path: "ver.tf", line: 2, message: "cloud declared in 'ver.tf' must be declared in `backend.tf`"},
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
					assert.ErrorContains(t, err, fmt.Sprintf("found %d errors, please fix them", numberAnnotations["error"]))
				}
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}
