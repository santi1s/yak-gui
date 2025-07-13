package certificate

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/doctolib/yak/internal/helper"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Generate a terraform resource for one cloudflare record
func buildCloudflareRecord(resourceName string, zoneID string, name string, value string) *hclwrite.Block {
	record := hclwrite.NewBlock("resource", []string{"cloudflare_record", resourceName})
	record.Body().SetAttributeTraversal("zone_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: zoneID,
		},
	})
	record.Body().SetAttributeValue("name", cty.StringVal(name))
	record.Body().SetAttributeValue("value", cty.StringVal(value))
	record.Body().SetAttributeValue("type", cty.StringVal("CNAME"))
	record.Body().SetAttributeValue("ttl", cty.NumberIntVal(120))
	return record
}

// Generate a terraform resource for one route53 record
func buildRoute53Record(resourceName string, zoneID string, name string, value string) *hclwrite.Block {
	record := hclwrite.NewBlock("resource", []string{"aws_route53_record", resourceName})
	record.Body().SetAttributeTraversal("zone_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: zoneID,
		},
	})
	record.Body().SetAttributeValue("name", cty.StringVal(name))
	record.Body().SetAttributeValue("records", cty.ListVal([]cty.Value{
		cty.StringVal(value)}))
	record.Body().SetAttributeValue("type", cty.StringVal("CNAME"))
	record.Body().SetAttributeValue("ttl", cty.NumberIntVal(120))
	return record
}

func searchAndPatchRecord(file *hclwrite.File, resourceName string, newResource hclwrite.Block) bool {
	for _, block := range file.Body().Blocks() {
		labels := block.Labels()
		if block.Type() == "resource" {
			if (labels[0] == "cloudflare_record" || labels[0] == "aws_route53_record") && labels[1] == resourceName {
				*block = newResource
				return true
			}
		}
	}
	return false
}

func openTFfile(filename string) (*hclwrite.File, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	file, diags := hclwrite.ParseConfig(content, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, errors.New("an error occurred")
	}
	return file, nil
}

func writeTFfile(filename string, file *hclwrite.File) error {
	if err := os.WriteFile(filename, file.Bytes(), 0600); err != nil {
		return err
	}
	return nil
}

func sanitizeValue(value string) string {
	return RemoveFinalDot(strings.ToLower(value))
}

type DCVRecord struct {
	// Name of the record, something like _8F84AA9A870742E368258A82DBFB8C70.doctolib.tech.
	Name string
	// Value of the record, something like 8f84aa9a870742e368258a82dbfb8c70.223720d98a6c295acbc6170754a0a9ef.9e3aaf2e75599987dc69.sectigo.com.
	Value string
}

func getDCVRecords(client *Gandi, config *CertConfig) ([]DCVRecord, error) {
	var records []DCVRecord
	// Get the certificate by tag
	certificate, err := client.GetCertificateBy(map[string]interface{}{
		"Tags":  config.Tags,
		"State": "being_renew",
	})
	if err != nil {
		return nil, err
	}
	// Get Domain Control Validation (DCV) records
	dcvDetails, err := client.GetDomainValidationDetails(certificate.ID, DomainValidationDetailsRequest{
		CSR:       certificate.CSR,
		DCVMethod: "dns",
	})
	if err != nil {
		return nil, err
	}
	rawRecords := dcvDetails.RawMessages
	if len(rawRecords) == 0 {
		return nil, fmt.Errorf("[Gandi] no DCV records found")
	}
	for _, record := range rawRecords {
		name := record[0]
		value := record[1]
		records = append(records, DCVRecord{
			Name:  name,
			Value: value,
		})
	}
	return records, nil
}

func RemoveFinalDot(name string) string {
	return strings.TrimSuffix(name, ".")
}

func (record *DCVRecord) generateTFResourceName() string {
	return "gandi-dcv-" + strings.ReplaceAll(RemoveFinalDot(record.domain()), ".", "_")
}

func (record *DCVRecord) domain() string {
	// remove first part of the name
	return RemoveFinalDot(strings.Join(strings.Split(record.Name, ".")[1:], "."))
}

func removeDomain(fqdn string) string {
	w := strings.Split(fqdn, ".")
	var remove = 2
	if w[len(w)-1] == "" {
		remove = 3
	}
	return strings.Join(strings.Split(fqdn, ".")[:len(w)-remove], ".")
}

// Generate every terraform resource for one record
func buildDCVRecords(config *CertConfig, dvcRecords []DCVRecord, hclFile *hclwrite.File) error {
	// Generate a terraform resource for each record
	for _, dcvRecord := range dvcRecords {
		name := dcvRecord.Name
		value := dcvRecord.Value

		resourceName := dcvRecord.generateTFResourceName()
		var newRecord *hclwrite.Block
		dnsProvider, err := config.DNSProvider()
		if err != nil {
			return err
		}
		switch dnsProvider {
		case "cloudflare":
			newRecord = buildCloudflareRecord(
				resourceName,
				config.Cloudflare.Zone,
				RemoveFinalDot(removeDomain(name)),
				sanitizeValue(value),
			)
		case "route53":
			newRecord = buildRoute53Record(
				resourceName,
				config.Route53.Zone,
				RemoveFinalDot(name),
				sanitizeValue(value))
		default:
			return fmt.Errorf("unsupported DNS provider: %s", dnsProvider)
		}

		patched := searchAndPatchRecord(hclFile, resourceName, *newRecord)
		if !patched {
			hclFile.Body().AppendNewline()
			hclFile.Body().AppendBlock(newRecord)
		}
	}
	return nil
}

