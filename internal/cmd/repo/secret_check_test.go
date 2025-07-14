package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/santi1s/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

type secretCheckUnitTest struct {
	Initial      []SecretUsage
	Expected     error
	VaultSecrets []helper.Secret
}

func TestCheckSecretsVersion(t *testing.T) {
	var testScenarios = map[string]secretCheckUnitTest{
		"without_version": {
			Initial: []SecretUsage{
				{File: "1.yml", SecretPath: "secret/path", VaultNamespace: "test"},
			},
			Expected: errSecretVersionCheck,
		},
		"with_invalid_version": {
			Initial: []SecretUsage{
				{File: "1.yml", SecretPath: "secret/path", SecretVersion: "latest", VaultNamespace: "test"},
			},
			Expected: errSecretVersionCheck,
		},
		"valid_version": {
			Initial: []SecretUsage{
				{File: "1.yml", SecretPath: "secret/path", SecretVersion: "1", VaultNamespace: "test"},
			},
			Expected: nil,
		},
		"specific_case_common_valid": {
			Initial: []SecretUsage{
				{File: "1.yml", SecretFolder: "common-shared", SecretPath: "secret/path", SecretVersion: "1", VaultNamespace: "common"},
				{File: "2.yml", SecretFolder: "common-prod", SecretPath: "secret/path", SecretVersion: "2", VaultNamespace: "common"},
			},
			Expected: nil,
		},
		"specific_case_common_invalid": {
			Initial: []SecretUsage{
				{File: "1.yml", SecretFolder: "common-shared", SecretPath: "secret/path", SecretVersion: "1", VaultNamespace: "common"},
				{File: "2.yml", SecretFolder: "common-shared", SecretPath: "secret/path", SecretVersion: "2", VaultNamespace: "common"},
			},
			Expected: errSecretVersionCheck,
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			result, err := checkSecretsVersion(v.Initial)
			if v.Expected == nil {
				assert.Nil(t, err, "error should be nil")
				assert.Equal(t, 0, len(result), "len(result) should equals 0")
			} else {
				assert.ErrorIs(t, err, v.Expected, "error should be errSecretVersionUsageCheck")
				assert.Equal(t, len(v.Initial), len(result), "len(result) should equals 1")
			}
		})
	}
}

