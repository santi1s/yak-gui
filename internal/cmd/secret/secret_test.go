package secret

import (
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	kv "github.com/hashicorp/vault-plugin-secrets-kv"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/helper/testhelpers/corehelpers"
	vaulthttp "github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/vault"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

const (
	testVaultToken = "token" // This is the root token
)

var (
	cluster *vault.TestCluster
	client  *api.Client
)

func setupVaultCluster(t *testing.T, testSecret ...helper.Secret) (*vault.TestCluster, *api.Client) {
	logger := corehelpers.NewTestLogger(t)
	logger.StopLogging()
	cluster = vault.NewTestCluster(t, &vault.CoreConfig{
		DevToken: testVaultToken,
		LogicalBackends: map[string]logical.Factory{
			"kv": kv.Factory,
		},
	}, &vault.TestClusterOptions{
		HandlerFunc: vaulthttp.Handler,
		NumCores:    1,
		Logger:      logger,
	})
	cluster.Start()

	core := cluster.Cores[0].Core
	vault.TestWaitActiveForwardingReady(t, core)

	client = cluster.Cores[0].Client

	if err := cluster.Cores[0].Client.Sys().Mount("kv", &api.MountInput{
		Type: "kv",
		Options: map[string]string{
			"version": "2",
		},
	}); err != nil {
		t.Fatal(err)
	}

	// write k/v data pairs into vault
	for _, v := range testSecret {
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			_, err := client.Logical().Write("kv/data/"+v.Path, map[string]interface{}{
				"data": v.Data,
			})
			if err != nil {
				if time.Now().Equal(deadline) || time.Now().After(deadline) {
					t.Fatal(err)
				}
				time.Sleep(100 * time.Millisecond)
			} else {
				break
			}
		}

		_, err := client.Logical().Write("kv/metadata/"+v.Path, map[string]interface{}{
			"custom_metadata": v.Metadata,
		})
		if err != nil {
			t.Fatal(err)
		}

		if v.Deleted {
			_, err := client.Logical().Delete("kv/data/" + v.Path)
			if err != nil {
				t.Fatal(err)
			}
		}

		if v.Destroyed {
			_, err := client.Logical().Write("kv/destroy/"+v.Path, map[string]interface{}{
				"versions": strconv.Itoa(v.Version),
			})
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	return cluster, client
}

func initE2ETest() {
	var c = `clusters:
  shared:
    endpoint: https://vault.local:8200
platforms:
  dev:
    clusters: [shared]
    awsProfile: dev-sso
    awsRegion: eu-central-1
    vaultRole: administrator
    environments:
      de: dev-aws-de-fra-1`
	helper.InitViper(c)

	// replace function used to get the vault client by a function that will return the test client
	helper.VaultLoginWithAwsAndGetClients = getTestVaultClient
	initConfig = func(cmd *cobra.Command, args []string) {}
	providedFlags = secretFlags{}
}

func getTestVaultClient(config *helper.VaultConfig) ([]*api.Client, error) {
	return []*api.Client{client}, nil
}

type secretUnitTest struct {
	Path     string
	Version  int
	Initial  []helper.Secret
	Update   map[string]interface{}
	Expected interface{}
}

type secretE2eTest struct {
	Initial           []helper.Secret
	Args              []string
	Input             string
	Expected          interface{}
	ExpectChangedData bool
}

type readFromIoreaderTest struct {
	initial  string
	expected interface{}
}

func TestReadFromStdin(t *testing.T) {
	var testScenarios = map[string]readFromIoreaderTest{
		"nominal": {
			initial: `key: value
key2: value2`,
			expected: map[string]interface{}{"key": "value", "key2": "value2"},
		},
		"with_colon": {
			initial:  `key: value:with:colons`,
			expected: map[string]interface{}{"key": "value:with:colons"},
		},
		"with_quotes": {
			initial: `key: "value"
key2: "value2"`,
			expected: map[string]interface{}{"key": "value", "key2": "value2"},
		},
		"multiline_folded": {
			initial: `key: >
  value
  value2`,
			expected: map[string]interface{}{"key": "value value2"},
		},
		"multiline_literal": {
			initial: `key: |
  value
  value2`,
			expected: map[string]interface{}{"key": "value\nvalue2"},
		},
		// FIXME: return an error when duplicat keys are found
		// 		"duplicate_key": {
		// 			initial: `key: value
		// key: value2`,
		// 			expected: helper.helper.SecretExpectedError{},
		// 		},
		"not_a_string": {
			initial: `foo:
  bar: baz
foo2: bar2`,
			expected: helper.SecretExpectedError{
				Contains: "error unmarshaling JSON: while decoding JSON: json: cannot unmarshal object into Go value of type string",
			},
		},
		"empty_value": {
			initial:  `foo:`,
			expected: map[string]interface{}{"foo": ""},
		},
	}
	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			r := strings.NewReader(v.initial)
			cli.SetIn(r)
			data, err := readFromStdin()

			if expectedString, ok := v.expected.(map[string]interface{}); ok {
				// if we expect a secret to be returned
				if assert.NoError(t, err) {
					assert.Equal(t, expectedString, data)
				}
			} else if expectedError, ok := v.expected.(helper.SecretExpectedError); ok {
				// if we expect an error to be returned
				assert.Error(t, err)
				assert.Equal(t, map[string]interface{}(nil), data)
				if expectedError.Error != nil && expectedError.Name != "" {
					// if we know which error we expect
					assert.ErrorIsf(t, err, expectedError.Error, "error should be %s", expectedError.Name)
				} else if expectedError.Contains != "" {
					assert.ErrorContains(t, err, expectedError.Contains)
				}
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}

type readFromInteractiveTest struct {
	input    string
	keysOnly bool
	expected interface{}
}

func TestReadFromInteractive(t *testing.T) {
	var testScenarios = map[string]readFromInteractiveTest{
		"nominal": {
			input:    "foo\nbar\nn",
			keysOnly: false,
			expected: map[string]interface{}{"foo": "bar"},
		},
		"nominal_only_keys": {
			input:    "foo\nn",
			keysOnly: true,
			expected: map[string]interface{}{"foo": ""},
		},
		"multiple_pairs": {
			input:    "foo\nbar\ny\nbaz\nqux\nn",
			keysOnly: false,
			expected: map[string]interface{}{"foo": "bar", "baz": "qux"},
		},
		"empty_key": {
			input:    "\nfoo\nn",
			keysOnly: true,
			expected: map[string]interface{}{"foo": ""},
		},
		"duplicate_key": {
			input:    "foo\ny\nbar\ny\nfoo\nbaz\nn",
			keysOnly: true,
			expected: map[string]interface{}{"foo": "", "bar": "", "baz": ""},
		},
		"empty_value": {
			input:    "foo\n\nbar\nn",
			keysOnly: false,
			expected: map[string]interface{}{"foo": "bar"},
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			r := strings.NewReader(v.input)

			cli.SetIn(r)
			cli.SetPasswordReader(&cli.IoReaderPasswordReader{Reader: r})
			cli.SetOut(io.Discard)

			pairs, err := readFromInteractive(v.keysOnly)

			if expectedMap, ok := v.expected.(map[string]interface{}); ok {
				// if we expect a map to be returned
				if assert.NoError(t, err) {
					assert.Equal(t, expectedMap, pairs)
				}
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}

func TestHasEmptyValueFalse(t *testing.T) {
	var x = map[string]interface{}{
		"key":  "value",
		"key2": "value2",
	}
	isEmpty := hasEmptyValue(x)
	assert.False(t, isEmpty)
}

func TestHasEmptyValueTrue(t *testing.T) {
	var x = map[string]interface{}{
		"key":  "",
		"key2": "value2",
	}
	isEmpty := hasEmptyValue(x)
	assert.True(t, isEmpty)
}

func TestHasEmptyValueEmpty(t *testing.T) {
	var x = map[string]interface{}{
		"key":  nil,
		"key2": "value2",
	}
	isEmpty := hasEmptyValue(x)
	assert.True(t, isEmpty)
}

func TestGetLatestVersion(t *testing.T) {
	var testScenarios = map[string]secretUnitTest{
		"latest_is_latest": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"bar": "baz"}},
			},
			Expected: 2,
		},
		"latest_deleted_latest_is_n-1": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"bar": "baz"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"baz": "foo"}, Deleted: true},
			},
			Expected: 2,
		},
		"latest_destroyed_latest_is_n-1": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"bar": "baz"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"baz": "foo"}, Version: 3, Destroyed: true},
			},
			Expected: 2,
		},
		"latest_deleted_latest_is_n-2": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"bar": "baz"}, Deleted: true},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"baz": "foo"}, Deleted: true},
			},
			Expected: 1,
		},
		"latest_destroyed_latest_is_n-2": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"bar": "baz"}, Version: 2, Destroyed: true},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"baz": "foo"}, Version: 3, Destroyed: true},
			},
			Expected: 1,
		},
		"latest_destroyed_n-1_deleted_latest_is_n-2": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"bar": "baz"}, Deleted: true},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"baz": "foo"}, Version: 3, Destroyed: true},
			},
			Expected: 1,
		},
		"latest_deleted_n-1_destroyed_latest_is_n-2": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"bar": "baz"}, Version: 2, Destroyed: true},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"baz": "foo"}, Deleted: true},
			},
			Expected: 1,
		},
		"no_latest_all_versions_are_deleted": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}, Deleted: true},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"bar": "baz"}, Deleted: true},
			},
			Expected: -1,
		},
		"no_latest_all_versions_are_destroyed": {
			Initial: []helper.Secret{
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"foo": "bar"}, Version: 1, Destroyed: true},
				{Path: "dev-aws-de-fra-1/test", Data: map[string]interface{}{"bar": "baz"}, Version: 2, Destroyed: true},
			},
			Expected: -1,
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			cluster, client := setupVaultCluster(t, v.Initial...)
			defer cluster.Cleanup()

			if expectedVersion, ok := v.Expected.(int); ok {
				// if we expect a secret to be returned
				version, err := GetLatestVersion([]*api.Client{client}, v.Initial[0].Path)
				if assert.NoError(t, err, "error should be nil") {
					assert.Equal(t, expectedVersion, version, "values should be equal")
				}
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}