// Add records for Domain Control Validation (DCV) by Gandi
// Open a pull request
func dcvRecordsPullRequest(client *Gandi, config *CertConfig) error {
	// Start a new branch
	branch := fmt.Sprintf("cloudflare_dcv_record_%s-%d",
		strings.ReplaceAll(config.Name, ".", "_"),
		time.Now().UnixMilli())
	tmpDir := os.TempDir()
	repoDir, err := os.MkdirTemp(tmpDir, "renew-cert")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(repoDir)
	log.Println("Cloning the repository terraform-infra to " + repoDir + "/terraform-infra")
	log.Println("Checkout branch " + branch)
	repository, worktree, err := helper.CloneRepositoryAndGetWorktree("terraform-infra", branch, repoDir, "main", 1)
	if err != nil {
		return err
	}
	tfFile, err := config.GetTfFilePath()
	if err != nil {
		return err
	}
	tfFileInRepo := path.Join(repoDir, "terraform-infra", tfFile)

	log.Println("Building the DCV records")
	hclFile, err := openTFfile(tfFileInRepo)
	if err != nil {
		return err
	}
	dcvRecords, err := getDCVRecords(client, config)
	if err != nil {
		return err
	}
	if err := buildDCVRecords(config, dcvRecords, hclFile); err != nil {
		return err
	}

	if err := writeTFfile(tfFileInRepo, hclFile); err != nil {
		return err
	}

	if providedFlags.dryRun {
		log.Println("Dry run: terraform resources will not be committed: ")
		_, err := exec.Command("git", "diff", tfFileInRepo).CombinedOutput()
		if err != nil {
			return err
		}
		return nil
	}
	err = helper.CommitAll(repository, worktree, "DCV records for Gandi Domain Control Validation")
	if err != nil {
		return err
	}

	log.Println("Creating the Pull Request")

	pr, err := helper.CreatePullRequest("terraform-infra", branch, "main",
		fmt.Sprintf("chore(%s): Add DCV records for cert %s", providedFlags.jiraTicket, config.Name),
		"Add DCV records for Gandi DCV (Domain Control Validation) for certificate "+config.Name,
		fmt.Sprintf("%s - Automated PR produced by 'yak certificate renew'", helper.JiraLink(providedFlags.jiraTicket)), "", false)
	if err != nil {
		return err
	}

	err = helper.CreateComment("terraform-infra", pr, "/fix_everything")
	if err != nil {
		return err
	}

	log.Println("Pull Request created: " + fmt.Sprintf("https://github.com/doctolib/terraform-infra/pull/%d\n", pr.GetNumber()))
	log.Println("Please review it and merge it to main before going further")

	// TODO Revert PR
	return nil
}

func certificateRenew(cmd *cobra.Command, args []string) error {
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

	log.Println("Starting the renewal of " + config.Name)

	// Get the certificate by tag
	certificate, err := client.GetCertificateBy(map[string]interface{}{
		"Tags":      config.Tags,
		"Status":    "valid",
		"Renewable": true,
	})

	if err != nil {
		return err
	}

	// Get the private key from the secret
	privateKey, err := getValueFromSecret(config.Secret, config.Secret.Keys.PrivateKey, "data", 0)
	if err != nil {
		return err
	}

	// Generate CSR
	csrConfig := path.Join(os.Getenv("TFINFRA_REPOSITORY_PATH"), config.Conf)
	csr, err := generateCSR(privateKey.(string), csrConfig)
	if err != nil {
		return err
	}

	if providedFlags.dryRun {
		log.Println("Dry run, do not ask to renew the certificate")
	} else if !providedFlags.dcvOnly {
		log.Println("Ask to the issuer for a renewal")
		response, err := client.Renew(certificate.ID, RenewCertificateRequest{
			CSR:       csr,
			DCVMethod: "dns",
		})

		if err != nil {
			return err
		}

		if response.Code != 0 && response.Code != 200 && response.Code != 202 {
			return errors.New("An error occurred :" + strconv.Itoa(response.Code) + " - " + response.Cause + " - " + response.Message)
		}

		// Get the new certificate details
		newCertificate, err := client.GetCertificateBy(map[string]interface{}{
			"CN":     certificate.CN,
			"Status": "pending",
			"State":  "being_renew",
		})

		if err != nil {
			return err
		}

		// Attach tags on the new certificates
		log.Println("Attach tags on new certificate")
		response, err = client.AttachTags(newCertificate.ID, AttachTagsRequest{Tags: config.Tags})
		if err != nil {
			return err
		}

		if response.Code != 0 && response.Code != 200 && response.Code != 202 {
			return errors.New("An error occured :" + strconv.Itoa(response.Code) + " - " + response.Cause + " - " + response.Message)
		}
	}

	// Add DCV records for domain control validation
	err = dcvRecordsPullRequest(client, config)
	if err != nil {
		return err
	}

	log.Println("Done")
	return nil
}

var certificateRenewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Ask the issuer to renew a certificate",
	RunE:  certificateRenew,
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
	certificateRenewCmd.PersistentFlags().StringVarP(&providedFlags.certificate, "certificate", "C", "", "certificate to renew")
	certificateRenewCmd.PersistentFlags().BoolVarP(&providedFlags.dcvOnly, "dcv-only", "D", false, "fetch DCV records only and open a pull request")
	certificateRenewCmd.PersistentFlags().StringVarP(&providedFlags.jiraTicket, "jira", "j", "", "Jira ticket")
	certificateRenewCmd.Flags().BoolVar(&providedFlags.dryRun, "dry-run", false, "do not create pull requests")
}
