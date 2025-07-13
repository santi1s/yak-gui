package secret

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
)

type JWTClientSpec struct {
	TargetService string `json:"target_service"`
	Algorithm     string `json:"algorithm"`
	LocalName     string `json:"local_name"`
	Secret        string `json:"secret"`
}

var (
	jwtClientCmd = &cobra.Command{
		Use:     "client",
		Short:   "Create vault secret for client side JWT token",
		RunE:    createJwtClientSecret,
		Example: "yak secret jwt client -P dev -p personal-assistant/jwt --local-name personal-assistant --target-service organization_admin --secret <HMAC-SHA256 Secret>",
	}

	jwtClientSecretKeyPrefix = "INTERSERVICE_CLIENT"
)

func generateSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateDataKey(parts ...string) string {
	return strings.Join(parts, "_")
}

func ValidateClientSecret(secret string) error {
	if len(secret) < 32 {
		return errInvalidJWTSecret
	}

	if match, _ := regexp.MatchString("^[a-fA-F0-9]+$", secret); !match {
		return errInvalidJWTSecret
	}
	return nil
}

func createJwtClientSecret(cmd *cobra.Command, _ []string) error {
	var (
		secretVault             *api.Secret
		err                     error
		jwtSecret               string
		jwtClientSecretMetadata = map[string]interface{}{
			"owner":  providedFlags.owner,
			"source": "JWT token",
			"usage":  "JWT interservice communication",
		}
	)

	if providedFlags.owner == "" {
		return errOwnerCantBeEmpty
	}

	if providedFlags.secret != "" {
		if ValidateClientSecret(providedFlags.secret) != nil {
			return errInvalidJWTSecret
		}
		jwtSecret = providedFlags.secret
	} else {
		jwtSecret, err = generateSecret(64) // Generate a 64-byte random secret
		if err != nil {
			return err
		}
	}

	clientSpec := JWTClientSpec{
		TargetService: providedFlags.targetService,
		Algorithm:     "HS256",
		LocalName:     providedFlags.localName,
		Secret:        jwtSecret,
	}

	config, err := helper.GetVaultConfig(providedFlags.platform, providedFlags.environment)
	if err != nil {
		return err
	}

	clients, err := helper.VaultLoginWithAwsAndGetClients(config)
	if err != nil {
		return err
	}

	clientSpecJSON, err := json.Marshal(clientSpec)
	if err != nil {
		return err
	}

	snakeCaseTargetService := strings.ReplaceAll(strings.ToUpper(providedFlags.targetService), "-", "_")

	JWTClientSecretData := map[string]interface{}{
		generateDataKey(jwtClientSecretKeyPrefix, snakeCaseTargetService, "CONFIG"): string(clientSpecJSON),
	}

	secretVersion, err := GetLatestVersion(clients, config.SecretPrefix+"/"+providedFlags.path)
	if err != nil {
		return err
	}

	if secretVersion != -1 {
		readSecret, err := ReadSecretMetadata(clients, config.SecretPrefix+"/"+providedFlags.path)
		if err != nil {
			return err
		}

		if readSecret.Data["custom_metadata"].(map[string]interface{})["owner"] != providedFlags.owner {
			return errOwnerMismatch
		}

		secretVault, err = PatchSecretData(clients, config.SecretPrefix+"/"+providedFlags.path, JWTClientSecretData)
		if err != nil {
			return err
		}
	} else {
		secretVault, err = WriteSecretData(clients, config.SecretPrefix+"/"+providedFlags.path, JWTClientSecretData)
		if err != nil {
			return err
		}
		err = WriteSecretMetadata(clients, config.SecretPrefix+"/"+providedFlags.path, jwtClientSecretMetadata)
		if err != nil {
			return fmt.Errorf("%s: %s", errMetadataCouldNotBeAdded, err)
		}

		secretVault.Data["custom_metadata"] = jwtClientSecretMetadata
	}
	return formatOutput(secretVault.Data)
}

func init() {
	jwtClientCmd.Flags().SortFlags = false
	jwtClientCmd.Flags().StringVarP(&providedFlags.owner, "owner", "o", "", "owner of the secret (mandatory)")
	jwtClientCmd.Flags().StringVarP(&providedFlags.localName, "local-name", "L", "", "Local name(mandatory)")
	jwtClientCmd.Flags().StringVarP(&providedFlags.targetService, "target-service", "T", "", "Target Service(mandatory)")
	jwtClientCmd.Flags().StringVarP(&providedFlags.secret, "secret", "S", "", "secret")
	err := jwtClientCmd.MarkFlagRequired("owner")
	if err != nil {
		panic(err)
	}
	_ = jwtClientCmd.MarkFlagRequired("local-name")
	_ = jwtClientCmd.MarkFlagRequired("target-service")
}
