package secret

import (
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"

	"github.com/stretchr/testify/assert"
)

func TestReadSecretMetadata(t *testing.T) {
	var testScenarios = map[string]secretUnitTest{
		"nominal": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "sre"}}},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "sre"}},
		},
		"non_existent_path": {
			Initial:  []helper.Secret{},
			Expected: helper.SecretExpectedError{Path: "dev-aws-de-fra-1/test", Name: "ErrSecretNotFound", Error: helper.ErrSecretNotFound},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			cluster, client := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			if expectedSecret, ok := v.Expected.(helper.Secret); ok {
				// if we expect a secret to be returned
				secret, err := ReadSecretMetadata([]*api.Client{client}, expectedSecret.Path)
				if assert.NoError(t, err, nil, "error should be nil") {
					assert.NotNil(t, secret.Data, "secret should not be nil")
					assert.Equal(t, expectedSecret.Metadata, secret.Data["custom_metadata"].(map[string]interface{}), "values should be equal")
				}
			} else if expectedError, ok := v.Expected.(helper.SecretExpectedError); ok {
				// if we expect an error to be returned
				secret, err := ReadSecretMetadata([]*api.Client{client}, expectedError.Path)
				assert.ErrorIsf(t, err, expectedError.Error, "error should be %s", expectedError.Name)
				assert.Nil(t, secret, "secret should be nil")
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}

func TestE2eMetadataGet(t *testing.T) {
	var testScenarios = map[string]secretE2eTest{
		"nominal": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "sre"}}},
			Args:     []string{"get", "--platform", "dev", "--environment", "de", "--path", "test"},
			Expected: helper.Secret{Data: map[string]interface{}{"owner": "sre"}},
		},
		"non_existent_path": {
			Args: []string{"get", "--platform", "dev", "--environment", "de", "--path", "foo"},
			Expected: helper.SecretExpectedError{
				Name:  "ErrSecretNotFound",
				Error: helper.ErrSecretNotFound,
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
				for kw, vw := range expectedSecret.Metadata {
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
