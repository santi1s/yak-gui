package certificate

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	repoSecretHelper "github.com/santi1s/yak/internal/cmd/repo"
	secretHelper "github.com/santi1s/yak/internal/cmd/secret"
	"github.com/santi1s/yak/internal/helper"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func bumpSecretPullRequest(config *CertConfig, platform string, repo string, mainBranch string) error {
	log.Println("Preparing the PR to update the secret on " + platform)

	// Start a new branch
	branch := fmt.Sprintf("bump_cert_secret_%s-%s-%d",
		strings.ReplaceAll(config.Name, ".", "_"),
		platform,
		time.Now().UnixMilli())

	tmpDir := os.TempDir()
	repoDir, err := os.MkdirTemp(tmpDir, repo)
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(repoDir)
	log.Println("Cloning the repository " + repo + " in " + repoDir)
	repository, worktree, err := helper.CloneRepositoryAndGetWorktree(repo, branch, repoDir, mainBranch, 1)
	if err != nil {
		return err
	}

	log.Println("Bumping the secret to its latest version")

	currentVersion, err := getValueFromSecret(config.Secret, "version", "metadata", 0)
	if err != nil {
		return err
	}
	version, _ := strconv.Atoi(currentVersion.(json.Number).String())

	// Bump the secret
	secretFlags := repoSecretHelper.RepoSecretFlags{
		Platform:    config.Secret.Platform,
		Environment: config.Secret.Env,
		Path:        config.Secret.Path,
		Version:     version,
	}
	secretFlags.SecretProd = platform == "common-prod"

	err = repoSecretHelper.DoSecretBump(&secretFlags, repoDir+"/"+repo)
	if err != nil {
		return err
	}

	status, err := worktree.Status()
	if err != nil {
		return err
	}

	if status.IsClean() {
		log.Println("No changes to commit, skipping PR creation")
		return nil
	}

	err = helper.CommitAll(repository, worktree, "Bump secret for "+config.Name+"("+platform+")")
	if err != nil {
		return err
	}

	// Create the PR

	certDetails, err := config.CertificateDescribeSecret(version)
	if err != nil {
		return err
	}

	yamlOutput, err := certDetails.ToYaml()
	if err != nil {
		return fmt.Errorf("failed to marshal certificate details to YAML: %v", err)
	}

	pr, err := helper.CreatePullRequest(repo, branch, mainBranch,
		fmt.Sprintf("chore(%s): Bump secret for %s (%s)", providedFlags.jiraTicket, config.Name, platform),
		fmt.Sprintf("Bump secret for %s on platform %s\n\n%s\n\n", config.Name, platform, yamlOutput),
		fmt.Sprintf("%s - Automated PR produced by 'yak certificate refresh-secret'", helper.JiraLink(providedFlags.jiraTicket)), "", false)
	if err != nil {
		return err
	}

	err = helper.CreateComment(repo, pr, "/fix_everything")
	if err != nil {
		return err
	}

	log.Println("Pull Request created: " + fmt.Sprintf("https://github.com/doctolib/"+repo+"/pull/%d\n", pr.GetNumber()))
	log.Println("Please review it and merge it to " + mainBranch + " before going further")

	return nil
}

func updateSecret(config *CertConfig, criteria map[string]interface{}) error {
	log.Println("Updating secret")

	// Update the secret with the new certificate
	name := config.Secret.Path
	env := config.Secret.Env
	platform := config.Secret.Platform

	// sre are the only users of this task for the moment
	vaultConfig, err := helper.GetVaultConfig(platform, env)
	if err != nil {
		return err
	}
	clients, err := helper.VaultLoginWithAwsAndGetClients(vaultConfig)
	if err != nil {
		return err
	}

	if env == "" {
		env = "common"
	}
	secretPath := path.Join(env, name)
	_, err = secretHelper.PatchSecretData(clients, secretPath, criteria)
	if err != nil {
		return err
	}
	log.Println("Secret updated")
	return nil
}

