package secret

import (
	"strconv"
	"strings"
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

func TestE2eSecretResync(t *testing.T) {
	var testScenarios = map[string]secretE2eTest{
		"existing_ci_version": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}},
			},
			Args:     []string{"resync", "--platform", "dev", "--environment", "de", "--path", "test"},
			Expected: "Nothing to do",
		},
		"existing_ci_version_with_version_specified": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}},
			},
			Args:     []string{"resync", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "1"},
			Expected: "Nothing to do",
		},
		"no_ci_secret": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
			},
			Args:     []string{"resync", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "1"},
			Expected: helper.Secret{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}, Version: 1},
		},
		"no_ci_version": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar", "baz": "bar2"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}},
			},
			Args:     []string{"resync", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "2"},
			Expected: helper.Secret{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "", "baz": ""}, Version: 2},
		},
		"no_ci_version_too_far": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar", "baz": "bar2"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar1", "baz": "bar2"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}},
			},
			Args:     []string{"resync", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "2"},
			Expected: helper.SecretExpectedError{Contains: "unexpected version"},
		},
		"existing_ci_version_both_deleted": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}, Deleted: true},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}, Deleted: true},
			},
			Args:     []string{"resync", "--platform", "dev", "--environment", "de", "--path", "test"},
			Expected: "Nothing to do",
		},
		"no_ci_version_base_deleted": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar", "baz": "bar2"}, Deleted: true},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}},
			},
			Args:     []string{"resync", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "2"},
			Expected: helper.Secret{Path: "ci/dev-aws-de-fra-1/test", Version: 2, Deleted: true},
		},
		"ci_version_base_deleted": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar", "baz": "bar2"}, Deleted: true},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "", "baz": ""}},
			},
			Args:     []string{"resync", "--platform", "dev", "--environment", "de", "--path", "test", "--version", "2"},
			Expected: helper.Secret{Path: "ci/dev-aws-de-fra-1/test", Version: 2, Deleted: true},
		},
		"ci_version_deleted": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}, Deleted: true},
			},
			Args:     []string{"resync", "--platform", "dev", "--environment", "de", "--path", "test"},
			Expected: helper.Secret{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": ""}, Version: 1, Deleted: false},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			initE2ETest()
			cluster, _ := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			r := strings.NewReader(v.Input)
			stdout, stderr, err := helper.ExecuteCobraCommand(secretCmd, v.Args, r)

			if v.Expected == nil {
				assert.Nil(t, err, "error should be nil")
			} else if expectedString, ok := v.Expected.(string); ok {
				assert.Nil(t, err, "error should be nil")
				assert.Contains(t, stdout, expectedString)
			} else if expectedSecret, ok := v.Expected.(helper.Secret); ok {
				// if we expect a secret to be returned
				assert.NoError(t, err)
				assert.Empty(t, stderr, "stderr should be empty")
				assert.Contains(t, stdout, "CI secret succesfully resynchronized")

				var secret *api.Secret
				if expectedSecret.Version == 0 {
					secret, err = client.Logical().Read("kv/data/" + expectedSecret.Path)
				} else {
					data := map[string][]string{
						"version": {strconv.Itoa(expectedSecret.Version)},
					}
					secret, err = client.Logical().ReadWithData("kv/data/"+expectedSecret.Path, data)
				}

				if expectedSecret.Data != nil {
					if assert.NoError(t, err, nil, "error should be nil") {
						assert.Equal(t, expectedSecret.Data, secret.Data["data"], "data should be equal")
					}
				}
				if expectedSecret.Deleted == true {
					assert.Empty(t, secret.Data["data"], "no secret data should be returned as the secret should be deleted")
					assert.NotEmpty(t, secret.Data["metadata"].(map[string]interface{})["deletion_time"], "we should get a deletion time on a deleted secret version")
				} else {
					if assert.NoError(t, err, nil, "error should be nil") {
						assert.Equal(t, expectedSecret.Data, secret.Data["data"], "data should be equal")
					}
					assert.Empty(t, secret.Data["metadata"].(map[string]interface{})["deletion_time"], "we should not get a deletion time on a non deleted secret version")
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
