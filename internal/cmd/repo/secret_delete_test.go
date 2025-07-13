package repo

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestSecretDelete(t *testing.T) {
	var testScenarios = map[string]kubeSecretE2eTest{
		"existing_file_existing_secret": {
			KubeSecret: []*testFile{{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      dev-aws-de-fra-1/foo:
        keys:
          - FOO
        version: 3
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`}},
			Args: []string{"secret", "delete", "--platform", "dev", "--environment", "de", "--path", "test", "--name", "app"},
			Expected: testFile{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      dev-aws-de-fra-1/foo:
        keys:
          - FOO
        version: 3
`},
		},
		"existing_file_existing_secret_common_shared": {
			KubeSecret: []*testFile{{
				Path: "common-shared/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/common
    vaultRole: foo
    secrets:
      common/foo:
        keys:
          - FOO
        version: 3
      common/test:
        keys:
          - foo
        version: 1
`}},
			Args: []string{"secret", "delete", "--platform", "common", "--path", "test", "--name", "app"},
			Expected: testFile{
				Path: "common-shared/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/common
    vaultRole: foo
    secrets:
      common/foo:
        keys:
          - FOO
        version: 3
`},
		},
		"existing_file_existing_secret_common_prod": {
			KubeSecret: []*testFile{{
				Path: "common-prod/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/common
    vaultRole: foo
    secrets:
      common/foo:
        keys:
          - FOO
        version: 3
      common/test:
        keys:
          - foo
        version: 1
`}},
			Args: []string{"secret", "delete", "--platform", "common", "--path", "test", "--name", "app", "--prod"},
			Expected: testFile{
				Path: "common-prod/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/common
    vaultRole: foo
    secrets:
      common/foo:
        keys:
          - FOO
        version: 3
`},
		},
		"existing_file_not_existing_secret": {
			KubeSecret: []*testFile{{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      dev-aws-de-fra-1/foo:
        keys:
          - FOO
        version: 3
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`}},
			Args:     []string{"secret", "delete", "--platform", "dev", "--environment", "fr", "--path", "test", "--name", "app"},
			Expected: helper.SecretExpectedError{Name: "errSecretDoesNotExist", Error: errSecretDoesNotExist},
		},
		"not_existing_file": {
			Args:     []string{"secret", "delete", "--platform", "dev", "--environment", "de", "--path", "test", "--name", "app"},
			Expected: helper.SecretExpectedError{Name: "errLogicalSecretResourceDoesNotExist", Error: errLogicalSecretResourceDoesNotExist},
		},
		"existing_file_last_secret": {
			KubeSecret: []*testFile{{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`}},
			Args:     []string{"secret", "delete", "--platform", "dev", "--environment", "de", "--path", "test", "--name", "app"},
			Expected: nil,
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			initE2ETest()

			dir, err := os.MkdirTemp("", "TestSecretDelete")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)

			err = os.Chdir(dir)
			if err != nil {
				t.Fatal(err)
			}
			cwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			err = os.Setenv("KUBE_REPOSITORY_PATH", cwd)
			if err != nil {
				t.Fatal(err)
			}

			err = helper.CreateDirectory(dir + "/configs")
			if err != nil {
				t.Fatal(err)
			}
			err = helper.CreateDirectory(dir + "/configs/vault-secrets")
			if err != nil {
				t.Fatal(err)
			}

			if v.KubeSecret != nil {
				for _, ks := range v.KubeSecret {
					err = helper.CreateDirectory(dir + "/configs/vault-secrets/" + filepath.Dir(ks.Path))
					if err != nil {
						t.Fatal(err)
					}
					file, err := os.Create(dir + "/configs/vault-secrets/" + ks.Path)
					if err != nil {
						t.Fatal(err)
					}
					_, err = file.Write([]byte(ks.Content))
					if err != nil {
						t.Fatal(err)
					}
					file.Close()
				}
			}

			_, _, err = helper.ExecuteCobraCommand(repoCmd, v.Args, strings.NewReader(""))
			if expectedFile, ok := v.Expected.(testFile); ok {
				assert.NoError(t, err, "error should be nil")
				content, err := os.ReadFile(dir + "/configs/vault-secrets/" + expectedFile.Path)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, expectedFile.Content, string(content), "should be equal")
			} else if expectedError, ok := v.Expected.(helper.SecretExpectedError); ok {
				assert.ErrorContains(t, err, expectedError.Error.Error(), "should have an error")
			} else if v.Expected == nil {
				for _, ks := range v.KubeSecret {
					if _, err := os.Stat(dir + "/configs/vault-secrets/" + ks.Path); err == nil {
						assert.Fail(t, "secret file should not exists anymore")
					} else if !errors.Is(err, os.ErrNotExist) {
						t.Fatal(err)
					}
				}
			} else {
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}
