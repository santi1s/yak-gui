package secret

import (
	"strconv"
	"strings"
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"

	"github.com/stretchr/testify/assert"
)

func TestDeleteSecret(t *testing.T) {
	var testScenarios = map[string]secretUnitTest{
		"with_version": {
			Version: 1,
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Version: 1, Data: nil},
		},
		"without_version": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Version: 1, Data: nil},
		},
		"non_existent_version": {
			Version: 2,
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Version: 1, Data: map[string]interface{}{"mykey": "myvalue"}},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			cluster, client := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			if expectedSecret, ok := v.Expected.(helper.Secret); ok {
				// if we expect a secret to be returned
				err := deleteSecret([]*api.Client{client}, expectedSecret.Path, v.Version)
				if assert.Nil(t, err, "we should not get any error while deleting a secret version") {
					data := map[string][]string{
						"version": {strconv.Itoa(expectedSecret.Version)},
					}

					secret, err := client.Logical().ReadWithData("kv/data/"+expectedSecret.Path, data)
					assert.Nil(t, err, "we should be able to read the deleted secret version but got an error")
					assert.NotNil(t, secret, "we should get a secret even if the version has been deleted")
					if expectedSecret.Data == nil {
						assert.Empty(t, secret.Data["data"], "no secret data should be returned as the secret should be deleted")
						assert.NotEmpty(t, secret.Data["metadata"].(map[string]interface{})["deletion_time"], "we should get a deletion time on a deleted secret version")
					}

					secret, err = client.Logical().ReadWithData("kv/data/ci/"+expectedSecret.Path, data)
					assert.Nil(t, err, "we should be able to read the deleted secret version but got an error")
					assert.NotNil(t, secret, "we should get a secret even if the version has been deleted")
					if expectedSecret.Data == nil {
						assert.Empty(t, secret.Data["data"], "no secret data should be returned as the secret should be deleted")
						assert.NotEmpty(t, secret.Data["metadata"].(map[string]interface{})["deletion_time"], "we should get a deletion time on a deleted secret version")
					}
				}
			} else if expectedError, ok := v.Expected.(helper.SecretExpectedError); ok {
				// if we expect an error to be returned
				_ = deleteSecret([]*api.Client{client}, expectedError.Path, v.Version)
				// assert.ErrorIsf(t, err, expectedError.Error, "error should be %s", expectedError.Name)
				// assert.Nil(t, secret, "secret should be nil")
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}

func TestE2eDelete(t *testing.T) {
	var testScenarios = map[string]secretE2eTest{
		"delete_latest": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Input:    "y\n",
			Args:     []string{"delete", "--platform", "dev", "--environment", "de", "--path", "test"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test"},
		},
		"delete_version": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Input:    "y\n",
			Args:     []string{"delete", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "4"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test"},
		},
	}
	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			initE2ETest()
			cluster, client := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			r := strings.NewReader(v.Input)
			stdout, stderr, err := helper.ExecuteCobraCommand(secretCmd, v.Args, r)

			if expectedSecret, ok := v.Expected.(helper.Secret); ok {
				// if we expect a secret to be returned
				assert.NoError(t, err)
				assert.Empty(t, stderr, "stderr should be empty")
				// TODO: parse output to check data instead of just checking strings
				assert.Contains(t, stdout, "secret version has been deleted", "stdout should contain 'secret version has been deleted'")
				secret, err := client.Logical().Read("kv/data/" + expectedSecret.Path)
				if assert.NoError(t, err) {
					assert.Nil(t, secret.Data["Data"])
				}
			} else if expectedError, ok := v.Expected.(helper.SecretExpectedError); ok {
				// if we expect an error to be returned
				assert.ErrorIsf(t, err, expectedError.Error, "error should be %s", expectedError.Name)
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}
