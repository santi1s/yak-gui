package repo

import (
	"strconv"
	"testing"

	"github.com/santi1s/yak/internal/helper"
	kv "github.com/hashicorp/vault-plugin-secrets-kv"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/helper/testhelpers/corehelpers"
	vaulthttp "github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/vault"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type kubeSecretE2eTest struct {
	VaultSecrets []helper.Secret
	KubeSecret   []*testFile
	Args         []string
	Expected     interface{}
}

type testFile struct {
	Path    string
	Content string
}

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
	vault.TestWaitActive(t, core)

	client = cluster.Cores[0].Client

	if err := client.Sys().Mount("kv", &api.MountInput{
		Type: "kv",
		Options: map[string]string{
			"version": "2",
		},
	}); err != nil {
		t.Fatal(err)
	}

	// write k/v data pairs into vault
	for _, v := range testSecret {
		_, err := client.Logical().Write("kv/data/"+v.Path, map[string]interface{}{
			"data": v.Data,
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Logical().Write("kv/metadata/"+v.Path, map[string]interface{}{
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
      de: dev-aws-de-fra-1
      fr: dev-aws-fr-par-1
  common:
    clusters: [shared]
    awsProfile: shared-sso
    awsRegion: eu-central-1
    vaultRole: administrator
    environments:`
	helper.InitViper(c)

	// replace function used to get the vault client by a function that will return the test client
	helper.VaultLoginWithAwsAndGetClients = getTestVaultClient
	initConfig = func(cmd *cobra.Command, args []string) {}
	providedRepoSecretFlags = RepoSecretFlags{}
}

func getTestVaultClient(config *helper.VaultConfig) ([]*api.Client, error) {
	return []*api.Client{client}, nil
}

func TestValidateTfeJwtSubjectFormat(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		wantErr  bool
		expected error
	}{
		{
			name:     "valid format with organization, project and workspace",
			subject:  "organization:myorg:project:myproject:workspace:myworkspace",
			wantErr:  false,
			expected: nil,
		},
		{
			name:     "valid format with wildcards",
			subject:  "organization:*:project:*:workspace:*",
			wantErr:  false,
			expected: nil,
		},
		{
			name:     "valid format with mixed wildcards and values",
			subject:  "organization:myorg:project:*:workspace:prod",
			wantErr:  false,
			expected: nil,
		},
		{
			name:     "valid format with numbers and special chars",
			subject:  "organization:my-org-123:project:my_project_v2:workspace:test-env",
			wantErr:  false,
			expected: nil,
		},
		{
			name:     "invalid format - missing workspace part",
			subject:  "organization:myorg:project:myproject",
			wantErr:  true,
			expected: errInvalidTfeJwtSubjectFormat,
		},
		{
			name:     "invalid format - missing project part",
			subject:  "organization:myorg:workspace:myworkspace",
			wantErr:  true,
			expected: errInvalidTfeJwtSubjectFormat,
		},
		{
			name:     "invalid format - missing organization part",
			subject:  "project:myproject:workspace:myworkspace",
			wantErr:  true,
			expected: errInvalidTfeJwtSubjectFormat,
		},
		{
			name:     "invalid format - wrong order",
			subject:  "project:myproject:organization:myorg:workspace:myworkspace",
			wantErr:  true,
			expected: errInvalidTfeJwtSubjectFormat,
		},
		{
			name:     "invalid format - empty string",
			subject:  "",
			wantErr:  true,
			expected: errInvalidTfeJwtSubjectFormat,
		},
		{
			name:     "invalid format - completely wrong format",
			subject:  "this-is-not-a-valid-format",
			wantErr:  true,
			expected: errInvalidTfeJwtSubjectFormat,
		},
		{
			name:     "invalid format - extra colons in values",
			subject:  "organization:my:org:project:my:project:workspace:my:workspace",
			wantErr:  true,
			expected: errInvalidTfeJwtSubjectFormat,
		},
		{
			name:     "invalid format - missing values",
			subject:  "organization::project::workspace:",
			wantErr:  true,
			expected: errInvalidTfeJwtSubjectFormat,
		},
		{
			name:     "invalid format - case sensitive",
			subject:  "Organization:myorg:Project:myproject:Workspace:myworkspace",
			wantErr:  true,
			expected: errInvalidTfeJwtSubjectFormat,
		},
		{
			name:     "valid format - spaces are allowed in values",
			subject:  "organization: myorg :project: myproject :workspace: myworkspace",
			wantErr:  false,
			expected: nil,
		},
		{
			name:     "invalid format - additional parts",
			subject:  "organization:myorg:project:myproject:workspace:myworkspace:extra:part",
			wantErr:  true,
			expected: errInvalidTfeJwtSubjectFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTfeJwtSubjectFormat(tt.subject)

			if tt.wantErr {
				assert.Error(t, err, "Expected an error for subject: %s", tt.subject)
				assert.Equal(t, tt.expected, err, "Expected specific error type")
			} else {
				assert.NoError(t, err, "Expected no error for subject: %s", tt.subject)
			}
		})
	}
}
