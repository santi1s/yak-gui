package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type YmlReferenceFile struct {
	LogicalSecret map[string]LogicalSecret `yaml:"vaultSecrets"`
}

// A struct representing a logical secret resources
type LogicalSecret struct {
	VaultNamespace string            `yaml:"vaultNamespace"`
	VaultRole      string            `yaml:"vaultRole"`
	TfeJwtSubjects []string          `yaml:"tfeJwtSubjects,omitempty"`
	Secrets        map[string]Secret `yaml:"secrets"`
	Name           string            `yaml:"-"`
}

type Secret struct {
	Keys    []string    `yaml:"keys"`
	Version interface{} `yaml:"version"`
}

func UnmarshalRefFile(path string, l *LogicalSecret) error {
	yml, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	r := &YmlReferenceFile{}
	err = yaml.Unmarshal(yml, r)
	if err != nil {
		return err
	}

	baseName := filepath.Base(path)
	fileName := baseName[:len(baseName)-len(filepath.Ext(baseName))]
	*l = r.LogicalSecret[fileName]
	l.Name = fileName

	return nil
}

// A struct to represent the usage of a secret
type SecretUsage struct {
	File           string
	SecretFolder   string
	SecretKeys     []string
	SecretPath     string
	SecretVersion  string
	VaultNamespace string
}

var errNotInRepoFolder = errors.New("are you sure to be in root folder of kube or terraform-infra repository?")
var errSecretVersionCheck = errors.New("at least one validation error occured in secret version check")
var errSecretExistenceCheck = errors.New("at least one validation error occured in secret existence check")
var errVaultRoleCheck = errors.New("at least one validation error occured in secret vault role check")
var errVaultNamespaceCheck = errors.New("at least one validation error occured in secret vault namespace check")
var errInvalidMetadataMapFromVault = errors.New("something went wrong while retrieving metadata from vault")
var errTfeJwtSubjectsCheck = errors.New("at least one validation error occured in JWT subjects check")

// getSecretUsageInYmlReferenceFiles will search in all .yml files of a given path
// It returns a list of SecretUsage and an error.
func getSecretUsageInYmlReferenceFiles(path string) ([]SecretUsage, error) {
	secretUsageList := []SecretUsage{}

	ymlReferenceFiles, err := filepath.Glob(path)
	if err != nil {
		return nil, err
	}

	for _, f := range ymlReferenceFiles {
		file, err := os.Open(f)
		if err != nil {
			return nil, err
		}

		logicalSecret := &LogicalSecret{}
		err = UnmarshalRefFile(f, logicalSecret)
		if err != nil {
			return nil, err
		}

		for k, v := range logicalSecret.Secrets {
			if len(providedRepoSecretFlags.ignoredPrefixes) > 0 && helper.HasAnyPrefix(k, providedRepoSecretFlags.ignoredPrefixes) {
				continue
			}

			var filename string
			if filepath.IsAbs(file.Name()) {
				wd, _ := os.Getwd()
				filename = strings.Replace(file.Name(), wd+"/", "", 1)
			} else {
				filename = file.Name()
			}

			var version string
			switch v.Version.(type) {
			case string:
				version = v.Version.(string)
			case int:
				version = strconv.Itoa(v.Version.(int))
			}

			namespace := filepath.Base(filepath.Dir(filename))
			folderName := ""
			if namespace == "common-shared" || namespace == "common-prod" {
				folderName = namespace
				namespace = "common"
			}

			secretUsageList = append(
				secretUsageList,
				SecretUsage{
					File:           filename,
					SecretFolder:   folderName,
					SecretKeys:     v.Keys,
					SecretPath:     k,
					SecretVersion:  version,
					VaultNamespace: namespace,
				},
			)
		}

		file.Close()
	}

	return secretUsageList, nil
}