func TestCheckSecretsExistence(t *testing.T) {
	var testScenarios = map[string]secretCheckUnitTest{
		"existing_secret_and_version": {
			Initial: []SecretUsage{
				{File: "configs/vault-secrets/dev/app.yml", SecretPath: "dev-aws-de-fra-1/test", SecretVersion: "1", VaultNamespace: "dev"},
			},
			Expected: nil,
			VaultSecrets: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}},
			},
		},
		"not_existing_secret": {
			Initial: []SecretUsage{
				{File: "configs/vault-secrets/dev/app.yml", SecretPath: "dev-aws-de-fra-1/test", SecretVersion: "1", VaultNamespace: "dev"},
			},
			Expected: errSecretExistenceCheck,
			VaultSecrets: []helper.Secret{
				{Path: "dev-aws-de-fra-1/foo", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "ci/dev-aws-de-fra-1/foo", Data: map[string]interface{}{"foo": ""}},
			},
		},
		"not_existing_version": {
			Initial: []SecretUsage{
				{File: "configs/vault-secrets/dev/app.yml", SecretPath: "dev-aws-de-fra-1/test", SecretVersion: "2", VaultNamespace: "dev"},
			},
			Expected: errSecretExistenceCheck,
			VaultSecrets: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}},
			},
		},
		"deleted_version": {
			Initial: []SecretUsage{
				{File: "configs/vault-secrets/dev/app.yml", SecretPath: "dev-aws-de-fra-1/test", SecretVersion: "2", VaultNamespace: "dev"},
			},
			Expected: errSecretExistenceCheck,
			VaultSecrets: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}, Deleted: true},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}, Deleted: true},
			},
		},
		"destroyed_version": {
			Initial: []SecretUsage{
				{File: "configs/vault-secrets/dev/app.yml", SecretPath: "dev-aws-de-fra-1/test", SecretVersion: "2", VaultNamespace: "dev"},
			},
			Expected: errSecretExistenceCheck,
			VaultSecrets: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}, Version: 1, Destroyed: true},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}, Version: 1, Destroyed: true},
			},
		},
		"invalid_version": {
			Initial: []SecretUsage{
				{File: "configs/vault-secrets/dev/app.yml", SecretPath: "dev-aws-de-fra-1/test", SecretVersion: "latest", VaultNamespace: "dev"},
			},
			Expected: errSecretExistenceCheck,
			VaultSecrets: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}},
			},
		},
		"common_shared": {
			Initial: []SecretUsage{
				{File: "configs/vault-secrets/common-shared/app.yml", SecretPath: "common/test", SecretVersion: "2", VaultNamespace: "common"},
			},
			Expected: errSecretExistenceCheck,
			VaultSecrets: []helper.Secret{
				{Path: "common/test", Data: map[string]interface{}{"foo": "bar"}, Deleted: true},
				{Path: "ci/common/test", Data: map[string]interface{}{"foo": ""}, Deleted: true},
			},
		},
		"common_prod": {
			Initial: []SecretUsage{
				{File: "configs/vault-secrets/common-prod/app.yml", SecretPath: "common/test", SecretVersion: "2", VaultNamespace: "common"},
			},
			Expected: errSecretExistenceCheck,
			VaultSecrets: []helper.Secret{
				{Path: "common/test", Data: map[string]interface{}{"foo": "bar"}, Deleted: true},
				{Path: "ci/common/test", Data: map[string]interface{}{"foo": ""}, Deleted: true},
			},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			initE2ETest()
			if v.VaultSecrets != nil {
				cluster, _ := setupVaultCluster(t, v.VaultSecrets...)
				defer cluster.Cleanup()
			}

			result, err := checkSecretsExistence(v.Initial)
			if v.Expected == nil {
				assert.Nil(t, err, "error should be nil")
				assert.Equal(t, 0, len(result), "len(result) should equals 0")
			} else {
				assert.ErrorIs(t, err, v.Expected, "error should be errSecretExistenceCheck")
				assert.Equal(t, 1, len(result), "len(result) should equals 1")
			}
		})
	}
}

func TestGetSecretUsageInYmlReferenceFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestGetSecretUsageInYmlReferenceFiles")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	err = helper.CreateDirectory(dir + "/configs/vault-secrets/dev")
	if err != nil {
		t.Fatal(err)
	}
	err = helper.CreateDirectory(dir + "/configs/vault-secrets/staging")
	if err != nil {
		t.Fatal(err)
	}

	file, err := os.Create(dir + "/configs/vault-secrets/dev/01-test.yml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.Write([]byte(`---
vaultSecrets:
  01-test:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      dev-aws-de-fra-1/test/path:
        keys:
          - NEWRELIC_EVENT_API_ACCOUNT_ID
        version: 3
      dev-aws-de-fra-1/test/path2:
        keys:
          - TRACKING_EVENTS_AWS_SECRET_ACCESS_KEY
        version: 4
      dev-aws-de-fra-1/test/path3:
        keys:
          - DIRECTORY_S3_SECRET_ACCESS_KEY
        version: latest
      dev-aws-de-fra-1/test/path4:
        keys:
          - TRACKING_EVENTS_AWS_SECRET_ACCESS_KEY
          - DIRECTORY_S3_SECRET_ACCESS_KEY`))
	if err != nil {
		t.Fatal(err)
	}
	file.Close()

	file, err = os.Create(dir + "/configs/vault-secrets/staging/02-test.yml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.Write([]byte(`---
vaultSecrets:
  02-test:
    vaultNamespace: doctolib/staging
    vaultRole: foo
    secrets:
      dev-aws-de-fra-1/test/path:
        keys:
          - NEWRELIC_EVENT_API_ACCOUNT_ID
        version: 3
      dev-aws-de-fra-1/test/path2:
        keys:
          - TRACKING_EVENTS_AWS_SECRET_ACCESS_KEY
        version: 4
      dev-aws-de-fra-1/test/path3:
        keys:
          - DIRECTORY_S3_SECRET_ACCESS_KEY
        version: latest
      dev-aws-de-fra-1/test/path4:
        keys:
          - TRACKING_EVENTS_AWS_SECRET_ACCESS_KEY
          - DIRECTORY_S3_SECRET_ACCESS_KEY`))
	if err != nil {
		t.Fatal(err)
	}
	file.Close()

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

	result, err := getSecretUsageInYmlReferenceFiles(dir + "/configs/vault-secrets/*/*.yml")

	assert.NoError(t, err, "there should be no error")
	assert.Equal(t, 8, len(result), "len(result) should be equals to 8")
}

func TestCheckVaultRole(t *testing.T) {
	var testScenarios = map[string]kubeSecretE2eTest{
		"role_ok": {
			KubeSecret: []*testFile{
				{
					Path: "dev/app.yml",
					Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      dev-aws-de-fra-1/test/path:
        keys:
          - NEWRELIC_EVENT_API_ACCOUNT_ID
        version: 3
`},
			},
			Expected: nil,
		},
		"role_not_ok": {
			KubeSecret: []*testFile{
				{
					Path: "dev/app.yml",
					Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole:
    secrets:
      dev-aws-de-fra-1/test/path:
        keys:
          - NEWRELIC_EVENT_API_ACCOUNT_ID
        version: 3
`},
			},
			Expected: errVaultRoleCheck,
		},
	}

	dir, err := os.MkdirTemp("", "TestCheckVaultRole")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
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

			result, err := checkVaultRole(dir + "/configs/vault-secrets/*/*.yml")

			if expectedError, ok := v.Expected.(error); ok {
				assert.ErrorIs(t, err, expectedError, "there should be an error")
				assert.Equal(t, 1, len(result), "len(result) should be equals to 1")
			} else {
				assert.NoError(t, err, "there should be no error")
				assert.Equal(t, 0, len(result), "len(result) should be equals to 0")
			}
		})
	}
}

func TestCheckVaultNamespace(t *testing.T) {
	var testScenarios = map[string]kubeSecretE2eTest{
		"namespace_ok": {
			KubeSecret: []*testFile{
				{
					Path: "dev/app.yml",
					Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      dev-aws-de-fra-1/test/path:
        keys:
          - NEWRELIC_EVENT_API_ACCOUNT_ID
        version: 3
`},
			},
			Expected: nil,
		},
		"namespace_empty": {
			KubeSecret: []*testFile{
				{
					Path: "dev/app.yml",
					Content: `---
vaultSecrets:
  app:
    vaultNamespace:
    vaultRole:
    secrets:
      dev-aws-de-fra-1/test/path:
        keys:
          - NEWRELIC_EVENT_API_ACCOUNT_ID
        version: 3
`},
			},
			Expected: errVaultNamespaceCheck,
		},
		"namespace_not_ok": {
			KubeSecret: []*testFile{
				{
					Path: "dev/app.yml",
					Content: `---
vaultSecrets:
  app:
    vaultNamespace: dev
    vaultRole:
    secrets:
      dev-aws-de-fra-1/test/path:
        keys:
          - NEWRELIC_EVENT_API_ACCOUNT_ID
        version: 3
`},
			},
			Expected: errVaultNamespaceCheck,
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "TestCheckVaultNamespace")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)

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

			result, err := checkVaultNamespace(dir + "/configs/vault-secrets/*/*.yml")

			if expectedError, ok := v.Expected.(error); ok {
				assert.ErrorIs(t, err, expectedError, "there should be an error")
				assert.Equal(t, 1, len(result), "len(result) should be equals to 1")
			} else {
				assert.NoError(t, err, "there should be no error")
				assert.Equal(t, 0, len(result), "len(result) should be equals to 0")
			}
		})
	}
}

