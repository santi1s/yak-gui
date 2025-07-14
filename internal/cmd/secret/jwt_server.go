package secret

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
)

type JWTServerSpec struct {
	ServiceName string            `json:"service_name"`
	LocalName   string            `json:"local_name"`
	Algorithm   string            `json:"algorithm"`
	Clients     map[string]string `json:"clients"`
}

var (
	jwtServerCmd = &cobra.Command{
		Use:     "server",
		Short:   "Create/update vault secret for server side JWT token",
		RunE:    createJwtServerSecret,
		Example: "yak secret jwt server -P dev -p doctolib/secrets --local-name monolith --service-name organization_admin --client-name personal-assistant --client-secret <HMAC-SHA256 Secret>",
	}

	jwtServerSecretKeyPrefix = "INTERSERVICE_SERVER"
)

func (s *JWTServerSpec) addClient(name, secret string) {
	s.Clients[name] = secret
}

func (s *JWTServerSpec) newFromInterface(jsonStr interface{}) error {
	if jsonString, ok := jsonStr.(string); ok {
		err := json.Unmarshal([]byte(jsonString), s)
		if err != nil {
			return errInvalidJSONSecretData
		}
		return nil
	}
	return errors.New("invalid json string")
}

func patchJWTSecretKeyData(clients []*api.Client, secretPath string, data map[string]interface{}) (*api.Secret, error) {
	const mount = "kv/data/"
	var payloadSpec, dataSpec JWTServerSpec

	latestVersion, err := GetLatestVersion(clients, secretPath)
	if err != nil {
		return nil, err
	}
	if latestVersion == -1 {
		return nil, helper.ErrSecretNotFound
	}

	currentSecret, err := ReadSecretData(clients, secretPath, latestVersion)
	if err != nil {
		return nil, err
	}

	payload := currentSecret.Data["data"].(map[string]interface{})

	for k, v := range data {
		_, exists := payload[k]
		if exists {
			err := payloadSpec.newFromInterface(payload[k])
			if err != nil {
				return nil, err
			}
			err = dataSpec.newFromInterface(data[k])
			if err != nil {
				return nil, err
			}
			for k, v := range dataSpec.Clients {
				payloadSpec.addClient(k, v)
			}
			payloadSpecJSON, err := json.Marshal(payloadSpec)
			if err != nil {
				return nil, err
			}
			payload[k] = string(payloadSpecJSON)
		} else {
			payload[k] = v
		}
	}

	s, err := PatchSecretData(clients, secretPath, payload)
	if err != nil {
		return s, err
	}

	return s, nil
}

func createJwtServerSecret(cmd *cobra.Command, _ []string) error {
	var (
		secretVault             *api.Secret
		err                     error
		jwtServerSecretMetadata = map[string]interface{}{
			"owner":  providedFlags.owner,
			"source": "JWT token",
			"usage":  "JWT interservice communication",
		}
	)

	err = ValidateClientSecret(providedFlags.clientSecret)
	if err != nil {
		return err
	}

	serverSpec := JWTServerSpec{
		ServiceName: providedFlags.serviceName,
		Algorithm:   "HS256",
		LocalName:   providedFlags.localName,
		Clients: map[string]string{
			providedFlags.clientName: providedFlags.clientSecret,
		},
	}

	config, err := helper.GetVaultConfig(providedFlags.platform, providedFlags.environment)
	if err != nil {
		return err
	}

	clients, err := helper.VaultLoginWithAwsAndGetClients(config)
	if err != nil {
		return err
	}

	serverSpecJSON, err := json.Marshal(serverSpec)
	if err != nil {
		return err
	}

	snakeCaseTargetService := strings.ReplaceAll(strings.ToUpper(providedFlags.serviceName), "-", "_")

	jwtServerSecretData := map[string]interface{}{
		generateDataKey(jwtServerSecretKeyPrefix, snakeCaseTargetService, "CONFIG"): string(serverSpecJSON),
	}

	secretVersion, err := GetLatestVersion(clients, config.SecretPrefix+"/"+providedFlags.path)
	if err != nil {
		return err
	}

	if secretVersion != -1 {
		secretVault, err = patchJWTSecretKeyData(clients, config.SecretPrefix+"/"+providedFlags.path, jwtServerSecretData)

		if err != nil {
			return err
		}
	} else {
		secretVault, err = WriteSecretData(clients, config.SecretPrefix+"/"+providedFlags.path, jwtServerSecretData)
		if err != nil {
			return err
		}
		err = WriteSecretMetadata(clients, config.SecretPrefix+"/"+providedFlags.path, jwtServerSecretMetadata)
		if err != nil {
			return fmt.Errorf("%s: %s", errMetadataCouldNotBeAdded, err)
		}

		secretVault.Data["custom_metadata"] = jwtServerSecretMetadata
	}
	if secretVault == nil {
		return errors.New("unexpected nil secret")
	}
	return cli.PrintJSON(secretVault.Data)
}

func init() {
	jwtServerCmd.Flags().SortFlags = false
	jwtServerCmd.Flags().StringVarP(&providedFlags.owner, "owner", "o", "", "owner of the secret (mandatory)")
	jwtServerCmd.Flags().StringVarP(&providedFlags.localName, "local-name", "L", "", "Local name (mandatory)")
	jwtServerCmd.Flags().StringVarP(&providedFlags.serviceName, "service-name", "S", "", "Service name (mandatory)")
	jwtServerCmd.Flags().StringVarP(&providedFlags.clientName, "client-name", "C", "", "Client name (mandatory)")
	jwtServerCmd.Flags().StringVar(&providedFlags.clientSecret, "client-secret", "", "Client secret (mandatory)")
	_ = jwtServerCmd.MarkFlagRequired("owner")
	_ = jwtServerCmd.MarkFlagRequired("local-name")
	_ = jwtServerCmd.MarkFlagRequired("service-name")
	_ = jwtServerCmd.MarkFlagRequired("client-name")
	_ = jwtServerCmd.MarkFlagRequired("client-secret")
}
