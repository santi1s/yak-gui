package secret

import (
	"os"
	"os/exec"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var (
	diffCmd = &cobra.Command{
		Use:   "diff",
		Short: "show the diff between two versions of a secret on vault",
		RunE:  secretDiff,
	}
)

func init() {
	diffCmd.Flags().IntVar(&providedFlags.baseVersion, "base-version", 0, "base version for the diff")
	err := diffCmd.MarkFlagRequired("base-version")
	if err != nil {
		panic(err)
	}

	diffCmd.Flags().IntVar(&providedFlags.diffVersion, "diff-version", 0, "diff version for the diff (latest version is used when not provided)")
}

func secretDiff(cmd *cobra.Command, args []string) error {
	if providedFlags.team != "" {
		awsFeatureTeamConfigFile, err := helper.AddAWSConfigProfileForFeatureTeam(providedFlags.team)
		if err != nil {
			return err
		}
		defer os.Remove(awsFeatureTeamConfigFile)
	}
	// create the Vault client
	config, err := helper.GetVaultConfig(providedFlags.platform, providedFlags.environment, providedFlags.team)
	if err != nil {
		return err
	}
	clients, err := helper.VaultLoginWithAwsAndGetClients(config)
	if err != nil {
		return err
	}

	result, err := compareSecrets(clients, config.SecretPrefix+"/"+providedFlags.path, providedFlags.baseVersion, providedFlags.diffVersion)
	if err != nil {
		return err
	}

	cli.Println(result)
	return nil
}

func compareSecrets(clients []*api.Client, secretPath string, baseVersion, diffVersion int) (string, error) {
	if baseVersion == diffVersion {
		return "", errSameVersionDiff
	}

	baseSecret, err := readSecret(clients, secretPath, baseVersion)
	if err != nil {
		return "", err
	}

	diffSecret, err := readSecret(clients, secretPath, diffVersion)
	if err != nil {
		return "", err
	}

	// if provided base version != diff version and if base version or diff version is 0 (latest), we need to check that latest is not the same version as the other provided one
	if baseSecret.Data["metadata"].(map[string]interface{})["version"] == diffSecret.Data["metadata"].(map[string]interface{})["version"] {
		return "", errSameVersionDiff
	}

	baseSecretYaml, err := yaml.Marshal(baseSecret.Data["data"])
	if err != nil {
		return "", err
	}

	diffSecretYaml, err := yaml.Marshal(diffSecret.Data["data"])
	if err != nil {
		return "", err
	}

	f1, err := os.CreateTemp("", "yak-kube-secret-fmt-diff")
	if err != nil {
		return "", err
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := os.CreateTemp("", "yak-kube-secret-fmt-diff")
	if err != nil {
		return "", err
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	_, err = f1.Write(baseSecretYaml)
	if err != nil {
		return "", err
	}

	_, err = f2.Write(diffSecretYaml)
	if err != nil {
		return "", err
	}

	data, err := exec.Command("diff", "--label=base-version", "--label=diff-version", "-u", f1.Name(), f2.Name()).CombinedOutput() //#nosec
	if len(data) > 0 {
		return string(data), nil
	}

	return "", err
}
