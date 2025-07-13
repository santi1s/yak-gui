package secret

import (
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"

	"github.com/stretchr/testify/assert"
)

func TestReadSecret(t *testing.T) {
	var testScenarios = map[string]secretUnitTest{
		"without_version": {
			Version:  0,
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
		},
		"with_version": {
			Version:  1,
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
		},
		"non_existent_path": {
			Version:  1,
			Initial:  []helper.Secret{},
			Expected: helper.SecretExpectedError{Path: "dev-aws-de-fra-1/test", Name: "helper.ErrSecretNotFound", Error: helper.ErrSecretNotFound},
		},
		"non_existent_version": {
			Version:  2,
			Initial:  []helper.Secret{},
			Expected: helper.SecretExpectedError{Path: "dev-aws-de-fra-1/test", Name: "helper.ErrSecretNotFound", Error: helper.ErrSecretNotFound},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			cluster, client := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			if expectedSecret, ok := v.Expected.(helper.Secret); ok {
				// if we expect a secret to be returned
				secret, err := ReadSecretData([]*api.Client{client}, expectedSecret.Path, v.Version)
				assert.Nil(t, err, nil, "error should be nil")
				assert.NotNil(t, secret.Data, "secret should not be nil")
				assert.Equal(t, secret.Data["data"].(map[string]interface{})["mykey"].(string), "myvalue", "values should be equal")
				assert.Equal(t, expectedSecret.Data, secret.Data["data"], "data should be equal")
			} else if expectedError, ok := v.Expected.(helper.SecretExpectedError); ok {
				// if we expect an error to be returned
				secret, err := ReadSecretData([]*api.Client{client}, expectedError.Path, v.Version)
				assert.ErrorIsf(t, err, expectedError.Error, "error should be %s", expectedError.Name)
				assert.Nil(t, secret, "secret should be nil")
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}

func TestE2eGet(t *testing.T) {
	var testScenarios = map[string]secretE2eTest{
		"nominal": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Args:     []string{"get", "--platform", "dev", "--environment", "de", "--path", "test"},
			Expected: helper.Secret{Data: map[string]interface{}{"mykey": "myvalue"}},
		},
		"nominal_with_existing_key": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Args:     []string{"get", "--platform", "dev", "--environment", "de", "--path", "test", "--data-key", "mykey"},
			Expected: helper.Secret{Data: map[string]interface{}{"mykey": "myvalue"}},
		},
		"nominal_with_existing_multiple_keys": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey1": "myvalue1", "mykey2": "myvalue2", "otherkey": "othervalue"}}},
			Args:     []string{"get", "--platform", "dev", "--environment", "de", "--path", "test", "--data-key", "mykey"},
			Expected: helper.Secret{Data: map[string]interface{}{"mykey1": "myvalue1", "mykey2": "myvalue2"}},
		},
		"non_existent_path": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Args:    []string{"get", "--platform", "dev", "--environment", "de", "--path", "toto"},
			Expected: helper.SecretExpectedError{
				Name:  "ErrSecretNotFound",
				Error: helper.ErrSecretNotFound,
			},
		},
		"non_existent_key": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Args:    []string{"get", "--platform", "dev", "--environment", "de", "--path", "test", "--data-key", "otherkey"},
			Expected: helper.SecretExpectedError{
				Name:  "ErrSecretDataKeyNotFound",
				Error: helper.ErrSecretDataKeyNotFound,
			},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			initE2ETest()
			cluster, _ := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			stdout, stderr, err := helper.ExecuteCobraCommand(secretCmd, v.Args)

			if expectedSecret, ok := v.Expected.(helper.Secret); ok {
				// if we expect a secret to be returned
				assert.NoError(t, err)
				assert.Empty(t, stderr, "stderr should be empty")
				// TODO: parse output to check data instead of just checking strings
				for kw, vw := range expectedSecret.Data {
					assert.Containsf(t, stdout, kw, "stdout should contain '%s'", kw)
					assert.Containsf(t, stdout, vw, "stdout should contain '%s'", vw)
				}
			} else if expectedError, ok := v.Expected.(helper.SecretExpectedError); ok {
				// if we expect an error to be returned
				assert.ErrorIsf(t, err, expectedError.Error, "error should be %s", expectedError.Name)
				assert.Empty(t, stdout, "stdout should be empty")
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}