// checkSecretsVersion verify for all secrets in secretUsageList:
// - secret always have a version
// - a secret path in the same Vault namespace used multiple time always use the same version
// - version of a secret is an integer
// - vaultRole and vaultNamespace are non empty string
// - vaultNamespace is starting with doctolib/
// It will prints validation errors found and returns an error if there is at least one validation error
func checkSecretsVersion(secretUsageList []SecretUsage) ([]string, error) {
	var errorList []string
	var checkError error

	// for each secret we will check all others secrets
	for _, secret := range secretUsageList {
		// if regexp fails, it smells like someone is trying to use latest as version
		if secret.SecretVersion == "latest" {
			line, _ := helper.FindStringInFileAndGetLineNumber(secret.File, secret.SecretPath)
			errorList = append(errorList, fmt.Sprintf("::error file=%s,line=%d::Secret path %s in namespace %s is using latest as version",
				secret.File, line, secret.SecretPath, secret.VaultNamespace))
			checkError = errSecretVersionCheck
			continue
		}

		version, err := strconv.Atoi(secret.SecretVersion)
		// check that version is an integer or it will make vault csi provider to fail even if the manifest is correct
		if err != nil {
			line, _ := helper.FindStringInFileAndGetLineNumber(secret.File, secret.SecretPath)
			errorList = append(errorList, fmt.Sprintf("::error file=%s,line=%d::Secret path %s in namespace %s is using non integer value as version",
				secret.File, line, secret.SecretPath, secret.VaultNamespace))
			checkError = errSecretVersionCheck
			continue
		}

		if version <= 0 {
			line, _ := helper.FindStringInFileAndGetLineNumber(secret.File, secret.SecretPath)
			errorList = append(errorList, fmt.Sprintf("::error file=%s,line=%d::Secret path %s in namespace %s is using version less or equal than 0",
				secret.File, line, secret.SecretPath, secret.VaultNamespace))
			checkError = errSecretVersionCheck
			continue
		}

		for _, secret2 := range secretUsageList {
			// check that versions are matching
			if secret.VaultNamespace == secret2.VaultNamespace &&
				secret.SecretPath == secret2.SecretPath &&
				secret.SecretFolder == secret2.SecretFolder &&
				secret.SecretVersion != secret2.SecretVersion {
				line, _ := helper.FindStringInFileAndGetLineNumber(secret.File, secret.SecretPath)
				errorList = append(errorList, fmt.Sprintf("::error file=%s,line=%d::Secret path %s in namespace %s is using version %s while using version %s in manifest %s",
					secret.File, line, secret.SecretPath, secret.VaultNamespace, secret.SecretVersion, secret2.SecretVersion, secret2.File))
				checkError = errSecretVersionCheck
			}
		}
	}

	return errorList, checkError
}

// checkSecretsVersion verify for all secrets in secretUsageList are existing in Vault clusters
// It will prints validation errors found and returns an error if there is at least one validation error
func checkSecretsExistence(secretUsageList []SecretUsage) ([]string, error) {
	var errorList []string
	var checkError error
	var err error

	clients := make(map[string][]*api.Client)
	config := make(map[string]*helper.VaultConfig)
	for _, secret := range secretUsageList {
		if _, exists := clients[secret.VaultNamespace]; !exists {
			config[secret.VaultNamespace], err = helper.GetVaultConfig(secret.VaultNamespace, "")
			if err != nil {
				return nil, err
			}
			clients[secret.VaultNamespace], err = helper.VaultLoginWithAwsAndGetClients(config[secret.VaultNamespace])
			if err != nil {
				return nil, err
			}
		}

		s, err := clients[secret.VaultNamespace][0].Logical().Read("kv/metadata/" + secret.SecretPath)
		if err != nil {
			return nil, errors.New("error while reading secret on " + clients[secret.VaultNamespace][0].Address() + ": " + err.Error())
		}

		if s == nil { // secret does not exist
			line, _ := helper.FindStringInFileAndGetLineNumber(secret.File, secret.SecretPath)
			errorList = append(errorList, fmt.Sprintf("::error file=%s,line=%d::Secret path %s in namespace %s does not exist on vault",
				secret.File, line, secret.SecretPath, secret.VaultNamespace))
			checkError = errSecretExistenceCheck
		} else {
			if versions, ok := s.Data["versions"].(map[string]interface{}); ok {
				if version, ok := versions[secret.SecretVersion].(map[string]interface{}); ok {
					if deletionTime, ok := version["deletion_time"].(string); ok {
						if deletionTime != "" {
							line, _ := helper.FindStringInFileAndGetLineNumber(secret.File, secret.SecretPath)
							errorList = append(errorList, fmt.Sprintf("::error file=%s,line=%d::Secret path %s in namespace %s using version %s that has been deleted from vault",
								secret.File, line, secret.SecretPath, secret.VaultNamespace, secret.SecretVersion))
							checkError = errSecretExistenceCheck
							continue
						}
					}
					if destroyed, ok := version["destroyed"].(bool); ok {
						if destroyed {
							line, _ := helper.FindStringInFileAndGetLineNumber(secret.File, secret.SecretPath)
							errorList = append(errorList, fmt.Sprintf("::error file=%s,line=%d::Secret path %s in namespace %s using version %s that has been destroyed from vault",
								secret.File, line, secret.SecretPath, secret.VaultNamespace, secret.SecretVersion))
							checkError = errSecretExistenceCheck
							continue
						}
					}
				} else {
					line, _ := helper.FindStringInFileAndGetLineNumber(secret.File, secret.SecretPath)
					errorList = append(errorList, fmt.Sprintf("::error file=%s,line=%d::Secret path %s in namespace %s using version %s does not exist on vault",
						secret.File, line, secret.SecretPath, secret.VaultNamespace, secret.SecretVersion))
					checkError = errSecretExistenceCheck
					continue
				}
			} else {
				return nil, errInvalidMetadataMapFromVault
			}

			data := map[string][]string{
				"version": {secret.SecretVersion},
			}

			ciSecret, err := clients[secret.VaultNamespace][0].Logical().ReadWithData("kv/data/ci/"+secret.SecretPath, data)
			if err != nil {
				return nil, errors.New("error while reading secret on " + clients[secret.VaultNamespace][0].Address() + ": " + err.Error())
			}

			if ciSecret != nil {
				for _, v := range secret.SecretKeys {
					if ciSecret.Data["data"].(map[string]interface{}) == nil {
						cli.Printf("secret.SecretKeys: %s", v)
					} else {
						if _, exists := ciSecret.Data["data"].(map[string]interface{})[v]; !exists {
							line, _ := helper.FindStringInFileAndGetLineNumber(secret.File, v)
							errorList = append(errorList, fmt.Sprintf("::error file=%s,line=%d::Secret key %s in secret path %s using verson %s in namespace %s does not exist on vault",
								secret.File, line, v, secret.SecretPath, secret.SecretVersion, secret.VaultNamespace))
							checkError = errSecretExistenceCheck
						}
					}
				}
			} else {
				return nil, errors.New("secret ci/" + secret.SecretPath + " in version " + secret.SecretVersion + " not found in " + secret.VaultNamespace)
			}
		}
	}

	return errorList, checkError
}

