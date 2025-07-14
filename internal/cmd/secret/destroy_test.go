package secret

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/santi1s/yak/internal/helper"

	"github.com/stretchr/testify/assert"
)

func TestE2eDestroy(t *testing.T) {
	var testScenarios = map[string]secretE2eTest{
		"destroy_3_versions_with_version_2_deleted": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Version: 1},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}, Version: 1},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Version: 2},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}, Version: 2},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Version: 3},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}, Version: 3},
			},
			Input:    "y\ny\n",
			Args:     []string{"destroy", "--platform", "dev", "--environment", "de", "--path", "test"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Version: 3, Metadata: map[string]interface{}{"vault_destroy_secret_not_before": time.Now().AddDate(0, 0, 8).Format("2006-01-02")}},
		},
		"destroy_1_version": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Input:    "y\ny\n",
			Args:     []string{"destroy", "--platform", "dev", "--environment", "de", "--path", "test"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Version: 1, Metadata: map[string]interface{}{"vault_destroy_secret_not_before": time.Now().AddDate(0, 0, 8).Format("2006-01-02")}},
		},
		"destroy_already_marked": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Deleted: true, Metadata: map[string]interface{}{"vault_destroy_secret_not_before": time.Now().AddDate(0, 0, 8).Format("2006-01-02")}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}, Deleted: true},
			},
			Input:    "y\ny\n",
			Args:     []string{"destroy", "--platform", "dev", "--environment", "de", "--path", "test"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Version: 1, Metadata: map[string]interface{}{"vault_destroy_secret_not_before": time.Now().AddDate(0, 0, 8).Format("2006-01-02")}},
		},
		"destroy_confirmation_no_1": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Deleted: true, Metadata: map[string]interface{}{"vault_destroy_secret_not_before": time.Now().AddDate(0, 0, 8).Format("2006-01-02")}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}, Deleted: true},
			},
			Input:    "n\n",
			Args:     []string{"destroy", "--platform", "dev", "--environment", "de", "--path", "test"},
			Expected: helper.SecretExpectedError{Error: errAskConfirmationNotConfirmed},
		},
		"destroy_confirmation_no_2": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Deleted: true, Metadata: map[string]interface{}{"vault_destroy_secret_not_before": time.Now().AddDate(0, 0, 8).Format("2006-01-02")}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}, Deleted: true},
			},
			Input:    "y\nn\n",
			Args:     []string{"destroy", "--platform", "dev", "--environment", "de", "--path", "test"},
			Expected: helper.SecretExpectedError{Error: errAskConfirmationNotConfirmed},
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
				assert.Contains(t, stdout, "all available versions of the secret have been deleted", "stdout should contain 'all available versions of the secret have been deleted'")
				assert.Contains(t, stdout, "secret has been marked for destruction", "stdout should contain 'secret has been marked for destruction'")
				secret, err := client.Logical().Read("kv/data/" + expectedSecret.Path)
				if assert.NoError(t, err) {
					assert.Nil(t, secret.Data["Data"])
					metadata, ok := secret.Data["metadata"].(map[string]interface{})["custom_metadata"].(map[string]interface{})
					if !ok {
						t.Fatal("we should have a custom_metadata map[string]interface{}")
					} else {
						assert.Equal(t, expectedSecret.Metadata, metadata, "values should be equal")
					}

					for i := 1; i <= expectedSecret.Version; i++ {
						data := map[string][]string{
							"version": {strconv.Itoa(i)},
						}
						t.Log("Executing assert on version " + strconv.Itoa(i))
						secret, err := client.Logical().ReadWithData("kv/data/"+expectedSecret.Path, data)
						assert.Nil(t, err, "we should be able to read the deleted secret version but got an error")
						assert.NotNil(t, secret, "we should get a secret even if the version has been deleted")

						if expectedSecret.Data == nil {
							assert.Empty(t, secret.Data["data"], "no secret data should be returned as the secret should be deleted")
							assert.NotEmpty(t, secret.Data["metadata"].(map[string]interface{})["deletion_time"], "we should get a deletion time on a deleted secret version")
						}

						secret, err = client.Logical().ReadWithData("kv/data/ci/"+expectedSecret.Path, data)
						assert.Nil(t, err, "we should be able to read the deleted secret version but got an error")
						assert.NotNil(t, secret, "we should get a secret even if the version has been deleted")
						if expectedSecret.Data == nil {
							assert.Empty(t, secret.Data["data"], "no secret data should be returned as the secret should be deleted")
							assert.NotEmpty(t, secret.Data["metadata"].(map[string]interface{})["deletion_time"], "we should get a deletion time on a deleted secret version")
						}
					}
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
