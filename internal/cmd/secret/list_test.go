package secret

import (
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"

	"github.com/stretchr/testify/assert"
)

func TestListSecret(t *testing.T) {
	var testScenarios = map[string]secretUnitTest{
		"existing_path": {
			Path: "dev-aws-de-fra-1",
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/foo", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "dev-aws-de-fra-1/bar", Data: map[string]interface{}{"mykey": "myvalue"}},
			},
			Expected: map[string]interface{}{"keys": []interface{}{"bar", "foo"}},
		},
		"non_existent_path": {
			Path: "common",
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/foo", Data: map[string]interface{}{"mykey": "myvalue"}},
			},
			Expected: helper.SecretExpectedError{Name: "helper.ErrSecretPathNotFound", Error: helper.ErrSecretPathNotFound},
		},
		"path_is_a_secret": {
			Path: "dev-aws-de-fra-1/foo",
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/foo", Data: map[string]interface{}{"mykey": "myvalue"}},
			},
			Expected: helper.SecretExpectedError{Name: "helper.ErrListSecretNotSecretPath", Error: helper.ErrListSecretNotSecretPath},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			cluster, client := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			if expectedData, ok := v.Expected.(map[string]interface{}); ok {
				// if we expect a secret to be returned
				secret, err := listSecret([]*api.Client{client}, v.Path)
				assert.Nil(t, err, nil, "error should be nil")
				assert.NotNil(t, secret.Data, "secret should not be nil")
				assert.Equal(t, expectedData, secret.Data, "data should be equal")
			} else if expectedError, ok := v.Expected.(helper.SecretExpectedError); ok {
				// if we expect an error to be returned
				secret, err := listSecret([]*api.Client{client}, v.Path)
				assert.ErrorIsf(t, err, expectedError.Error, "error should be %s", expectedError.Name)
				assert.Nil(t, secret, "secret should be nil")
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}

func TestE2eList(t *testing.T) {
	var testScenarios = map[string]secretE2eTest{
		"nominal": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test/foo", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "dev-aws-de-fra-1/test/bar", Data: map[string]interface{}{"mykey": "myvalue"}},
			},
			Args:     []string{"list", "--platform", "dev", "--environment", "de", "--path", "test"},
			Expected: "---\nkeys:\n- bar\n- foo\n\n",
		},
		"nominal_with_slash_suffix": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test/foo", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "dev-aws-de-fra-1/test/bar", Data: map[string]interface{}{"mykey": "myvalue"}},
			},
			Args:     []string{"list", "--platform", "dev", "--environment", "de", "--path", "test/"},
			Expected: "---\nkeys:\n- bar\n- foo\n\n",
		},
		"non_existent_path": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test/foo", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Args:     []string{"list", "--platform", "dev", "--environment", "de", "--path", "test2"},
			Expected: helper.SecretExpectedError{Name: "ErrSecretPathNotFound", Error: helper.ErrSecretPathNotFound},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			initE2ETest()
			cluster, _ := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			stdout, stderr, err := helper.ExecuteCobraCommand(secretCmd, v.Args)

			if expectedData, ok := v.Expected.(string); ok {
				// if we expect a secret to be returned
				assert.NoError(t, err)
				assert.Empty(t, stderr, "stderr should be empty")
				assert.Containsf(t, stdout, expectedData, "stdout should contain '%s'", stdout)
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
