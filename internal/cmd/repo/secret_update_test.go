package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santi1s/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestSecretUpdate(t *testing.T) {
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
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`}},
			Args: []string{"secret", "update", "--platform", "dev", "--environment", "de", "--path", "test", "--name", "app", "--all-keys"},
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
		"existing_file_all_keys_common_shared": {
			VaultSecrets: []helper.Secret{{Path: "common/test", Data: map[string]interface{}{"foo": "bar"}}},
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
			Args: []string{"secret", "update", "--platform", "common", "--path", "test", "--name", "app", "--all-keys"},
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
      common/test:
        keys:
          - foo
        version: 1
`},
		},
		"existing_file_all_keys_common_prod": {
			VaultSecrets: []helper.Secret{{Path: "common/test", Data: map[string]interface{}{"foo": "bar"}}},
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
			Args: []string{"secret", "update", "--platform", "common", "--path", "test", "--name", "app", "--all-keys", "--prod"},
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
      common/test:
        keys:
          - foo
        version: 1
`},
		},
		"existing_file_specific_keys": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar", "bar": "baz"}}},
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
			Args: []string{"secret", "update", "--platform", "dev", "--environment", "de", "--path", "test", "--name", "app", "--keys", "bar"},
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
          - bar
        version: 1
`},
		},
		"existing_file_specific_keys_doesnt_exist": {
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
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`}},
			Args:     []string{"secret", "update", "--platform", "dev", "--environment", "de", "--path", "test", "--name", "app", "--keys", "bar"},
			Expected: helper.SecretExpectedError{Name: "errKeyNotFound", Error: fmt.Errorf("%v: bar", errKeyNotFound)},
		},
		"existing_file_version": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}, {Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "baz"}}},
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
			Args: []string{"secret", "update", "--platform", "dev", "--environment", "de", "--path", "test", "--name", "app", "--version", "2"},
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
        version: 2
`},
		},
		"existing_file_all_keys_and_version": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}, {Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "baz"}}},
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
			Args: []string{"secret", "update", "--platform", "dev", "--environment", "de", "--path", "test", "--name", "app", "--all-keys", "--version", "2"},
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
        version: 2
`},
		},
		"existing_file_specific_keys_and_version": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar", "bar": "baz"}}, {Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar", "bar": "foo"}}},
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
			Args: []string{"secret", "update", "--platform", "dev", "--environment", "de", "--path", "test", "--name", "app", "--version", "2", "--keys", "bar"},
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
          - bar
        version: 2
`},
		},
		"existing_file_vault_role": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar", "bar": "baz"}}},
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
			Args: []string{"secret", "update", "--platform", "dev", "--name", "app", "--vault-role", "bar"},
			Expected: testFile{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: bar
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
		"existing_file_vault_role_and_other_params": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar", "bar": "baz"}}},
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
			Args:     []string{"secret", "update", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "2", "--all-keys", "--name", "app", "--vault-role", "bar"},
			Expected: helper.SecretExpectedError{Name: "errVaultRoleAndOtherFlagsCantBeSetTogether", Error: errVaultRoleAndOtherFlagsCantBeSetTogether},
		},
		"keys_and_all_keys_together": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			Args:         []string{"secret", "update", "--platform", "dev", "--environment", "de", "--path", "test2", "--name", "app", "--all-keys", "--keys", "foo"},
			Expected:     helper.SecretExpectedError{Name: "errAllKeysAndKeysCantBeSetTogether", Error: errAllKeysAndKeysCantBeSetTogether},
		},
		"non_existing_file": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			Args:         []string{"secret", "update", "--platform", "dev", "--environment", "de", "--path", "test2", "--name", "app", "--all-keys"},
			Expected:     helper.SecretExpectedError{Name: "errLogicalSecretResourceDoesNotExist", Error: errLogicalSecretResourceDoesNotExist},
		},
		"secret_does_not_exist": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
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
          - FOO
        version: 3
