package secret

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"

	"github.com/stretchr/testify/assert"
)

func TestWriteSecretData(t *testing.T) {
	data := map[string]interface{}{
		"key": "value",
	}
	ciData := map[string]interface{}{
		"key": "",
	}

	secretPath := "dev-aws-de-fra-1/test" //#nosec G101
	cluster, client := setupVaultCluster(t)
	defer cluster.Cleanup()

	secret, err := WriteSecretData([]*api.Client{client}, secretPath, data)
	if assert.NoError(t, err, nil, "error should be nil") {
		var wantedValue json.Number = "1"
		assert.Equal(t, wantedValue, secret.Data["version"], "values should be equal")
	}

	payload := map[string][]string{
		"version": {"1"},
	}
	secret, err = client.Logical().ReadWithData("kv/data/"+secretPath, payload)
	if assert.NoError(t, err, nil, "error should be nil") {
		assert.Equal(t, data, secret.Data["data"], "data should be equal")
	}

	secret, err = client.Logical().ReadWithData("kv/data/ci/"+secretPath, payload)
	if assert.NoError(t, err, nil, "error should be nil") {
		assert.Equal(t, ciData, secret.Data["data"], "data should be equal")
	}
}

func TestWriteSecretMetadata(t *testing.T) {
	metadata := map[string]interface{}{
		"owner":  "owner",
		"source": "source",
		"usage":  "usage",
	}

	secretPath := "dev-aws-de-fra-1/test" //#nosec G101
	cluster, client := setupVaultCluster(t, helper.Secret{Path: secretPath, Data: map[string]interface{}{"key": "value"}})
	defer cluster.Cleanup()

	err := WriteSecretMetadata([]*api.Client{client}, secretPath, metadata)
	assert.NoError(t, err, nil, "error should be nil")

	wantedMetadata := metadata
	secret, err := client.Logical().Read("kv/metadata/" + secretPath)
	if assert.NoError(t, err, nil, "error should be nil") {
		assert.Equal(t, wantedMetadata, secret.Data["custom_metadata"], "metadata should be equal")
	}
}

func TestE2eCreate(t *testing.T) {
	var testScenarios = map[string]secretE2eTest{
		"interactive": {
			Initial:  nil,
			Input:    "my_key\nmy_value\nn",
			Args:     []string{"create", "--platform", "dev", "--environment", "de", "--path", "test", "--owner", "owner", "--usage", "usage", "--source", "source", "--interactive"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"my_key": "my_value"}, Metadata: map[string]interface{}{"owner": "owner", "usage": "usage", "source": "source"}},
		},
		"not_interactive": {
			Initial:  nil,
			Input:    "my_key: my_value",
			Args:     []string{"create", "--platform", "dev", "--environment", "de", "--path", "test", "--owner", "owner", "--usage", "usage", "--source", "source"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"my_key": "my_value"}, Metadata: map[string]interface{}{"owner": "owner", "usage": "usage", "source": "source"}},
		},
		"already_exists": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Input:   "mykey: myvalue",
			Args:    []string{"create", "--platform", "dev", "--environment", "de", "--path", "test", "--owner", "owner", "--usage", "usage", "--source", "source"},
			Expected: helper.SecretExpectedError{
				Name:  "errSecretAlreadyExists",
				Error: errSecretAlreadyExists,
			},
		},
		"empty_input": {
			Initial: nil,
			Input:   "",
			Args:    []string{"create", "--platform", "dev", "--environment", "de", "--path", "test", "--owner", "owner", "--usage", "usage", "--source", "source"},
			Expected: helper.SecretExpectedError{
				Name:  "errEmptyInput",
				Error: errEmptyInput,
			},
		},
		"insufficient_input": {
			Initial: nil,
			Input:   "mykey:",
			Args:    []string{"create", "--platform", "dev", "--environment", "de", "--path", "test", "--owner", "owner", "--usage", "usage", "--source", "source"},
			Expected: helper.SecretExpectedError{
				Name:  "errCantContainEmptyValue",
				Error: errCantContainEmptyValue,
			},
		},
		"missing_platform": {
			Initial: nil,
			Input:   "mykey: myvalue",
			Args:    []string{"create", "--environment", "de", "--path", "test", "--owner", "owner", "--usage", "usage", "--source", "source"},
			Expected: helper.SecretExpectedError{
				Name:  "errEnvironmentCantBeSetWithoutPlatform",
				Error: helper.ErrEnvironmentCantBeSetWithoutPlatform,
			},
		},
		"missing_env": {
			Initial:  nil,
			Input:    "my_key: my_value",
			Args:     []string{"create", "--platform", "dev", "--path", "test", "--owner", "owner", "--usage", "usage", "--source", "source"},
			Expected: helper.Secret{Path: "common/test", Data: map[string]interface{}{"my_key": "my_value"}, Metadata: map[string]interface{}{"owner": "owner", "usage": "usage", "source": "source"}},
		},
		"missing_owner_metadata": {
			Initial: nil,
			Input:   "mykey: myvalue",
			Args:    []string{"create", "--platform", "dev", "--environment", "de", "--path", "test", "--owner", "", "--source", "source", "--usage", "usage"},
			Expected: helper.SecretExpectedError{
				Name:  "errOwnerCantBeEmpty",
				Error: errOwnerCantBeEmpty,
			},
		},
		"missing_source_metadata": {
			Initial: nil,
			Input:   "mykey: myvalue",
			Args:    []string{"create", "--platform", "dev", "--environment", "de", "--path", "test", "--owner", "owner", "--source", "", "--usage", "usage"},
			Expected: helper.SecretExpectedError{
				Name:  "errSourceCantBeEmpty",
				Error: errSourceCantBeEmpty,
			},
		},
		"missing_usage_metadata": {
			Initial: nil,
			Input:   "mykey: myvalue",
			Args:    []string{"create", "--platform", "dev", "--environment", "de", "--path", "test", "--owner", "owner", "--source", "source", "--usage", ""},
			Expected: helper.SecretExpectedError{
				Name:  "errUsageCantBeEmpty",
				Error: errUsageCantBeEmpty,
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
				assert.Contains(t, stdout, "version: 1", "stdout should contain 'version: 1'")
				secret, err := client.Logical().Read("kv/data/" + expectedSecret.Path)
				if assert.NoError(t, err, nil, "error should be nil") {
					assert.Equal(t, expectedSecret.Data, secret.Data["data"], "data should be equal")
				}
				metadata, err := client.Logical().Read("kv/metadata/" + expectedSecret.Path)
				if assert.NoError(t, err, nil, "error should be nil") {
					assert.Equal(t, expectedSecret.Metadata, metadata.Data["custom_metadata"], "metadata should be equal")
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