func checkVaultRole(path string) ([]string, error) {
	var errorList []string
	var checkError error

	ymlReferenceFiles, err := filepath.Glob(path)
	if err != nil {
		return nil, err
	}

	for _, f := range ymlReferenceFiles {
		file, err := os.Open(f)
		if err != nil {
			return nil, err
		}

		logicalSecret := LogicalSecret{}
		err = UnmarshalRefFile(f, &logicalSecret)
		if err != nil {
			return nil, err
		}

		file.Close()

		if logicalSecret.VaultRole == "" {
			line, _ := helper.FindStringInFileAndGetLineNumber(file.Name(), "vaultRole")
			errorList = append(errorList, fmt.Sprintf("::error file=%s,line=%d::Vault role must be a non empty string",
				file.Name(), line))
			checkError = errVaultRoleCheck
		}
	}

	return errorList, checkError
}

func checkVaultNamespace(path string) ([]string, error) {
	var errorList []string
	var checkError error

	ymlReferenceFiles, err := filepath.Glob(path)
	if err != nil {
		return nil, err
	}

	for _, f := range ymlReferenceFiles {
		file, err := os.Open(f)
		if err != nil {
			return nil, err
		}

		logicalSecret := LogicalSecret{}
		err = UnmarshalRefFile(f, &logicalSecret)
		if err != nil {
			return nil, err
		}

		file.Close()

		if !strings.HasPrefix(logicalSecret.VaultNamespace, "doctolib/") {
			line, _ := helper.FindStringInFileAndGetLineNumber(file.Name(), "vaultNamespace")
			errorList = append(errorList, fmt.Sprintf("::error file=%s,line=%d::Vault namespace must be a non empty string starting by doctolib/",
				file.Name(), line))
			checkError = errVaultNamespaceCheck
		}
	}

	return errorList, checkError
}

// checkTfeJwtSubjects verifies JWT subjects format and alphabetical sorting
func checkTfeJwtSubjects(path string) ([]string, error) {
	var errorList []string
	var checkError error

	ymlReferenceFiles, err := filepath.Glob(path)
	if err != nil {
		return nil, err
	}

	for _, f := range ymlReferenceFiles {
		file, err := os.Open(f)
		if err != nil {
			return nil, err
		}

		logicalSecret := LogicalSecret{}
		err = UnmarshalRefFile(f, &logicalSecret)
		if err != nil {
			return nil, err
		}

		file.Close()

		// Check JWT subjects if they exist
		if len(logicalSecret.TfeJwtSubjects) > 0 {
			// Check if JWT subjects are sorted alphabetically
			sorted := make([]string, len(logicalSecret.TfeJwtSubjects))
			copy(sorted, logicalSecret.TfeJwtSubjects)
			sort.Strings(sorted)

			for i, subject := range logicalSecret.TfeJwtSubjects {
				if subject != sorted[i] {
					line, _ := helper.FindStringInFileAndGetLineNumber(file.Name(), "tfeJwtSubjects")
					errorList = append(errorList, fmt.Sprintf("::error file=%s,line=%d::JWT subjects are not sorted alphabetically",
						file.Name(), line))
					checkError = errTfeJwtSubjectsCheck
					break
				}
			}

			// Check format of each JWT subject
			for _, subject := range logicalSecret.TfeJwtSubjects {
				if err := validateTfeJwtSubjectFormat(subject); err != nil {
					line, _ := helper.FindStringInFileAndGetLineNumber(file.Name(), subject)
					errorList = append(errorList, fmt.Sprintf("::error file=%s,line=%d::Invalid TFE JWT subject format: %s",
						file.Name(), line, subject))
					checkError = errTfeJwtSubjectsCheck
				}
			}
		}
	}

	return errorList, checkError
}

