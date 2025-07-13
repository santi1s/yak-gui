package secret

import (
	"strings"
	"testing"

	"github.com/doctolib/yak/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestSecretListDuplicates(t *testing.T) {
	tests := []struct {
		name           string
		secretValue    string
		args           []string
		expectedOutput string
		expectedError  error
	}{
		{
			name:           "empty_secret_value",
			secretValue:    "",
			args:           []string{"list-duplicates", "--platform", "dev", "--environment", "de", "--path", "dev-aws-de-fra-1/test/path"},
			expectedOutput: "",
			expectedError:  errEmptyInput,
		},
		{
			name:           "no_secret_found",
			secretValue:    "s3cr3t_v@lu3",
			args:           []string{"list-duplicates", "--platform", "dev", "--environment", "de", "--path", "dev-aws-de-fra-1/test/path"},
			expectedOutput: "Walking through path dev-aws-de-fra-1/test/path ...\nSecret not found.\n",
			expectedError:  nil,
		},
		{
			name:           "no_duplicates_found_single_secret",
			secretValue:    "autre-mot-pour-dire-chocolatine",
			args:           []string{"list-duplicates", "--platform", "dev", "--environment", "de", "--path", "dev-aws-de-fra-1/test/path"},
			expectedOutput: "Walking through path dev-aws-de-fra-1/test/path ...\nNo duplicates found for provided secret value. Secret located at dev-aws-de-fra-1/test/path/to/secret2:test-key1\n",
			expectedError:  nil,
		},
		{
			name:           "duplicates_found",
			secretValue:    "chocolatine",
			args:           []string{"list-duplicates", "--platform", "dev", "--environment", "de", "--path", "dev-aws-de-fra-1/test/path"},
			expectedOutput: "Walking through path dev-aws-de-fra-1/test/path ...\nDuplicates found:\n\t* dev-aws-de-fra-1/test/path/to/secret1\n\t\t- test-key1\n\t* dev-aws-de-fra-1/test/path/to/secret2\n\t\t- test-key2\n",
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vaultSecrets := []helper.Secret{
				{
					Path: "dev-aws-de-fra-1/test/path/to/secret1",
					Data: map[string]interface{}{
						"test-key1": "chocolatine",
					},
				},
				{
					Path: "dev-aws-de-fra-1/test/path/to/secret2",
					Data: map[string]interface{}{
						"test-key1": "autre-mot-pour-dire-chocolatine",
						"test-key2": "chocolatine",
					},
				},
			}

			initE2ETest()
			cluster, _ := setupVaultCluster(t, vaultSecrets...)
			defer cluster.Cleanup()
			r := strings.NewReader(tt.secretValue)
			stdout, _, err := helper.ExecuteCobraCommand(secretCmd, tt.args, r)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedOutput, stdout)
		})
	}
}
