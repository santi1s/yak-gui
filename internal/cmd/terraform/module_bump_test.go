package terraform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestUniqDirs(t *testing.T) {
	initial := []string{"/tmp/dir1/foo", "/tmp/dir1/bar", "/tmp/dir2/baz"}
	expected := []string{"/tmp/dir1", "/tmp/dir2"}
	result := uniqDir(initial)
	assert.Equal(t, result, expected, "array should be equal")
}

func TestBump(t *testing.T) {
	var testScenarios = map[string]e2eTest{
		"nothing_to_do": {
			args:           []string{"module", "bump", "-n", "foo/bar", "-v", "1.2.3", "--check"},
			expectedStdout: "",
			expectedStderr: "",
		},
		"works_ok_checked": {
			initial: []testFile{
				{
					path: "/bar/foo/foo.tf",
					content: `
module "vault" {
  source               = "tfe.doctolib.net/doctolib/foo/bar"
  version              = "1.3.0"
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
  version    = "0.1.1"
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
  version              = "1.3.0"
  vpc_id               = data.aws_vpc.fra_vpc.id
}
`,
				},
			},
			args:           []string{"module", "bump", "-n", "foo/bar", "-v", "1.2.3", "--check"},
			expectedStdout: "module bumped in bar/foo/foo.tf\n",
			expectedStderr: "Issues were detected in the following 1 directories:\nfoo/bar\n",
		},
	}

	for k, v := range testScenarios {
		initE2ETest()

		dir, err := os.MkdirTemp("", "TestBump")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dir)
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
			output, stderr, err := helper.ExecuteCobraCommand(terraformCmd, v.args, strings.NewReader(""))
			assert.NoError(t, err, "error should be nil")
			if v.expectedStdout != "" {
				assert.Equal(t, v.expectedStdout, output, "should be equal")
			}

			if v.expectedStderr != "" {
				assert.Equal(t, v.expectedStderr, stderr, "should be equal")
			}
		})
	}
}