func secretCheck(cmd *cobra.Command, args []string) error {
	if !providedRepoSecretFlags.checkVaultRole && !providedRepoSecretFlags.checkVaultNamespace && !providedRepoSecretFlags.checkExistence && !providedRepoSecretFlags.checkVersion && !providedRepoSecretFlags.checkTfeJwtSubjects {
		providedRepoSecretFlags.checkAll = true
	}

	err := checkRepositoryPath()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	secretsUsageList, err := getSecretUsageInYmlReferenceFiles(cwd + "/configs/vault-secrets/*/*.yml")
	if err != nil {
		return err
	}
	if len(secretsUsageList) == 0 {
		return errNotInRepoFolder
	}

	if providedRepoSecretFlags.checkVaultRole || providedRepoSecretFlags.checkAll {
		cli.Println("Executing secrets vault role check...")
		errorList, err := checkVaultRole(cwd + "/configs/vault-secrets/*/*.yml")
		for _, e := range errorList {
			cli.Println(e)
		}
		if err != nil {
			return err
		}
		cli.Print("\n")
	}

	if providedRepoSecretFlags.checkVaultNamespace || providedRepoSecretFlags.checkAll {
		cli.Println("Executing secrets vault namespace check...")
		errorList, err := checkVaultNamespace(cwd + "/configs/vault-secrets/*/*.yml")
		for _, e := range errorList {
			cli.Println(e)
		}
		if err != nil {
			return err
		}
		cli.Print("\n")
	}

	if providedRepoSecretFlags.checkVersion || providedRepoSecretFlags.checkAll {
		cli.Println("Executing secrets version usage check...")
		errorList, err := checkSecretsVersion(secretsUsageList)
		for _, e := range errorList {
			cli.Println(e)
		}
		if err != nil {
			return err
		}
		cli.Print("\n")
	}

	if providedRepoSecretFlags.checkExistence || providedRepoSecretFlags.checkAll {
		cli.Println("Executing secrets existence check...")
		errorList, err := checkSecretsExistence(secretsUsageList)
		for _, e := range errorList {
			cli.Println(e)
		}
		if err != nil {
			return err
		}
		cli.Print("\n")
	}

	if providedRepoSecretFlags.checkTfeJwtSubjects || providedRepoSecretFlags.checkAll {
		cli.Println("Executing JWT subjects check...")
		errorList, err := checkTfeJwtSubjects(cwd + "/configs/vault-secrets/*/*.yml")
		for _, e := range errorList {
			cli.Println(e)
		}
		if err != nil {
			return err
		}
		cli.Print("\n")
	}

	cli.Println("secrets check passed!")
	return nil
}

var secretCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "execute ci check for secrets in kube repository",
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.Parent().ResetFlags()
		cmd.SilenceUsage = true
	},
	RunE: secretCheck,
	Args: cobra.ExactArgs(0),
}

func init() {
	secretCheckCmd.Flags().BoolVar(&providedRepoSecretFlags.checkVaultRole, "check-vault-role", false, "check that all logical secret resources have a correct value for vault role")
	secretCheckCmd.Flags().BoolVar(&providedRepoSecretFlags.checkVaultNamespace, "check-vault-namespace", false, "check that all logical secret resources have a correct value for vault namespace")
	secretCheckCmd.Flags().BoolVar(&providedRepoSecretFlags.checkVersion, "check-version", false, "check that a single secret path always use the same version and that no latest version is used")
	secretCheckCmd.Flags().BoolVar(&providedRepoSecretFlags.checkExistence, "check-existence", false, "check that all secret path and associated version are existing")
	secretCheckCmd.Flags().BoolVar(&providedRepoSecretFlags.checkTfeJwtSubjects, "check-tfe-jwt-subjects", false, "check that JWT subjects are properly formatted and sorted alphabetically")
	secretCheckCmd.Flags().BoolVar(&providedRepoSecretFlags.checkAll, "check-all", false, "execute all checks")
	secretCheckCmd.Flags().StringSliceVarP(&providedRepoSecretFlags.ignoredPrefixes, "ignored-prefixes", "i", []string{}, "ignore secret paths starting with this prefix")
	secretCheckCmd.MarkFlagsMutuallyExclusive("check-all", "check-version")
	secretCheckCmd.MarkFlagsMutuallyExclusive("check-all", "check-existence")
	secretCheckCmd.MarkFlagsMutuallyExclusive("check-all", "check-tfe-jwt-subjects")
}