`}},
			Args:     []string{"secret", "update", "--platform", "dev", "--environment", "de", "--path", "test2", "--version", "1", "--name", "app", "--all-keys"},
			Expected: helper.SecretExpectedError{Name: "helper.ErrSecretNotFound", Error: helper.ErrSecretNotFound},
		},
		"secret_version_does_not_exist": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
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
          - FOO
        version: 3
`}},
			Args:     []string{"secret", "update", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "2", "--name", "app", "--all-keys"},
			Expected: helper.SecretExpectedError{Name: "helper.ErrSecretNotFound", Error: helper.ErrSecretNotFound},
		},
		"non_existing_secret": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar", "bar": "baz"}}},
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
			Args:     []string{"secret", "update", "--platform", "dev", "--environment", "de", "--path", "test2", "--name", "app", "--keys", "bar"},
			Expected: helper.SecretExpectedError{Name: "errSecretDoesNotExist", Error: errSecretDoesNotExist},
		},
		"add_jwt_subjects": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			KubeSecret: []*testFile{{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:existing:project:existing:workspace:existing
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`}},
			Args: []string{"secret", "update", "--platform", "dev", "--name", "app", "--add-tfe-jwt-subjects", "organization:new:project:new:workspace:new,organization:*:project:*:workspace:*"},
			Expected: testFile{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:*:project:*:workspace:*
      - organization:existing:project:existing:workspace:existing
      - organization:new:project:new:workspace:new
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`},
		},
		"remove_jwt_subjects": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			KubeSecret: []*testFile{{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:keep:project:keep:workspace:keep
      - organization:remove:project:remove:workspace:remove
      - organization:*:project:*:workspace:*
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`}},
			Args: []string{"secret", "update", "--platform", "dev", "--name", "app", "--remove-tfe-jwt-subjects", "organization:remove:project:remove:workspace:remove"},
			Expected: testFile{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:*:project:*:workspace:*
      - organization:keep:project:keep:workspace:keep
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`},
		},
		"add_and_remove_jwt_subjects": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			KubeSecret: []*testFile{{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:keep:project:keep:workspace:keep
      - organization:remove:project:remove:workspace:remove
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`}},
			Args: []string{"secret", "update", "--platform", "dev", "--name", "app", "--add-tfe-jwt-subjects", "organization:new:project:new:workspace:new", "--remove-tfe-jwt-subjects", "organization:remove:project:remove:workspace:remove"},
			Expected: testFile{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:keep:project:keep:workspace:keep
      - organization:new:project:new:workspace:new
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`},
		},
		"add_jwt_subjects_to_empty_list": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
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
			Args: []string{"secret", "update", "--platform", "dev", "--name", "app", "--add-tfe-jwt-subjects", "organization:new:project:new:workspace:new"},
			Expected: testFile{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:new:project:new:workspace:new
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`},
		},
		"vault_role_and_jwt_subjects": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
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
			Args: []string{"secret", "update", "--platform", "dev", "--name", "app", "--vault-role", "new-role", "--add-tfe-jwt-subjects", "organization:new:project:new:workspace:new"},
			Expected: testFile{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: new-role
    tfeJwtSubjects:
      - organization:new:project:new:workspace:new
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`},
		},
		"add_jwt_subjects_invalid_format": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
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
			Args:     []string{"secret", "update", "--platform", "dev", "--name", "app", "--add-tfe-jwt-subjects", "invalid-format"},
			Expected: helper.SecretExpectedError{Name: "errInvalidTfeJwtSubjectFormat", Error: errInvalidTfeJwtSubjectFormat},
		},
		"add_jwt_subjects_missing_parts": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
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
			Args:     []string{"secret", "update", "--platform", "dev", "--name", "app", "--add-tfe-jwt-subjects", "organization:myorg:project:myproject"},
			Expected: helper.SecretExpectedError{Name: "errInvalidTfeJwtSubjectFormat", Error: errInvalidTfeJwtSubjectFormat},
		},
		"remove_jwt_subjects_not_found": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			KubeSecret: []*testFile{{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:existing:project:existing:workspace:existing
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`}},
			Args:     []string{"secret", "update", "--platform", "dev", "--name", "app", "--remove-tfe-jwt-subjects", "organization:nonexistent:project:nonexistent:workspace:nonexistent"},
			Expected: helper.SecretExpectedError{Name: "errTfeJwtSubjectNotFound", Error: errTfeJwtSubjectNotFound},
		},
		"add_jwt_subjects_already_exists": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			KubeSecret: []*testFile{{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:existing:project:existing:workspace:existing
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`}},
			Args:     []string{"secret", "update", "--platform", "dev", "--name", "app", "--add-tfe-jwt-subjects", "organization:existing:project:existing:workspace:existing"},
			Expected: helper.SecretExpectedError{Name: "errTfeJwtSubjectAlreadyExists", Error: fmt.Errorf("TFE JWT subject 'organization:existing:project:existing:workspace:existing' already exists in current list")},
		},
		"add_jwt_subjects_multiple_with_duplicate": {
			VaultSecrets: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}}},
			KubeSecret: []*testFile{{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:existing:project:existing:workspace:existing
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
`}},
			Args:     []string{"secret", "update", "--platform", "dev", "--name", "app", "--add-tfe-jwt-subjects", "organization:new:project:new:workspace:new,organization:existing:project:existing:workspace:existing"},
			Expected: helper.SecretExpectedError{Name: "errTfeJwtSubjectAlreadyExists", Error: fmt.Errorf("TFE JWT subject 'organization:existing:project:existing:workspace:existing' already exists in current list")},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			initE2ETest()
			if v.VaultSecrets != nil {
				cluster, _ := setupVaultCluster(t, v.VaultSecrets...)
				defer cluster.Cleanup()
			}

			dir, err := os.MkdirTemp("", "TestSecretUpdate")
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
