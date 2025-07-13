package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestSecretAdd(t *testing.T) {
	var testScenarios = map[string]kubeSecretE2eTest{
		"existing_file_all_keys": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
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
`}},
			Args: []string{"secret", "add", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "1", "--name", "app", "--all-keys"},
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
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`},
		},
		"new_file_all_keys": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			Args:         []string{"secret", "add", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "1", "--name", "app", "--vault-role", "foo", "--all-keys"},
			Expected: testFile{
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
`},
		},
		"new_file_all_keys_common_shared": {
			VaultSecrets: []helper.Secret{{Path: "common/test", Data: map[string]interface{}{"foo": "bar"}}},
			Args:         []string{"secret", "add", "--platform", "common", "--path", "test", "--version", "1", "--name", "app", "--vault-role", "foo", "--all-keys"},
			Expected: testFile{
				Path: "common-shared/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/common
    vaultRole: foo
    secrets:
      common/test:
        keys:
          - foo
        version: 1
`},
		},
		"new_file_all_keys_common_prod": {
			VaultSecrets: []helper.Secret{{Path: "common/test", Data: map[string]interface{}{"foo": "bar"}}},
			Args:         []string{"secret", "add", "--platform", "common", "--path", "test", "--version", "1", "--name", "app", "--vault-role", "foo", "--all-keys", "--prod"},
			Expected: testFile{
				Path: "common-prod/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/common
    vaultRole: foo
    secrets:
      common/test:
        keys:
          - foo
        version: 1
`},
		},
		"new_file_specific_keys": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar", "baz": "qux"}}},
			Args:         []string{"secret", "add", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "1", "--name", "app", "--vault-role", "foo", "--keys", "foo,baz"},
			Expected: testFile{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - baz
          - foo
        version: 1
`},
		},
		"non_existent_key": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			Args:         []string{"secret", "add", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "1", "--name", "app", "--vault-role", "foo", "--keys", "bar"},
			Expected:     helper.SecretExpectedError{Name: "errKeyNotFound", Error: fmt.Errorf("%v: bar", errKeyNotFound)},
		},
		"new_file_missing_vault_role": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			Args:         []string{"secret", "add", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "1", "--name", "app", "--all-keys"},
			Expected:     helper.SecretExpectedError{Name: "errVaultRoleEmpty", Error: errVaultRoleEmpty},
		},
		"secret_does_not_exist": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			Args:         []string{"secret", "add", "--platform", "dev", "--environment", "de", "--path", "test2", "--version", "1", "--name", "app", "--all-keys", "--vault-role", "test"},
			Expected:     helper.SecretExpectedError{Name: "helper.ErrSecretNotFound", Error: helper.ErrSecretNotFound},
		},
		"secret_version_does_not_exist": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			Args:         []string{"secret", "add", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "2", "--name", "app", "--all-keys", "--vault-role", "test"},
			Expected:     helper.SecretExpectedError{Name: "helper.ErrSecretNotFound", Error: helper.ErrSecretNotFound},
		},
		"existing_file_vault_role_provided": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
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
`}},
			Args:     []string{"secret", "add", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "1", "--name", "app", "--vault-role", "foo", "--all-keys"},
			Expected: helper.SecretExpectedError{Name: "errVaultRoleNotEmpty", Error: errVaultRoleNotEmpty},
		},
		"secret_already_exists": {
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
			Args:     []string{"secret", "add", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "1", "--name", "app", "--all-keys"},
			Expected: helper.SecretExpectedError{Name: "errSecretAlreadyExists", Error: errSecretAlreadyExists},
		},
		"secret_already_exists_common_shared": {
			KubeSecret: []*testFile{{
				Path: "common-shared/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/common
    vaultRole: foo
    secrets:
      common/test:
        keys:
          - foo
        version: 1
`}},
			Args:     []string{"secret", "add", "--platform", "common", "--path", "test", "--version", "1", "--name", "app", "--all-keys"},
			Expected: helper.SecretExpectedError{Name: "errSecretAlreadyExists", Error: errSecretAlreadyExists},
		},
		"secret_already_exists_common_prod": {
			KubeSecret: []*testFile{{
				Path: "common-prod/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/common
    vaultRole: foo
    secrets:
      common/test:
        keys:
          - foo
        version: 1
`}},
			Args:     []string{"secret", "add", "--platform", "common", "--path", "test", "--version", "1", "--name", "app", "--all-keys", "--prod"},
			Expected: helper.SecretExpectedError{Name: "errSecretAlreadyExists", Error: errSecretAlreadyExists},
		},
		"keys_not_provided": {
			Args:     []string{"secret", "add", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "1", "--name", "app", "--vault-role", "foo"},
			Expected: helper.SecretExpectedError{Name: "errKeysNotProvided", Error: errKeysNotProvided},
		},
		"new_file_with_jwt_subjects": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			Args:         []string{"secret", "add", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "1", "--name", "app", "--vault-role", "foo", "--all-keys", "--tfe-jwt-subjects", "organization:myorg:project:myproject:workspace:prod,organization:*:project:*:workspace:*"},
			Expected: testFile{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:*:project:*:workspace:*
      - organization:myorg:project:myproject:workspace:prod
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`},
		},
		"new_file_with_single_jwt_subject": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			Args:         []string{"secret", "add", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "1", "--name", "app", "--vault-role", "foo", "--all-keys", "--tfe-jwt-subjects", "organization:myorg:project:myproject:workspace:prod"},
			Expected: testFile{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:myorg:project:myproject:workspace:prod
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`},
		},
		"new_file_without_jwt_subjects": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			Args:         []string{"secret", "add", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "1", "--name", "app", "--vault-role", "foo", "--all-keys"},
			Expected: testFile{
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
`},
		},
		"new_file_with_invalid_jwt_subject_format": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			Args:         []string{"secret", "add", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "1", "--name", "app", "--vault-role", "foo", "--all-keys", "--tfe-jwt-subjects", "invalid-format"},
			Expected:     helper.SecretExpectedError{Name: "errInvalidTfeJwtSubjectFormat", Error: errInvalidTfeJwtSubjectFormat},
		},
		"new_file_with_missing_jwt_subject_parts": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			Args:         []string{"secret", "add", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "1", "--name", "app", "--vault-role", "foo", "--all-keys", "--tfe-jwt-subjects", "organization:myorg:project:myproject"},
			Expected:     helper.SecretExpectedError{Name: "errInvalidTfeJwtSubjectFormat", Error: errInvalidTfeJwtSubjectFormat},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			initE2ETest()
			if v.VaultSecrets != nil {
				cluster, _ := setupVaultCluster(t, v.VaultSecrets...)
				defer cluster.Cleanup()
			}

			dir, err := os.MkdirTemp("", "TestSecretAdd")
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
			} else {
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}
