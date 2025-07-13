package secret

import (
	"strings"
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"

	"github.com/stretchr/testify/assert"
)

func TestPatchSecretMetadata(t *testing.T) {
	secretPath := "dev-aws-de-fra-1/test" //#nosec G101

	testScenarios := map[string]secretUnitTest{
		"add custom metadata": {
			Initial:  []helper.Secret{{Path: secretPath, Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar"}}},
			Update:   map[string]interface{}{"custom_metadata": map[string]interface{}{"bar": "baz"}},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Metadata: map[string]interface{}{"bar": "baz", "foo": "bar"}},
		},
		"update custom metadata": {
			Initial:  []helper.Secret{{Path: secretPath, Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar", "bar": "baz"}}},
			Update:   map[string]interface{}{"custom_metadata": map[string]interface{}{"foo": "baz"}},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Metadata: map[string]interface{}{"bar": "baz", "foo": "baz"}},
		},
		"remove custom metadata": {
			Initial:  []helper.Secret{{Path: secretPath, Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar", "bar": "baz"}}},
			Update:   map[string]interface{}{"custom_metadata": map[string]interface{}{"foo": nil}},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Metadata: map[string]interface{}{"bar": "baz"}},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			cluster, client := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			if expectedSecret, ok := v.Expected.(helper.Secret); ok {
				// if we expect a secret to be returned
				secret, err := PatchSecretMetadata([]*api.Client{client}, secretPath, v.Update)
				assert.NoError(t, err, nil, "error should be nil")
				assert.Nil(t, secret, "secret should be nil")

				secret, err = client.Logical().Read("kv/metadata/" + secretPath)
				assert.NoError(t, err, nil, "error should be nil")
				assert.Equal(t, expectedSecret.Metadata, secret.Data["custom_metadata"], "should be equal")
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}

func TestE2eMetadataUpdate(t *testing.T) {
	var testScenarios = map[string]secretE2eTest{
		"update": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar"}}},
			Args:     []string{"metadata", "update", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "foo", "--value", "baz"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Metadata: map[string]interface{}{"foo": "baz"}},
		},
		"update_non_existing_path": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar"}}},
			Args:    []string{"metadata", "update", "--platform", "dev", "--environment", "de", "--path", "non_existent", "--key", "foo", "--value", "baz"},
			Expected: helper.SecretExpectedError{
				Name:  "ErrSecretNotFound",
				Error: helper.ErrSecretNotFound,
			},
		},
		"update_key_not_found": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar"}}},
			Args:    []string{"metadata", "update", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "bar", "--value", "baz"},
			Expected: helper.SecretExpectedError{
				Name:  "errMetadataKeyNotFound",
				Error: errMetadataKeyNotFound,
			},
		},
		"update_key_empty": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar"}}},
			Args:    []string{"metadata", "update", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "", "--value", "baz"},
			Expected: helper.SecretExpectedError{
				Name:  "errKeyParameterCantBeEmpty",
				Error: errKeyParameterCantBeEmpty,
			},
		},
		"update_value_empty": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar"}}},
			Args:    []string{"metadata", "update", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "foo", "--value", ""},
			Expected: helper.SecretExpectedError{
				Name:  "errValueParameterCantBeEmpty",
				Error: errValueParameterCantBeEmpty,
			},
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
				assert.Contains(t, stdout, "metadata foo has been updated", "stdout should contain 'metadata foo has been updated'")
				secret, err := client.Logical().Read("kv/metadata/" + expectedSecret.Path)
				if assert.NoError(t, err, nil, "error should be nil") {
					assert.Equal(t, expectedSecret.Metadata, secret.Data["custom_metadata"], "data should be equal")
				}
			} else if expectedError, ok := v.Expected.(helper.SecretExpectedError); ok {
				assert.ErrorIsf(t, err, expectedError.Error, "error should be %s", expectedError.Name)
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}
