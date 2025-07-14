package secret

import (
	"strings"
	"testing"

	"github.com/santi1s/yak/internal/helper"

	"github.com/stretchr/testify/assert"
)

func TestE2eMetadataDelete(t *testing.T) {
	var testScenarios = map[string]secretE2eTest{
		"delete": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar", "bar": "baz"}}},
			Input:    "y\n",
			Args:     []string{"metadata", "delete", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "foo"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Metadata: map[string]interface{}{"bar": "baz"}},
		},
		"delete_not_confirmed": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar"}}},
			Input:   "n\n",
			Args:    []string{"metadata", "delete", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "foo"},
			Expected: helper.SecretExpectedError{
				Name:  "errAskConfirmationNotConfirmed",
				Error: errAskConfirmationNotConfirmed,
			},
		},
		"delete_non_existing_path": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar"}}},
			Input:   "mykey\nmy_new_value\nn",
			Args:    []string{"metadata", "delete", "--platform", "dev", "--environment", "de", "--path", "non_existent", "--key", "foo"},
			Expected: helper.SecretExpectedError{
				Name:  "helper.ErrSecretNotFound",
				Error: helper.ErrSecretNotFound,
			},
		},
		"delete_key_not_found": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar"}}},
			Args:    []string{"metadata", "delete", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "bar"},
			Expected: helper.SecretExpectedError{
				Name:  "errMetadataKeyNotFound",
				Error: errMetadataKeyNotFound,
			},
		},
		"delete_key_empty": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"foo": "bar"}}},
			Args:    []string{"metadata", "delete", "--platform", "dev", "--environment", "de", "--path", "test", "--key", ""},
			Expected: helper.SecretExpectedError{
				Name:  "errKeyParameterCantBeEmpty",
				Error: errKeyParameterCantBeEmpty,
			},
		},
		"delete_metadata_empty": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Args:    []string{"metadata", "delete", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "foo"},
			Expected: helper.SecretExpectedError{
				Name:  "errMetadataEmpty",
				Error: errMetadataEmpty,
			},
		},
		"delete_metadata_owner": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"owner": "foo"}}},
			Args:    []string{"metadata", "delete", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "owner"},
			Expected: helper.SecretExpectedError{
				Name:  "errMetadataNotDeletable",
				Error: errMetadataNotDeletable,
			},
		},
		"delete_metadata_source": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"source": "foo"}}},
			Args:    []string{"metadata", "delete", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "source"},
			Expected: helper.SecretExpectedError{
				Name:  "errMetadataNotDeletable",
				Error: errMetadataNotDeletable,
			},
		},
		"delete_metadata_usage": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"usage": "foo"}}},
			Args:    []string{"metadata", "delete", "--platform", "dev", "--environment", "de", "--path", "test", "--key", "usage"},
			Expected: helper.SecretExpectedError{
				Name:  "errMetadataNotDeletable",
				Error: errMetadataNotDeletable,
			},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			initE2ETest()
			cluster, client := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			r := strings.NewReader(v.Input)
			_, stderr, err := helper.ExecuteCobraCommand(secretCmd, v.Args, r)

			if expectedSecret, ok := v.Expected.(helper.Secret); ok {
				// if we expect a secret to be returned
				assert.NoError(t, err)
				assert.Empty(t, stderr, "stderr should be empty")
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
