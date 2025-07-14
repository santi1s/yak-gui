package repo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santi1s/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestSecretBump(t *testing.T) {
	var testScenarios = map[string]kubeSecretE2eTest{
		"existing_files": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}, {Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "baz"}}},
			KubeSecret: []*testFile{
				{
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
`},
				{
					Path: "dev/app2.yml",
					Content: `---
vaultSecrets:
  app2:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`},
				{
					Path: "staging/app3.yml",
					Content: `---
vaultSecrets:
  app3:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      staging/test:
        keys:
          - foo
        version: 1
`},
			},
			Args: []string{"secret", "bump", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "2"},
			Expected: []testFile{
				{
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
        version: 2
`},
				{
					Path: "dev/app2.yml",
					Content: `---
vaultSecrets:
  app2:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 2
`},
				{
					Path: "staging/app3.yml",
					Content: `---
vaultSecrets:
  app3:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      staging/test:
        keys:
          - foo
        version: 1
`},
			},
		},
		"existing_files_common_shared": {
			VaultSecrets: []helper.Secret{{Path: "common/test", Data: map[string]interface{}{"foo": "bar"}}, {Path: "common/test", Data: map[string]interface{}{"foo": "baz"}}},
			KubeSecret: []*testFile{
				{
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
`},
				{
					Path: "common-shared/app2.yml",
					Content: `---
vaultSecrets:
  app2:
    vaultNamespace: doctolib/common
    vaultRole: foo
    secrets:
      common/test:
        keys:
          - foo
        version: 1
`},
				{
					Path: "staging/app3.yml",
					Content: `---
vaultSecrets:
  app3:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      staging/test:
        keys:
          - foo
        version: 1
`},
			},
			Args: []string{"secret", "bump", "--platform", "common", "--path", "test", "--version", "2"},
			Expected: []testFile{
				{
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
        version: 2
`},
				{
					Path: "common-shared/app2.yml",
					Content: `---
vaultSecrets:
  app2:
    vaultNamespace: doctolib/common
    vaultRole: foo
    secrets:
      common/test:
        keys:
          - foo
        version: 2
`},
				{
					Path: "staging/app3.yml",
					Content: `---
vaultSecrets:
  app3:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      staging/test:
        keys:
          - foo
        version: 1
`},
			},
		},
		"existing_files_common_prod": {
			VaultSecrets: []helper.Secret{{Path: "common/test", Data: map[string]interface{}{"foo": "bar"}}, {Path: "common/test", Data: map[string]interface{}{"foo": "baz"}}},
			KubeSecret: []*testFile{
				{
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
`},
				{
					Path: "common-prod/app2.yml",
					Content: `---
vaultSecrets:
  app2:
    vaultNamespace: doctolib/common
    vaultRole: foo
    secrets:
      common/test:
        keys:
          - foo
        version: 1
`},
				{
					Path: "staging/app3.yml",
					Content: `---
vaultSecrets:
  app3:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      staging/test:
        keys:
          - foo
        version: 1
`},
			},
			Args: []string{"secret", "bump", "--platform", "common", "--path", "test", "--version", "2", "--prod"},
			Expected: []testFile{
				{
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
        version: 2
`},
				{
					Path: "common-prod/app2.yml",
					Content: `---
vaultSecrets:
  app2:
    vaultNamespace: doctolib/common
    vaultRole: foo
    secrets:
      common/test:
        keys:
          - foo
        version: 2
`},
				{
					Path: "staging/app3.yml",
					Content: `---
vaultSecrets:
  app3:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      staging/test:
        keys:
          - foo
        version: 1
`},
			},
		},
		"no_reference_files_found": {
			Args:     []string{"secret", "bump", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "2"},
			Expected: helper.SecretExpectedError{Name: "errNoReferenceFilesFounds", Error: errNoReferenceFilesFound},
		},
		"version_doesnt_exist": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			KubeSecret: []*testFile{
				{
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
			Args:     []string{"secret", "bump", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "100"},
			Expected: helper.SecretExpectedError{Name: "helper.ErrSecretNotFound", Error: helper.ErrSecretNotFound},
		},
		"path_doesnt_exist": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			KubeSecret: []*testFile{
				{
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
			Args:     []string{"secret", "bump", "--platform", "dev", "--environment", "de", "--path", "notfound", "--version", "1"},
			Expected: helper.SecretExpectedError{Name: "helper.ErrSecretNotFound", Error: helper.ErrSecretNotFound},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			initE2ETest()
			if v.VaultSecrets != nil {
				cluster, _ := setupVaultCluster(t, v.VaultSecrets...)
				defer cluster.Cleanup()
			}

			dir, err := os.MkdirTemp("", "TestSecretBump")
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
			if expectedFile, ok := v.Expected.([]testFile); ok {
				assert.NoError(t, err, "error should be nil")
				for _, f := range expectedFile {
					content, err := os.ReadFile(dir + "/configs/vault-secrets/" + f.Path)
					if err != nil {
						t.Fatal(err)
					}
					assert.Equal(t, f.Content, string(content), "should be equal")
				}
			} else if expectedError, ok := v.Expected.(helper.SecretExpectedError); ok {
				assert.ErrorIs(t, err, expectedError.Error, "should have an error")
			} else {
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}
