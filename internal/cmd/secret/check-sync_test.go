package secret

import (
	"strings"
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

type checkSyncUnitTest struct {
	Initial  []helper.Secret
	Expected interface{}
	WalkPath string
}

func TestGetNamespaceSyncInfo(t *testing.T) {
	var testScenarios = map[string]checkSyncUnitTest{
		"ok_not_deleted": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
		},
		"ok_deleted": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
		},
		"ko_current_secret_version_mismatch": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue2"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
		},
		"ko_current_ci_version_mismatch": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey2": ""}},
			},
		},
		"ko_ci_version_deleted": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}, Deleted: true},
			},
		},
		"ko_secret_version_deleted": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}, Deleted: true},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			cluster, client := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			result, err := getNamespaceSyncInfo([]*api.Client{client})
			if strings.HasPrefix(k, "ok") {
				if assert.Nil(t, err, "error should be nil") {
					assert.Empty(t, result, "should be empty")
				}
			} else if strings.HasPrefix(k, "ko") {
				if assert.Nil(t, err, "error should be nil") {
					if strings.Contains(k, "mismatch") {
						assert.NotEqual(t, result[v.Initial[0].Path].CI, result[v.Initial[0].Path].Secret, "should not be equal")
					} else if strings.Contains(k, "deleted") {
						if strings.Contains(k, "ci") {
							assert.NotEmpty(t, result[v.Initial[0].Path].CIVersions["1"].(map[string]interface{})["deletion_time"], "should not be empty")
							assert.Empty(t, result[v.Initial[0].Path].SecretVersions["1"].(map[string]interface{})["deletion_time"], "should be empty")
						} else {
							assert.Empty(t, result[v.Initial[0].Path].CIVersions["1"].(map[string]interface{})["deletion_time"], "should be empty")
							assert.NotEmpty(t, result[v.Initial[0].Path].SecretVersions["1"].(map[string]interface{})["deletion_time"], "should not be empty")
						}
					}
				}
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}

func TestVaultWalkPath(t *testing.T) {
	var testScenarios = map[string]checkSyncUnitTest{
		"ok_with_results_1": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: []string{"ci/dev-aws-de-fra-1/test", "dev-aws-de-fra-1/test"},
			WalkPath: "",
		},
		"ok_with_results_2": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: []string{"ci/dev-aws-de-fra-1/test", "dev-aws-de-fra-1/test"},
			WalkPath: "/",
		},
		"ok_with_results_3": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: []string{"ci/dev-aws-de-fra-1/test"},
			WalkPath: "ci/",
		},
		"ok_with_results_4": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: []string{"ci/dev-aws-de-fra-1/test"},
			WalkPath: "/ci/",
		},
		"ok_with_results_5": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": "myvalue"}},
				{Path: "ci/dev-aws-de-fra-1/test", Data: map[string]interface{}{"mykey": ""}},
			},
			Expected: []string{"ci/dev-aws-de-fra-1/test"},
			WalkPath: "/ci",
		},
		"ok_without_results": {
			Initial:  []helper.Secret{},
			Expected: []string{},
			WalkPath: "/nonexisting",
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			cluster, client := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			result, err := helper.WalkVaultPath([]*api.Client{client}, v.WalkPath)
			if expected, ok := v.Expected.([]string); ok {
				if assert.Nil(t, err, "error should be nil") {
					assert.Equal(t, expected, result, "should be equal")
				}
			} else if _, ok := v.Expected.(error); ok {
				if assert.Error(t, err, "should be an error") {
					assert.Nil(t, result, "should be nil")
				}
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}
