package secret

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/santi1s/yak/internal/helper"

	"github.com/stretchr/testify/assert"
)

var (
	testJWTClientSpec = JWTClientSpec{
		TargetService: "target_service",
		Algorithm:     "HS256",
		LocalName:     "local_name",
		Secret:        testJWTClientSecretHMAC,
	}
	testJWTClientSpecNonCamel = JWTClientSpec{
		TargetService: "target-service",
		Algorithm:     "HS256",
		LocalName:     "local_name",
		Secret:        testJWTClientSecretHMAC,
	}
	testJWTClientSpec2 = JWTClientSpec{
		TargetService: "target_service",
		Algorithm:     "HS256",
		LocalName:     "local_name",
		Secret:        testJWTClientSecretHMAC2,
	}

	testJWTClientSecretHMAC, _  = generateSecret(64)
	testJWTClientSecretHMAC2, _ = generateSecret(64)
)

func TestE2EjwtClientSecret(t *testing.T) {
	clientSpecJSON, _ := json.Marshal(testJWTClientSpec)
	clientSpecJSON2, _ := json.Marshal(testJWTClientSpec2)
	clientSpecJSONNonCamel, _ := json.Marshal(testJWTClientSpecNonCamel)

	var testScenarios = map[string]secretE2eTest{
		"normal_usage": {
			Initial: nil,
			Input:   "",
			Args:    []string{"jwt", "client", "--owner", "sre", "--platform", "dev", "--environment", "de", "--path", "test", "--local-name", "local_name", "--target-service", "target_service", "--secret", testJWTClientSpec.Secret},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test",
				Data:     map[string]interface{}{"INTERSERVICE_CLIENT_TARGET_SERVICE_CONFIG": string(clientSpecJSON)},
				Metadata: map[string]interface{}{"owner": "sre", "usage": "JWT interservice communication", "source": "JWT token"}},
			ExpectChangedData: false,
		},
		"normal_usage_non_snake_case_service": {
			Initial: nil,
			Input:   "",
			Args:    []string{"jwt", "client", "--owner", "sre", "--platform", "dev", "--environment", "de", "--path", "test", "--local-name", "local_name", "--target-service", "target-service", "--secret", testJWTClientSpec.Secret},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test",
				Data:     map[string]interface{}{"INTERSERVICE_CLIENT_TARGET_SERVICE_CONFIG": string(clientSpecJSONNonCamel)},
				Metadata: map[string]interface{}{"owner": "sre", "usage": "JWT interservice communication", "source": "JWT token"}},
			ExpectChangedData: false,
		},
		"normal_usage_no_jwt_secret": {
			Initial: nil,
			Input:   "",
			Args:    []string{"jwt", "client", "--owner", "sre", "--platform", "dev", "--environment", "de", "--path", "test", "--local-name", "local_name", "--target-service", "target_service"},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test",
				Data:     map[string]interface{}{"INTERSERVICE_CLIENT_TARGET_SERVICE_CONFIG": string(clientSpecJSON)},
				Metadata: map[string]interface{}{"owner": "sre", "usage": "JWT interservice communication", "source": "JWT token"}},
			ExpectChangedData: true,
		},
		"update_jwt_secret": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test",
				Data:     map[string]interface{}{"INTERSERVICE_CLIENT_TARGET_SERVICE_CONFIG": string(clientSpecJSON)},
				Metadata: map[string]interface{}{"owner": "sre", "usage": "JWT interservice communication", "source": "JWT token"}}},
			Input: "",
			Args:  []string{"jwt", "client", "--owner", "sre", "--platform", "dev", "--environment", "de", "--path", "test", "--local-name", "local_name", "--target-service", "target_service", "--secret", testJWTClientSpec2.Secret},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test",
				Data:     map[string]interface{}{"INTERSERVICE_CLIENT_TARGET_SERVICE_CONFIG": string(clientSpecJSON2)},
				Metadata: map[string]interface{}{"owner": "sre", "usage": "JWT interservice communication", "source": "JWT token"}},
			ExpectChangedData: false,
		},
		"missing_owner_metadata": {
			Initial: nil,
			Input:   "",
			Args:    []string{"jwt", "client", "--owner", "", "--platform", "dev", "--environment", "de", "--path", "test", "--local-name", "local_name", "--target-service", "target_service"},
			Expected: helper.SecretExpectedError{
				Name:  "errOwnerCantBeEmpty",
				Error: errOwnerCantBeEmpty,
			},
			ExpectChangedData: false,
		},
		"missing_platform": {
			Initial: nil,
			Input:   "",
			Args:    []string{"jwt", "client", "--owner", "sre", "--environment", "de", "--path", "test", "--local-name", "local_name", "--target-service", "target_service", "--secret", testJWTClientSpec.Secret},
			Expected: helper.SecretExpectedError{
				Name:  "errEnvironmentCantBeSetWithoutPlatform",
				Error: helper.ErrEnvironmentCantBeSetWithoutPlatform,
			},
			ExpectChangedData: false,
		},
		"missing_env": {
			Initial: nil,
			Input:   "",
			Args:    []string{"jwt", "client", "--owner", "sre", "--platform", "dev", "--path", "test", "--local-name", "local_name", "--target-service", "target_service", "--secret", testJWTClientSpec.Secret},
			Expected: helper.Secret{Path: "common/test",
				Data:     map[string]interface{}{"INTERSERVICE_CLIENT_TARGET_SERVICE_CONFIG": string(clientSpecJSON)},
				Metadata: map[string]interface{}{"owner": "sre", "usage": "JWT interservice communication", "source": "JWT token"}},
			ExpectChangedData: false,
		},
		"invalid_jwt_secret": {
			Initial: nil,
			Input:   "",
			Args:    []string{"jwt", "client", "--owner", "sre", "--platform", "dev", "--path", "test", "--local-name", "local_name", "--target-service", "target_service", "--secret", "XXXXXXX"},
			Expected: helper.SecretExpectedError{
				Name:  "errInvalidJWTSecret",
				Error: errInvalidJWTSecret,
			},
			ExpectChangedData: false,
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
				secret, err := client.Logical().Read("kv/data/" + expectedSecret.Path)
				if assert.NoError(t, err, nil, "error should be nil") {
					if !v.ExpectChangedData {
						assert.Equal(t, expectedSecret.Data, secret.Data["data"], "data should be equal")
					} else {
						var secretJWTClientSpec JWTClientSpec
						payload := secret.Data["data"].(map[string]interface{})
						config, _ := payload["INTERSERVICE_CLIENT_TARGET_SERVICE_CONFIG"].(string)

						assert.NoError(t, json.Unmarshal([]byte(config), &secretJWTClientSpec), "Unmarshal error should be nil")
						assert.Equal(t, testJWTClientSpec.TargetService, secretJWTClientSpec.TargetService, "TargetService should be equal")
						assert.Equal(t, testJWTClientSpec.LocalName, secretJWTClientSpec.LocalName, "LocalName should be equal")
						assert.Equal(t, testJWTClientSpec.Algorithm, secretJWTClientSpec.Algorithm, "Algorithm should be equal")
						assert.Nil(t, ValidateClientSecret(secretJWTClientSpec.Secret), "JWT secret should be valid")
					}
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
