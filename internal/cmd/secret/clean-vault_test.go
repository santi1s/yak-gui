package secret

import (
	"testing"
	"time"

	"github.com/santi1s/yak/internal/helper"
	"github.com/hashicorp/vault/api"

	"github.com/stretchr/testify/assert"
)

func TestE2eDestroySecret(t *testing.T) {
	var testScenarios = map[string]secretE2eTest{
		"existing_secret_to_be_destroyed": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz", "vault_destroy_secret_not_before": "2020-01-12"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Destroyed: true},
			Args:     []string{"clean-vault", "--platform", "dev", "--force"},
		},
		"existing_secret_to_be_kept_without_destroy_metadata": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Destroyed: false, Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz"}},
			Args:     []string{"clean-vault", "--platform", "dev", "--force"},
		},
		"existing_secret_to_be_kept_with_destroy_metadata": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz", "vault_destroy_secret_not_before": time.Now().AddDate(0, 0, 3).Format("2006-01-02")}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Destroyed: false, Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz", "vault_destroy_secret_not_before": time.Now().AddDate(0, 0, 3).Format("2006-01-02")}},
			Args:     []string{"clean-vault", "--platform", "dev", "--force"},
		},
		"existing_secret_to_be_kept_with_destroy_metadata_same_day": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz", "vault_destroy_secret_not_before": time.Now().Format("2006-01-02")}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Destroyed: false, Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz", "vault_destroy_secret_not_before": time.Now().Format("2006-01-02")}},
			Args:     []string{"clean-vault", "--platform", "dev", "--force"},
		},
		"non_existent_secret": {
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Destroyed: true, Data: map[string]interface{}{"mykey": "myvalue"}},
			Args:     []string{"clean-vault", "--platform", "dev", "--force"},
		},
		"dry_run": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz", "vault_destroy_secret_not_before": "2020-01-12"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Destroyed: false, Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz", "vault_destroy_secret_not_before": "2020-01-12"}},
			Args:     []string{"clean-vault", "--platform", "dev"},
		},
		"existing_secret_to_be_destroyed_all_platforms": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz", "vault_destroy_secret_not_before": "2020-01-12"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Destroyed: true, Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz", "vault_destroy_secret_not_before": "2020-01-12"}},
			Args:     []string{"clean-vault", "--all-platforms", "--force"},
		},
		"existing_secret_to_be_kept_without_destroy_metadata_all_platforms": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Destroyed: false, Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz"}},
			Args:     []string{"clean-vault", "--all-platforms", "--force"},
		},
		"existing_secret_to_be_kept_with_destroy_metadata_all_platforms": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz", "vault_destroy_secret_not_before": time.Now().AddDate(0, 0, 3).Format("2006-01-02")}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Destroyed: false, Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz", "vault_destroy_secret_not_before": time.Now().AddDate(0, 0, 3).Format("2006-01-02")}},
			Args:     []string{"clean-vault", "--all-platforms", "--force"},
		},
		"existing_secret_to_be_kept_with_destroy_metadata_same_day_all_platforms": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz", "vault_destroy_secret_not_before": time.Now().Format("2006-01-02")}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Destroyed: false, Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz", "vault_destroy_secret_not_before": time.Now().Format("2006-01-02")}},
			Args:     []string{"clean-vault", "--all-platforms", "--force"},
		},
		"non_existent_secret_all_platforms": {
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Destroyed: true, Data: map[string]interface{}{"mykey": "myvalue"}},
			Args:     []string{"clean-vault", "--all-platforms", "--force"},
		},
		"dry_run_all_platforms": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz", "vault_destroy_secret_not_before": "2020-01-12"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test", Destroyed: false, Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"owner": "foo", "source": "bar", "usage": "baz", "vault_destroy_secret_not_before": "2020-01-12"}},
			Args:     []string{"clean-vault", "--all-platforms"},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			initE2ETest()
			cluster, client := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			if expectedSecret, ok := v.Expected.(helper.Secret); ok {
				_, stderr, err := helper.ExecuteCobraCommand(secretCmd, v.Args)
				secretCmd.RemoveCommand(cleanVaultCmd)
				initCleanVaultCmdFlags()
				secretCmd.AddCommand(cleanVaultCmd)

				assert.Empty(t, stderr, "stderr should be empty")

				if expectedSecret.Destroyed {
					if assert.Nil(t, err, "we should not get any error while destroying a secret") {
						secret, err := client.Logical().Read("kv/data/" + expectedSecret.Path)
						assert.Nil(t, err, "we should not get an error")
						assert.Nil(t, secret, "we should get nil as secret has it has been destroyed")

						secret, err = client.Logical().Read("kv/data/ci/" + expectedSecret.Path)
						assert.Nil(t, err, "we should not get an error")
						assert.Nil(t, secret, "we should get nil as secret has it has been destroyed")

						secret, err = client.Logical().Read("kv/metadata/" + expectedSecret.Path)
						assert.Nil(t, err, "we should not get an error")
						assert.Nil(t, secret, "we should get nil as secret has it has been destroyed")

						secret, err = client.Logical().Read("kv/metadata/ci/" + expectedSecret.Path)
						assert.Nil(t, err, "we should not get an error")
						assert.Nil(t, secret, "we should get nil as secret has it has been destroyed")
					}
				} else {
					if assert.Nil(t, err, "we should not get any error while destroying a secret") {
						secret, err := client.Logical().Read("kv/data/" + expectedSecret.Path)
						assert.Nil(t, err, "we should not get an error")
						assert.NotNil(t, secret, "we should get a secret has it should not have been destroyed")

						secret, err = client.Logical().Read("kv/data/ci/" + expectedSecret.Path)
						assert.Nil(t, err, "we should not get an error")
						assert.NotNil(t, secret, "we should get a secret has it should not have been destroyed")

						secret, err = client.Logical().Read("kv/metadata/" + expectedSecret.Path)
						assert.Nil(t, err, "we should not get an error")
						assert.NotNil(t, secret, "we should get a secret has it should not have been destroyed")

						secret, err = client.Logical().Read("kv/metadata/ci/" + expectedSecret.Path)
						assert.Nil(t, err, "we should not get an error")
						assert.NotNil(t, secret, "we should get a secret has it should not have been destroyed")
					}
				}
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}

func TestDestroySecret(t *testing.T) {
	var testScenarios = map[string]secretUnitTest{
		"existing_secret": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Metadata: map[string]interface{}{"vault_destroy_secret_not_before": "2020-01-12"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test"},
		},
		"non_existing_secret": {
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test"},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			cluster, client := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			if expectedSecret, ok := v.Expected.(helper.Secret); ok {
				err := destroySecret([]*api.Client{client}, expectedSecret.Path)
				if assert.Nil(t, err, "err should be nil") {
					secret, err := client.Logical().Read("kv/data/" + expectedSecret.Path)
					assert.Nil(t, err, "we should not get an error")
					assert.Nil(t, secret, "we should get nil as secret has it has been destroyed")

					secret, err = client.Logical().Read("kv/data/ci/" + expectedSecret.Path)
					assert.Nil(t, err, "we should not get an error")
					assert.Nil(t, secret, "we should get nil as secret has it has been destroyed")

					secret, err = client.Logical().Read("kv/metadata/" + expectedSecret.Path)
					assert.Nil(t, err, "we should not get an error")
					assert.Nil(t, secret, "we should get nil as secret has it has been destroyed")

					secret, err = client.Logical().Read("kv/metadata/ci/" + expectedSecret.Path)
					assert.Nil(t, err, "we should not get an error")
					assert.Nil(t, secret, "we should get nil as secret has it has been destroyed")
				}
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}