func TestCheckTfeJwtSubjects(t *testing.T) {
	var testScenarios = map[string]kubeSecretE2eTest{
		"jwt_subjects_ok": {
			KubeSecret: []*testFile{
				{
					Path: "dev/app.yml",
					Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:doctolib:project:project1:workspace:workspace1
      - organization:doctolib:project:project2:workspace:workspace2
    secrets:
      dev-aws-de-fra-1/test/path:
        keys:
          - NEWRELIC_EVENT_API_ACCOUNT_ID
        version: 3
`},
			},
			Expected: nil,
		},
		"jwt_subjects_sorted_ok": {
			KubeSecret: []*testFile{
				{
					Path: "dev/app.yml",
					Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:doctolib:project:project1:workspace:workspace1
      - organization:doctolib:project:project2:workspace:workspace2
      - organization:doctolib:project:project3:workspace:workspace3
    secrets:
      dev-aws-de-fra-1/test/path:
        keys:
          - NEWRELIC_EVENT_API_ACCOUNT_ID
        version: 3
`},
			},
			Expected: nil,
		},
		"jwt_subjects_not_sorted": {
			KubeSecret: []*testFile{
				{
					Path: "dev/app.yml",
					Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:doctolib:project:project2:workspace:workspace2
      - organization:doctolib:project:project1:workspace:workspace1
    secrets:
      dev-aws-de-fra-1/test/path:
        keys:
          - NEWRELIC_EVENT_API_ACCOUNT_ID
        version: 3
`},
			},
			Expected: errTfeJwtSubjectsCheck,
		},
		"jwt_subjects_invalid_format": {
			KubeSecret: []*testFile{
				{
					Path: "dev/app.yml",
					Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - invalid-format-subject
    secrets:
      dev-aws-de-fra-1/test/path:
        keys:
          - NEWRELIC_EVENT_API_ACCOUNT_ID
        version: 3
`},
			},
			Expected: errTfeJwtSubjectsCheck,
		},
		"jwt_subjects_empty_ok": {
			KubeSecret: []*testFile{
				{
					Path: "dev/app.yml",
					Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    secrets:
      dev-aws-de-fra-1/test/path:
        keys:
          - NEWRELIC_EVENT_API_ACCOUNT_ID
        version: 3
`},
			},
			Expected: nil,
		},
		"jwt_subjects_with_wildcards": {
			KubeSecret: []*testFile{
				{
					Path: "dev/app.yml",
					Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:*:project:*:workspace:*
      - organization:doctolib:project:*:workspace:workspace1
    secrets:
      dev-aws-de-fra-1/test/path:
        keys:
          - NEWRELIC_EVENT_API_ACCOUNT_ID
        version: 3
`},
			},
			Expected: nil,
		},
		"jwt_subjects_missing_colon": {
			KubeSecret: []*testFile{
				{
					Path: "dev/app.yml",
					Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization:doctolib:project:project1workspace:workspace1
    secrets:
      dev-aws-de-fra-1/test/path:
        keys:
          - NEWRELIC_EVENT_API_ACCOUNT_ID
        version: 3
`},
			},
			Expected: errTfeJwtSubjectsCheck,
		},
		"jwt_subjects_empty_segment": {
			KubeSecret: []*testFile{
				{
					Path: "dev/app.yml",
					Content: `---
vaultSecrets:
  app:
    vaultNamespace: doctolib/dev
    vaultRole: foo
    tfeJwtSubjects:
      - organization::project:project1:workspace:workspace1
    secrets:
      dev-aws-de-fra-1/test/path:
        keys:
          - NEWRELIC_EVENT_API_ACCOUNT_ID
        version: 3
`},
			},
			Expected: errTfeJwtSubjectsCheck,
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "TestCheckTfeJwtSubjects")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)

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

			result, err := checkTfeJwtSubjects(dir + "/configs/vault-secrets/*/*.yml")

			if expectedError, ok := v.Expected.(error); ok {
				assert.ErrorIs(t, err, expectedError, "there should be an error")
				assert.GreaterOrEqual(t, len(result), 1, "len(result) should be at least 1")
			} else {
				assert.NoError(t, err, "there should be no error")
				assert.Equal(t, 0, len(result), "len(result) should be equals to 0")
			}
		})
	}
}
