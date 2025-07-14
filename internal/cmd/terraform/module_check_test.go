package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santi1s/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestModuleCheck(t *testing.T) {
	var testScenarios = map[string]e2eTest{
		"nothing_to_do": {
			args:     []string{"module", "check"},
			expected: nil,
		},
		"whitelisted_paths": {
			initial: []testFile{
				{
					path: "/envs/shared/terraform/02_tfe_staging/test.tf",
					content: `
module "vault" {
  source 			   = "git@github.com:doctolib/terraform-aws-asg.git?ref=3.3.0"
  resource_name_prefix = "${var.vault_env}-de-fra"
  vpc_id               = data.aws_vpc.fra_vpc.id
}
`,
				},
			},
			args:     []string{"module", "check"},
			expected: nil,
		},
		"works_ok_checked": {
			initial: []testFile{
				{
					path: "/bar/foo/foo.tf",
					content: `
module "vault" {
  source               = "tfe.doctolib.net/doctolib/foo/bar"
  version              = "0.0.0-pr123"
  resource_name_prefix = "${var.vault_env}-de-fra"
  vpc_id               = data.aws_vpc.fra_vpc.id
}

module "cf_custom_pages" {
  source     = "tfe.doctolib.net/doctolib/custom-pages/cloudflare"
  version    = "0.1.1"
  account_id = "account_id"
}
`,
				},
				{
					path: "/.terraform/modules/bar.tf",
					content: `
module "vault" {
  source               = "tfe.doctolib.net/doctolib/foo/bar"
  version              = "1.3.0"
  resource_name_prefix = "${var.vault_env}-de-fra"
  vpc_id               = data.aws_vpc.fra_vpc.id
}

module "cf_custom_pages" {
  source     = "tfe.doctolib.net/doctolib/custom-pages/cloudflare"
  version    = "0.c.3"
  account_id = "account_id"
}
`,
				},
				{
					path: "/foo/bar/baz.tf",
					content: `
module "vault" {
  source               = "tfe.doctolib.net/doctolib/foo/bar"
  resource_name_prefix = "${var.vault_env}-de-fra"
  version              = "10.3.2-1.2"
  vpc_id               = data.aws_vpc.fra_vpc.id
}

module "test" {
	source               = "tfe.doctolib.net/doctolib/foo/bar"
	resource_name_prefix = "${var.vault_env}-de-fra"
	vpc_id               = data.aws_vpc.fra_vpc.id
  }

module "datadog" {
	source               = "../../terraform/modules/datadog_monitors"
	resource_name_prefix = "${var.vault_env}-de-fra"
	vpc_id               = data.aws_vpc.fra_vpc.id
  }
`,
				},
				{
					path: "/foo/bar/bar/baz.tf",
					content: `
module "vault" {
  source               = "tfe.doctolib.net/doctolib/foo/bar"
  resource_name_prefix = "${var.vault_env}-de-fra"
  version              = "version"
  vpc_id               = data.aws_vpc.fra_vpc.id
  name                 = "vault-2"
}

module "vault_test" {
	source               = "tfe.doctolib.net/doctolib/foo/bar"
	resource_name_prefix = "${var.vault_env}-de-fra"
	version              = "1.2.3"
	vpc_id               = data.aws_vpc.fra_vpc.id
	name                 = "vault-2"
  }

module "test_foo_bar" {
	source               = "tfe.doctolib.net/doctolib/foo/bar"
	version              = "0.0.0-pr123"
	resource_name_prefix = "${var.vault_env}-de-fra"
	vpc_id               = data.aws_vpc.fra_vpc.id
  }

module "datadog_whitelisted" {
	source               = "../../../../terraform/modules/datadog-monitors"
	resource_name_prefix = "${var.vault_env}-de-fra"
	vpc_id               = data.aws_vpc.fra_vpc.id
  }

module "github-aws-runners_multi-runner_whitelisted" {
	source               = "github-aws-runners/github-runner/aws//modules/multi-runner"
	resource_name_prefix = "${var.vault_env}-de-fra"
	vpc_id               = data.aws_vpc.fra_vpc.id
  }

module "philips_multi-runner_whitelisted" {
	source               = "philips-labs/github-runner/aws//modules/multi-runner"
	resource_name_prefix = "${var.vault_env}-de-fra"
	vpc_id               = data.aws_vpc.fra_vpc.id
  }

module "philips_webhook_whitelisted" {
	source               = "philips-labs/github-runner/aws//modules/webhook-github-app"
	resource_name_prefix = "${var.vault_env}-de-fra"
	vpc_id               = data.aws_vpc.fra_vpc.id
  }
`,
				},
			},
			args: []string{"module", "check"},
			expected: []githubAnnotation{
				{level: "error", path: "foo/bar/baz.tf", line: 2, message: "Module vault does not have a valid version 10.3.2-1.2"},
				{level: "error", path: "foo/bar/baz.tf", line: 9, message: "Module test does not have a version"},
				{level: "error", path: "foo/bar/baz.tf", line: 15, message: "Module datadog relative sources are forbidden, please use --allow-relative-sources"},
				{level: "error", path: "bar/foo/foo.tf", line: 2, message: "Module vault with version 0.0.0-pr123 has a prerelease version"},
				{level: "error", path: "foo/bar/bar/baz.tf", line: 2, message: "Module vault with version version does not have a semver valid version"},
				{level: "error", path: "foo/bar/bar/baz.tf", line: 18, message: "Module test_foo_bar with version 0.0.0-pr123 has a prerelease version"},
			},
		},
		"relative_path_allowed": {
			initial: []testFile{
				{
					path: "foo/bar/test.tf",
					content: `
module "should_be_allowed" {
	source = "../path/to/module"
}					
`,
				},
			},
			args:     []string{"module", "check", "--allow-relative-sources"},
			expected: nil,
		},
		"relative_path_denied": {
			initial: []testFile{
				{
					path: "/test2.tf",
					content: `
module "should_be_denied" {
    source = "../path/to/module"
}
					`,
				},
			},
			args: []string{"module", "check"},
			expected: []githubAnnotation{
				{level: "error", path: "test2.tf", line: 2, message: "Module should_be_denied relative sources are forbidden, please use --allow-relative-sources"},
			},
		},
	}

	for k, v := range testScenarios {
		initE2ETest()

		dir, err := os.MkdirTemp("", "TestModuleCheck")
		if err != nil {
			t.Fatal(err)
		}
		// defer os.RemoveAll(dir)
		err = helper.CreateDirectory(dir)
		if err != nil {
			t.Fatal(err)
		}

		if v.initial != nil {
			for _, testFile := range v.initial {
				err = helper.CreateDirectory(dir + filepath.Dir(testFile.path))
				if err != nil {
					t.Fatal(err)
				}
				file, err := os.Create(dir + testFile.path)
				if err != nil {
					t.Fatal(err)
				}
				_, err = file.Write([]byte(testFile.content))
				if err != nil {
					t.Fatal(err)
				}
				file.Close()
			}
		}

		err = os.Chdir(dir)
		if err != nil {
			panic(err)
		}

		t.Run(k, func(t *testing.T) {
			stdout, stderr, err := helper.ExecuteCobraCommand(terraformCmd, v.args, strings.NewReader(""))
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
					assert.ErrorContains(t, err, fmt.Sprintf("found %d errors in modules, please fix them", numberAnnotations["error"]))
				}
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}
