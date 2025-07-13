package certificate

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type CertDetailsFromVaultSecret struct {
	Version  int
	Platform string
	Path     string
}
type CertDetailsFromVault struct {
	VaultSecret         *CertDetailsFromVaultSecret
	KeyMatchCertificate bool
	SubjectCN           string
	Issuer              string
	NotBefore           string
	NotAfter            string
	SerialNumber        string
	Hostnames           []string
}

func (d *CertDetailsFromVault) ToYaml() ([]byte, error) {
	yamlData, err := yaml.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal certificate details to YAML: %v", err)
	}
	return yamlData, nil
}

func DiffCertificateDetails(certDetails, certDetailsBase *CertDetailsFromVault) (string, error) {
	y, err := certDetails.ToYaml()
	if err != nil {
		return "", nil
	}
	tmpFile, err := os.CreateTemp("", "certDetails.yaml")
	if err != nil {
		return "", err
	}
	if _, err := tmpFile.WriteString(string(y)); err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())
	yBase, err := certDetailsBase.ToYaml()
	if err != nil {
		return "", nil
	}
	tmpFileBase, err := os.CreateTemp("", "certDetailsBase.yaml")
	if err != nil {
		return "", err
	}
	if _, err := tmpFileBase.WriteString(string(yBase)); err != nil {
		return "", err
	}
	defer os.Remove(tmpFileBase.Name())

	// Use diff command to compare the two YAML outputs
	diffCmd := exec.Command("diff", "-aburN", tmpFileBase.Name(), tmpFile.Name()) //#nosec
	log.Println(diffCmd.String())
	diffOutput, _ := diffCmd.CombinedOutput()
	return string(diffOutput), nil
}

func (config *CertConfig) CertificateDescribeSecret(selectedVersion int) (*CertDetailsFromVault, error) {
	// Get the latest version of the secret
	if selectedVersion == 0 {
		v, err := getValueFromSecret(config.Secret, "version", "metadata", 0)
		if err != nil {
			return nil, err
		}
		selectedVersion, err = strconv.Atoi(v.(json.Number).String())
		if err != nil {
			return nil, err
		}
	}

	// Get the certificate from the secret
	cert, err := getValueFromSecret(config.Secret, config.Secret.Keys.Certificate, "data", selectedVersion)
	if err != nil {
		return nil, err
	}

	// Get the private key from the secret
	privateKey, err := getValueFromSecret(config.Secret, config.Secret.Keys.PrivateKey, "data", selectedVersion)
	if err != nil {
		return nil, err
	}

	// Check that the certificate match the private_key
	err = comparePublicKeys(privateKey.(string), cert.(string))
	if err != nil {
		return nil, err
	}
	keyMatchCertificate := true

	parsedCert, err := parseCertificate(cert.(string))
	if err != nil {
		return nil, err
	}

	return &CertDetailsFromVault{
		VaultSecret: &CertDetailsFromVaultSecret{
			Version:  selectedVersion,
			Platform: config.Secret.Platform,
			Path:     config.Secret.Path,
		},
		KeyMatchCertificate: keyMatchCertificate,
		SubjectCN:           parsedCert.Subject.CommonName,
		Issuer:              parsedCert.Issuer.CommonName,
		NotBefore:           parsedCert.NotBefore.String(),
		NotAfter:            parsedCert.NotAfter.String(),
		SerialNumber:        parsedCert.SerialNumber.String(),
		Hostnames:           parsedCert.DNSNames,
	}, nil
}

func certificateDescribe(_ *cobra.Command, _ []string) error {
	if providedFlags.certificate == "" {
		return errCertificateCantBeEmpty
	}

	// Find the certificate configuration
	config, err := getCertificateConfig()
	if err != nil {
		return err
	}
	if providedFlags.diffVersion > 0 {
		certDetails, err := config.CertificateDescribeSecret(providedFlags.version)
		if err != nil {
			return err
		}
		certDetailsDiff, err := config.CertificateDescribeSecret(providedFlags.diffVersion)
		if err != nil {
			return err
		}
		diffOutput, err := DiffCertificateDetails(certDetails, certDetailsDiff)
		if err != nil {
			fmt.Println(diffOutput)
			return err
		}
		fmt.Println(diffOutput)
	} else {
		certDetails, err := config.CertificateDescribeSecret(providedFlags.version)
		if err != nil {
			return err
		}

		yamlOutput, err := certDetails.ToYaml()
		if err != nil {
			return fmt.Errorf("failed to marshal certificate details to YAML: %v", err)
		}

		// Print the YAML output
		fmt.Println(string(yamlOutput))
	}

	return nil
}

var certificateDescribeSecretCmd = &cobra.Command{
	Use:   "describe-secret",
	Short: "Describe a X509 certificate stored in vault",
	RunE:  certificateDescribe,
	PreRun: func(cmd *cobra.Command, args []string) {
		if err := cmd.Parent().MarkPersistentFlagRequired("certificate"); err != nil {
			panic(err)
		}
	},
	Args: cobra.ExactArgs(0),
}

func init() {
	certificateDescribeSecretCmd.PersistentFlags().StringVarP(&providedFlags.certificate, "certificate", "C", "", "describe certificate stored in vault")
	certificateDescribeSecretCmd.PersistentFlags().IntVarP(&providedFlags.version, "version", "v", 0, "version of the secret to get, default to latest")
	certificateDescribeSecretCmd.PersistentFlags().IntVarP(&providedFlags.diffVersion, "diff-version", "", 0, "version of the secret to diff from")
}
