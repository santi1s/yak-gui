package secret

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/doctolib/yak/internal/helper"

	"github.com/stretchr/testify/assert"
)

var (
	testJWTServerSpec = JWTServerSpec{
		ServiceName: "service_name",
		Algorithm:   "HS256",
		LocalName:   "local_name",
		Clients:     map[string]string{"client_name": testJWTClientHMAC},
	}
	testJWTServerSpecNonCamel = JWTServerSpec{
		ServiceName: "service-name",
		Algorithm:   "HS256",
		LocalName:   "local_name",
		Clients:     map[string]string{"client_name": testJWTClientHMAC},
	}
	testJWTServerSpec2 = JWTServerSpec{
		ServiceName: "service_name",
		Algorithm:   "HS256",
		LocalName:   "local_name",
		Clients:     map[string]string{"client_name": testNewJWTClientHMAC},
	}

	testJWTServerSpecMultiClient = JWTServerSpec{
		ServiceName: "service_name",
		Algorithm:   "HS256",
		LocalName:   "local_name",
		Clients:     map[string]string{"client_name": testJWTClientHMAC, "new_client_name": testJWTClientHMAC},
	}
	testJWTClientHMAC, _    = generateSecret(64)
	testNewJWTClientHMAC, _ = generateSecret(64)
)

func TestE2EjwtServerSecret(t *testing.T) {
	serverSpecJSON, _ := json.Marshal(testJWTServerSpec)
	serverSpecJSON2, _ := json.Marshal(testJWTServerSpec2)
	testJWTServerSpecNonCamelJSON, _ := json.Marshal(testJWTServerSpecNonCamel)
	serverMultiClientSpecJSON, _ := json.Marshal(testJWTServerSpecMultiClient)
	var testScenarios = map[string]secretE2eTest{
		"normal_usage_new_server_config": {
			Initial: nil,
			Input:   "",
			Args: []string{"jwt", "server", "--owner", "sre", "--platform", "dev", "--environment", "de", "--path", "test", "--local-name", "local_name", "--service-name",
				"service_name", "--client-name", "client_name", "--client-secret", testJWTClientHMAC},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test",
				Data:     map[string]interface{}{"INTERSERVICE_SERVER_SERVICE_NAME_CONFIG": string(serverSpecJSON)},
				Metadata: map[string]interface{}{"owner": "sre", "usage": "JWT interservice communication", "source": "JWT token"}},
		},
		"normal_usage_new_non_snake_case_service_server_config": {
			Initial: nil,
			Input:   "",
			Args: []string{"jwt", "server", "--owner", "sre", "--platform", "dev", "--environment", "de", "--path", "test", "--local-name", "local_name", "--service-name",
				"service-name", "--client-name", "client_name", "--client-secret", testJWTClientHMAC},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test",
				Data:     map[string]interface{}{"INTERSERVICE_SERVER_SERVICE_NAME_CONFIG": string(testJWTServerSpecNonCamelJSON)},
				Metadata: map[string]interface{}{"owner": "sre", "usage": "JWT interservice communication", "source": "JWT token"}},
		},
		"normal_usage_server_config_new_client_secret": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test",
				Data:     map[string]interface{}{"INTERSERVICE_SERVER_SERVICE_NAME_CONFIG": string(serverSpecJSON)},
				Metadata: map[string]interface{}{"owner": "sre", "usage": "JWT interservice communication", "source": "JWT token"}}},
			Input: "",
			Args: []string{"jwt", "server", "--owner", "sre", "--platform", "dev", "--environment", "de", "--path", "test", "--local-name", "local_name", "--service-name",
				"service_name", "--client-name", "new_client_name", "--client-secret", testJWTClientHMAC},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test",
				Data:     map[string]interface{}{"INTERSERVICE_SERVER_SERVICE_NAME_CONFIG": string(serverMultiClientSpecJSON)},
				Metadata: map[string]interface{}{"owner": "sre", "usage": "JWT interservice communication", "source": "JWT token"}},
		},
		"normal_usage_server_config_replace_client_secret": {
			Initial: []helper.Secret{{Path: "dev-aws-de-fra-1/test",
				Data:     map[string]interface{}{"INTERSERVICE_SERVER_SERVICE_NAME_CONFIG": string(serverSpecJSON)},
				Metadata: map[string]interface{}{"owner": "sre", "usage": "JWT interservice communication", "source": "JWT token"}}},
			Input: "",
			Args: []string{"jwt", "server", "--owner", "sre", "--platform", "dev", "--environment", "de", "--path", "test", "--local-name", "local_name", "--service-name",
				"service_name", "--client-name", "client_name", "--client-secret", testNewJWTClientHMAC},
			Expected: helper.Secret{Path: "dev-aws-de-fra-1/test",
				Data:     map[string]interface{}{"INTERSERVICE_SERVER_SERVICE_NAME_CONFIG": string(serverSpecJSON2)},
				Metadata: map[string]interface{}{"owner": "sre", "usage": "JWT interservice communication", "source": "JWT token"}},
		},
		"missing_platform": {
			Initial: nil,
			Input:   "",
			Args: []string{"jwt", "server", "--owner", "sre", "--environment", "de", "--path", "test", "--local-name", "local_name", "--service-name",
				"service_name", "--client-name", "client_name", "--client-secret", testJWTClientHMAC},
			Expected: helper.SecretExpectedError{
				Name:  "errEnvironmentCantBeSetWithoutPlatform",
				Error: helper.ErrEnvironmentCantBeSetWithoutPlatform,
			},
		},
		"missing_env": {
			Initial: nil,
			Input:   "",
			Args: []string{"jwt", "server", "--owner", "sre", "--platform", "dev", "--path", "test", "--local-name", "local_name", "--service-name",
				"service_name", "--client-name", "client_name", "--client-secret", testJWTClientHMAC},
			Expected: helper.Secret{Path: "common/test",
				Data:     map[string]interface{}{"INTERSERVICE_SERVER_SERVICE_NAME_CONFIG": string(serverSpecJSON)},
				Metadata: map[string]interface{}{"owner": "sre", "usage": "JWT interservice communication", "source": "JWT token"}},
		},
		"invalid_jwt_secret": {
			Initial: nil,
			Input:   "",
			Args: []string{"jwt", "server", "--owner", "sre", "--platform", "dev", "--environment", "de", "--path", "test", "--local-name", "local_name", "--service-name",
				"service_name", "--client-name", "client_name", "--client-secret", "XXXXXX"},
			Expected: helper.SecretExpectedError{
				Name:  "errInvalidJWTSecret",
				Error: errInvalidJWTSecret,
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
			if expectedSecret, ok := v.Expected.(helper.Secret); ok {
				// if we expect a secret to be returned
				assert.NoError(t, err)
				assert.Empty(t, stderr, "stderr should be empty")
				secret, err := client.Logical().Read("kv/data/" + expectedSecret.Path)
				if assert.NoError(t, err, nil, "error should be nil") {
					assert.Equal(t, expectedSecret.Data, secret.Data["data"], "data should be equal")
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