func certificateRefreshSecret(_ *cobra.Command, _ []string) error {
	if providedFlags.certificate == "" {
		return errCertificateCantBeEmpty
	}
	if providedFlags.jiraTicket == "" {
		return errJiraTicketCantBeEmpty
	}
	// Find the certificate configuration
	config, err := getCertificateConfig()
	if err != nil {
		return err
	}

	client, err := GandiClient()
	if err != nil {
		return err
	}

	log.Println("Starting to refresh the secret for certificate " + config.Name)

	// Get the certificate by tag
	certificate, err := client.GetCertificateBy(map[string]interface{}{
		"Tags":      config.Tags,
		"Status":    "valid",
		"Renewable": false,
	})

	if err != nil {
		log.Println(err.Error())
		log.Println("No certificate found. checking if one is pending.")
		certificate, err := client.GetCertificateBy(map[string]interface{}{
			"Tags":   config.Tags,
			"Status": "pending",
			"State":  "being_renew",
		})
		if err != nil {
			return err
		}

		log.Println("Asking for the DCV again. Ensure that cloudflare records PR have been merged")
		_, err = client.AskDomainValidation(certificate.ID)
		if err != nil {
			return err
		}
		log.Println("Exiting. Re-run the task later when Gandi will have validated the domain.")
	}
	// Check that certificate expiration date is good (+1month)
	err = checkExpiration(certificate)
	if err != nil {
		return err
	}

	// Get the private key from the secret
	privateKey, err := getValueFromSecret(config.Secret, config.Secret.Keys.PrivateKey, "data", 0)
	if err != nil {
		return err
	}

	// Check that the certificate match the private_key
	err = comparePublicKeys(privateKey.(string), certificate.Cert)
	if err != nil {
		return err
	}

	log.Println("Combine intermediate and certificate")
	// Combine the renewed certificate with the intermediate certificate
	intermediateCert, err := client.GetIntermediate(certificate)
	if err != nil {
		return err
	}

	combinedCertificate := certificate.Cert + intermediateCert

	// Get the certificate from the secret
	oldCertificate, err := getValueFromSecret(config.Secret, config.Secret.Keys.Certificate, "data", 0)
	if err != nil {
		return err
	}

	// Check if there is a change in certificate
	shouldUpdateSecret := compareCertificates(oldCertificate.(string), combinedCertificate)

	if providedFlags.dryRun {
		log.Println("Dry run, secret will not be updated")
		return nil
	}

	if shouldUpdateSecret {
		// Update the certificate with certificate + intermediate
		criteria := map[string]interface{}{
			config.Secret.Keys.Certificate: combinedCertificate,
		}

		//  Update the intermediate certificate alone, if asked
		if config.Secret.Keys.Intermediate != "" {
			criteria[config.Secret.Keys.Intermediate] = string(intermediateCert)
		}

		err = updateSecret(config, criteria)
		if err != nil {
			return err
		}
	} else {
		log.Println("Skipping the secret update (already up to date)")
		log.Println("For idempotence, try to do the bump PR anyway")
	}

	log.Println("Preparing Bump secret PRs on kube and terraform-infra repositories")

	if config.Secret.Platform == "common" {
		// Need to separate non-prod and prod bumps for common secrets
		err = bumpSecretPullRequest(config, "common-shared", "kube", "master")
		if err != nil {
			return err
		}
		err = bumpSecretPullRequest(config, "common-shared", "terraform-infra", "main")
		if err != nil {
			return err
		}
		err = bumpSecretPullRequest(config, "common-prod", "kube", "master")
		if err != nil {
			return err
		}
		err = bumpSecretPullRequest(config, "common-prod", "terraform-infra", "main")
		if err != nil {
			return err
		}
	} else {
		err = bumpSecretPullRequest(config, config.Secret.Platform, "kube", "master")
		if err != nil {
			return err
		}
		err = bumpSecretPullRequest(config, config.Secret.Platform, "terraform-infra", "main")
		if err != nil {
			return err
		}
	}
	log.Println("Done")
	return nil
}

var certificateRefreshSecretCmd = &cobra.Command{
	Use:   "refresh-secret",
	Short: "Get the latest certificate from the provider and refresh its secret",
	RunE:  certificateRefreshSecret,
	PreRun: func(cmd *cobra.Command, args []string) {
		if err := cmd.Parent().MarkPersistentFlagRequired("certificate"); err != nil {
			panic(err)
		}
		if err := cmd.Parent().MarkPersistentFlagRequired("jira"); err != nil {
			panic(err)
		}
	},
	Args: cobra.ExactArgs(0),
}

func init() {
	certificateRefreshSecretCmd.PersistentFlags().StringVarP(&providedFlags.certificate, "certificate", "C", "", "certificate to refresh")
	certificateRefreshSecretCmd.Flags().BoolVar(&providedFlags.dryRun, "dry-run", false, "do not update secrets, do not create pull requests")
	certificateRefreshSecretCmd.PersistentFlags().StringVarP(&providedFlags.jiraTicket, "jira", "j", "", "Jira ticket")
}
