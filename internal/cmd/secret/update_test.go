package secret

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"

	"github.com/stretchr/testify/assert"
)

func TestPatchSecretData(t *testing.T) {
	var testScenarios = map[string]secretUnitTest{
		"add_key": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}},
			},
			Update:   map[string]interface{}{"bar": "baz"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"bar": "baz", "foo": "bar"}, Version: 2},
		},
		"update_key": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}},
			},
			Update:   map[string]interface{}{"foo": "baz"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "baz"}, Version: 2},
		},
		"remove_key": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}},
			},
			Update:   map[string]interface{}{"foo": nil},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{}, Version: 2},
		},
		"update_when_latest_version_is_deleted": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"bar": ""}, Deleted: true},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"bar": ""}, Deleted: true},
			},
			Update:   map[string]interface{}{"baz": "foo"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"baz": "foo", "foo": "bar"}, Version: 3},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			cluster, client := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			if expectedSecret, ok := v.Expected.(helper.Secret); ok {
				// if we expect a secret to be returned
				secret, err := PatchSecretData([]*api.Client{client}, expectedSecret.Path, v.Update)
				if assert.NoError(t, err, "error should be nil") {
					actualVersion, err := secret.Data["version"].(json.Number).Int64()
					if assert.NoError(t, err, "error should be nil") {
						assert.Equal(t, expectedSecret.Version, int(actualVersion), "values should be equal")
					}
				}
				secret2, err := client.Logical().ReadWithData("kv/data/"+expectedSecret.Path, map[string][]string{
					"version": {strconv.Itoa(expectedSecret.Version)},
				})
				if assert.NoError(t, err, nil, "error should be nil") {
					assert.Equal(t, expectedSecret.Data, secret2.Data["data"], "data should be equal")
				}
				secret2, err = client.Logical().ReadWithData("kv/data/ci/"+expectedSecret.Path, map[string][]string{
					"version": {strconv.Itoa(expectedSecret.Version)},
				})
				if assert.NoError(t, err, nil, "error should be nil") {
					data := make(map[string]interface{})
					for k := range expectedSecret.Data {
						data[k] = ""
					}
					assert.Equal(t, data, secret2.Data["data"], "data should be equal")
				}
			} else if expectedError, ok := v.Expected.(helper.SecretExpectedError); ok {
				// if we expect an error to be returned
				t.Log(expectedError.Name)
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}

func TestE2eUpdate(t *testing.T) {
	var testScenariosE2eUpdate = map[string]secretE2eTest{
		"interactive": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Input:    "mykey\nmy_new_value\nn",
			Args:     []string{"update", "--platform", "dev", "--environment", "de", "--path", "test", "--interactive"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "my_new_value"}},
		},
		"not_interactive": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Input:    "mykey: my_new_value",
			Args:     []string{"update", "--platform", "dev", "--environment", "de", "--path", "test"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "my_new_value"}},
		},
		"update_non_existing": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Input:   "mykey\nmy_new_value\nn",
			Args:    []string{"update", "--platform", "dev", "--environment", "de", "--path", "non_existent", "--interactive"},
			Expected: helper.SecretExpectedError{
				Name:  "ErrSecretNotFound",
				Error: helper.ErrSecretNotFound,
			},
		},
		"update_non_existing_not_interactive": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Input:   "mykey: my_new_value",
			Args:    []string{"update", "--platform", "dev", "--environment", "de", "--path", "non_existent"},
			Expected: helper.SecretExpectedError{
				Name:  "ErrSecretNotFound",
				Error: helper.ErrSecretNotFound,
			},
		},
		"empty_input": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Input:   "",
			Args:    []string{"update", "--platform", "dev", "--environment", "de", "--path", "test"},
			Expected: helper.SecretExpectedError{
				Name:  "errEmptyInput",
				Error: errEmptyInput,
			},
		},
		"insufficient_input": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}}},
			Input:   "mykey:",
			Args:    []string{"update", "--platform", "dev", "--environment", "de", "--path", "test"},
			Expected: helper.SecretExpectedError{
				Name:  "errCantContainEmptyValue",
				Error: errCantContainEmptyValue,
			},
		},
		"remove_key": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue", "mykey2": "myvalue2"}}},
			Input:    "",
			Args:     []string{"update", "--platform", "dev", "--environment", "de", "--path", "test", "--remove", "--keys", "mykey"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey2": "myvalue2"}},
		},
		"remove_multiple_keys": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue", "mykey2": "myvalue2", "mykey3": "myvalue3"}}},
			Input:    "",
			Args:     []string{"update", "--platform", "dev", "--environment", "de", "--path", "test", "--remove", "--keys", "mykey,mykey2"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey3": "myvalue3"}},
		},
		"interactive_remove_key": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue", "mykey2": "myvalue2"}}},
			Input:    "mykey\nn",
			Args:     []string{"update", "--platform", "dev", "--environment", "de", "--path", "test", "--interactive", "--remove"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey2": "myvalue2"}},
		},
		"interactive_remove_multiple_keys": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue", "mykey2": "myvalue2", "mykey3": "myvalue3"}}},
			Input:    "mykey\ny\nmykey2\nn",
			Args:     []string{"update", "--platform", "dev", "--environment", "de", "--path", "test", "--interactive", "--remove"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey3": "myvalue3"}},
		},
		"stdin_remove_key": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue", "mykey2": "myvalue2"}}},
			Input:    "mykey:",
			Args:     []string{"update", "--platform", "dev", "--environment", "de", "--path", "test", "--remove"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey2": "myvalue2"}},
		},
		"stdin_remove_multiple_keys": {
			Initial:  []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue", "mykey2": "myvalue2", "mykey3": "myvalue3"}}},
			Input:    "mykey:\nmykey2:",
			Args:     []string{"update", "--platform", "dev", "--environment", "de", "--path", "test", "--remove"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey3": "myvalue3"}},
		},
		"stdin_bad_input": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue", "mykey2": "myvalue2", "mykey3": "myvalue3"}}},
			Input:   "mykey: value\nmykey2:",
			Args:    []string{"update", "--platform", "dev", "--environment", "de", "--path", "test", "--remove"},
			Expected: helper.SecretExpectedError{
				Name:  "errCantContainValue",
				Error: errCantContainValue,
			},
		},
	}

	for k, v := range testScenariosE2eUpdate {
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
				assert.Contains(t, stdout, "version: 2", "stdout should contain 'version: 2'")
				secret, err := client.Logical().Read("kv/data/" + expectedSecret.Path)
				if assert.NoError(t, err, nil, "error should be nil") {
					assert.Equal(t, expectedSecret.Data, secret.Data["data"], "data should be equal")
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
