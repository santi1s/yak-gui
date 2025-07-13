package secret

import (
	"strings"
	"testing"

	"github.com/doctolib/yak/internal/helper"

	"github.com/stretchr/testify/assert"
)

func TestE2eMetadataCreate(t *testing.T) {
	var testScenarios = map[string]secretE2eTest{
		"create": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar"}}},
			Args:     []string{"metadata", "create", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "bar", "--value", "baz"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Metadata: map[string]interface{}{"bar": "baz", "foo": "bar"}},
		},
		"create_non_existing_path": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar"}}},
			Args:    []string{"metadata", "update", "--platform", "dev", "--environment", "de", "--path", "non_existent", "--key", "bar", "--value", "baz"},
			Expected: helper.SecretExpectedError{
				Name:  "ErrSecretNotFound",
				Error: helper.ErrSecretNotFound,
			},
		},
		"create_key_empty": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar"}}},
			Args:    []string{"metadata", "update", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "", "--value", "baz"},
			Expected: helper.SecretExpectedError{
				Name:  "errKeyParameterCantBeEmpty",
				Error: errKeyParameterCantBeEmpty,
			},
		},
		"create_value_empty": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar"}}},
			Args:    []string{"metadata", "update", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "bar", "--value", ""},
			Expected: helper.SecretExpectedError{
				Name:  "errValueParameterCantBeEmpty",
				Error: errValueParameterCantBeEmpty,
			},
		},
		"create_metadata_empty": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Args:    []string{"metadata", "create", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "foo", "--value", "bar"},
			Expected: helper.SecretExpectedError{
				Name:  "errMetadataEmpty",
				Error: errMetadataEmpty,
			},
		},
		"create_metadata_already_exists": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar"}}},
			Args:    []string{"metadata", "create", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "foo", "--value", "baz"},
			Expected: helper.SecretExpectedError{
				Name:  "errMetadataAlreadyExists",
				Error: errMetadataAlreadyExists,
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
				assert.Contains(t, stdout, "metadata bar has been created", "stdout should contain 'metadata bar has been created'")
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
