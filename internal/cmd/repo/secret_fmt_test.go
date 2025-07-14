package repo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santi1s/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestSecretFmt(t *testing.T) {
	var testScenarios = map[string]kubeSecretE2eTest{
		"fmt_ok": {
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
			Args: []string{"secret", "fmt"},
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
		"fmt_not_ok": {
			KubeSecret: []*testFile{{
				Path: "dev/app.yml",
				Content: `vaultSecrets:
  app:
    secrets:
      dev-aws-de-fra-1/test:
        keys:
          - foo
          - bar
        version: "1"
      dev-aws-de-fra-1/test2:
        keys: []
        version: latest
      dev-aws-de-fra-1/test3:
        keys: []
        version: "latest"
      dev-aws-de-fra-1/foo:
        keys:
          - FOO
        version: 3
    vaultRole: foo
    vaultNamespace: doctolib/dev
`}},
			Args: []string{"secret", "fmt"},
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
          - foo
        version: 1
      dev-aws-de-fra-1/test2:
        keys: []
        version: latest
      dev-aws-de-fra-1/test3:
        keys: []
        version: latest
`},
		},
		"fmt_check_ok": {
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
			Args:     []string{"secret", "fmt", "--check"},
			Expected: helper.SecretExpectedError{Error: nil},
		},
		"fmt_check_not_ok": {
			KubeSecret: []*testFile{{
				Path: "dev/app.yml",
				Content: `---
vaultSecrets:
  app:
    vaultRole: foo
    secrets:
      dev-aws-de-fra-1/foo:
        keys:
          - FOO
          - BAR
        version: 3
      dev-aws-de-fra-1/test:
        keys:
          - foo
        version: 1
    vaultNamespace: doctolib/dev
`}},
			Args:     []string{"secret", "fmt", "--check"},
			Expected: helper.SecretExpectedError{Name: "errFilesNotCorrectlyFormatted", Error: errFilesNotCorrectlyFormatted},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			initE2ETest()
			if v.VaultSecrets != nil {
				cluster, _ := setupVaultCluster(t, v.VaultSecrets...)
				defer cluster.Cleanup()
			}

			dir, err := os.MkdirTemp("", "TestSecretFmt")
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
				assert.ErrorIs(t, err, expectedError.Error, "error should match")
			} else {
				t.Log("expected value is of unsupported type")
			}
		})
	}
}
