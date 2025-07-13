package secret

import (
	"bytes"
	"strings"
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

func TestE2eDiff(t *testing.T) {
	var testScenarios = map[string]secretE2eTest{
		"changes_in_key_and_value": {
			Initial: []helper.Secret{
				{
					Path: "dev-aws-de-fra-1/test",
					Data: map[string]interface{}{"foo": "bar"},
				},
				{
					Path: "dev-aws-de-fra-1/test",
					Data: map[string]interface{}{"bar": "baz"},
				},
			},
			Args:     []string{"diff", "--platform", "dev", "--environment", "de", "--path", "test", "--base-version", "1", "--diff-version", "0"},
			Expected: "--- base-version\n+++ diff-version\n@@ -1 +1 @@\n-foo: bar\n+bar: baz\n",
		},
		"changes_in_key": {
			Initial: []helper.Secret{
				{
					Path: "dev-aws-de-fra-1/test",
					Data: map[string]interface{}{"foo": "bar"},
				},
				{
					Path: "dev-aws-de-fra-1/test",
					Data: map[string]interface{}{"bar": "bar"},
				},
			},
			Args:     []string{"diff", "--platform", "dev", "--environment", "de", "--path", "test", "--base-version", "1", "--diff-version", "0"},
			Expected: "--- base-version\n+++ diff-version\n@@ -1 +1 @@\n-foo: bar\n+bar: bar\n",
		},
		"changes_in_value": {
			Initial: []helper.Secret{
				{
					Path: "dev-aws-de-fra-1/test",
					Data: map[string]interface{}{"foo": "bar"},
				},
				{
					Path: "dev-aws-de-fra-1/test",
					Data: map[string]interface{}{"foo": "baz"},
				},
			},
			Args:     []string{"diff", "--platform", "dev", "--environment", "de", "--path", "test", "--base-version", "1", "--diff-version", "0"},
			Expected: "--- base-version\n+++ diff-version\n@@ -1 +1 @@\n-foo: bar\n+foo: baz\n",
		},
		"no_diff_between_versions": {
			Initial: []helper.Secret{
				{
					Path: "dev-aws-de-fra-1/test",
					Data: map[string]interface{}{"foo": "bar"},
				},
				{
					Path: "dev-aws-de-fra-1/test",
					Data: map[string]interface{}{"foo": "baz"},
				},
				{
					Path: "dev-aws-de-fra-1/test",
					Data: map[string]interface{}{"foo": "bar"},
				},
			},
			Args:     []string{"diff", "--platform", "dev", "--environment", "de", "--path", "test", "--base-version", "1", "--diff-version", "0"},
			Expected: "",
		},
		"same_version_given_using_latest_version": {
			Initial: []helper.Secret{
				{
					Path: "dev-aws-de-fra-1/test",
					Data: map[string]interface{}{"foo": "bar"},
				},
				{
					Path: "dev-aws-de-fra-1/test",
					Data: map[string]interface{}{"foo": "baz"},
				},
				{
					Path: "dev-aws-de-fra-1/test",
					Data: map[string]interface{}{"foo": "bar"},
				},
			},
			Args: []string{"diff", "--platform", "dev", "--environment", "de", "--path", "test", "--base-version", "3", "--diff-version", "0"},
			Expected: helper.SecretExpectedError{
				Name:  "errSameVersionDiff",
				Error: errSameVersionDiff,
			},
		},
		"same_base_and_diff_versions": {
			Initial: []helper.Secret{
				{
					Path: "dev-aws-de-fra-1/test",
					Data: map[string]interface{}{"foo": "bar"},
				},
			},
			Args: []string{"diff", "--platform", "dev", "--environment", "de", "--path", "test", "--base-version", "0", "--diff-version", "0"},
			Expected: helper.SecretExpectedError{
				Name:  "errSameVersionDiff",
				Error: errSameVersionDiff,
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
			if expectedOutput, ok := v.Expected.(string); ok {
				assert.NoError(t, err)
				assert.Empty(t, stderr, "stderr should be empty")
				assert.NotEmpty(t, stdout, "stdout should not be empty")

				result, err := compareSecrets([]*api.Client{client}, "kv/data/"+v.Initial[0].Path, 1, len(v.Initial))
				var resultDiff bytes.Buffer
				resultDiff.WriteString(result)
				if assert.NoError(t, err, nil, "error should be nil") {
					assert.Equal(t, expectedOutput, resultDiff.String(), "diff string should be equal")
				}
			} else if expectedError, ok := v.Expected.(helper.SecretExpectedError); ok {
				assert.ErrorIsf(t, err, expectedError.Error, "error should be %s", expectedError.Name)
			}
		})
	}
}
