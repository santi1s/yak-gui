package terraform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/santi1s/yak/cli"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/spf13/cobra"
)

var (
	reportCmd = &cobra.Command{
		Use:   "report",
		Short: "Generate providers and modules report",
		RunE:  report,
	}

	errDDEnvNotSet     = errors.New("datadog DD_API_KEY is not set")
	errDDMetricNotSent = errors.New("provider version metrics could not be sent to datadog")
)

func cleanVersion(version string) string {
	//comma
	version = strings.ReplaceAll(version, ", ", "/")

	//greater than
	version = strings.ReplaceAll(version, ">", "gt")

	//lower than
	version = strings.ReplaceAll(version, "<", "lt")

	//equal
	version = strings.ReplaceAll(version, "=", "eq")

	return version
}

func report(cmd *cobra.Command, args []string) error {
	if providedFlags.dryRun {
		cli.Println("[DRY-RUN] Running report in dry-run: no metric will be sent to Datadog")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	pathsToCheck, err := GetTerraformDirs(cwd)
	if err != nil {
		return fmt.Errorf("error getting directories containing terraform files")
	}

	re := regexp.MustCompile(`.*workspaces\s*{[^}]*name\s*=\s*"([^"]*).*`)
	for path := range pathsToCheck {
		readFile, err := os.ReadFile(path + "/backend.tf")
		if os.IsNotExist(err) {
			moduleRepositoryName := filepath.Base(path)
			if strings.HasPrefix(moduleRepositoryName, "terraform-") {
				cli.Printf("Module: %s\n", moduleRepositoryName)
				getMetrics(moduleRepositoryName, path, false)
			} else {
				cli.Printf("Ignoring %s as it's not a TFE module.\n", path)
			}
		} else if err != nil {
			_, _ = cli.PrintlnErr(err)
			continue
		} else {
			if len(re.FindStringSubmatch(string(readFile))) > 0 {
				workspaceName := strings.TrimSpace(re.FindStringSubmatch(string(readFile))[1])
				cli.Printf("Workspace: %s\n", workspaceName)
				getMetrics(workspaceName, path, true)
			} else {
				cli.Printf("Ignoring %s as it's not a TFE workspace.\n", path)
			}
		}

		cli.Println("")
	}

	return nil
}

func getMetrics(name string, path string, isWorkspace bool) {
	modules, diags := tfconfig.LoadModule(path)
	if diags.HasErrors() {
		_, _ = cli.PrintfErr("error reading module %s: %s", path, diags)
	}

	for _, mc := range modules.ModuleCalls {
		m := Module{
			Name:    mc.Name,
			Source:  mc.Source,
			Version: mc.Version,
		}

		cli.Printf("\tModule: %s - Version: %s - Resource: %s\n", m.ModuleName(), m.Version, m.Name)
		if !providedFlags.dryRun {
			err := sendModuleMetrics(name, m, isWorkspace)
			if err != nil {
				_, _ = cli.PrintErr(err)
			}
		}
	}

	for _, rp := range modules.RequiredProviders {
		// Submit metrics returns "Payload accepted" response
		if rp.Source != "" {
			cli.Printf("\tProvider: %s - Version constraint: %s\n", rp.Source, cleanVersion(rp.VersionConstraints[0]))
			if !providedFlags.dryRun {
				err := sendProviderMetrics(name, rp.Source, cleanVersion(rp.VersionConstraints[0]), isWorkspace)
				if err != nil {
					_, _ = cli.PrintErr(err)
				}
			}
		}
	}
}

func sendModuleMetrics(name string, m Module, isWorkspace bool) error {
	var errRet error
	// Set datadog creds in context
	apiKeyAuth := os.Getenv("DD_API_KEY")
	if apiKeyAuth != "" {
		ctx := context.WithValue(
			context.Background(),
			datadog.ContextAPIKeys,
			map[string]datadog.APIKey{
				"apiKeyAuth": {
					Key: apiKeyAuth,
				},
			},
		)

		metric := "tfe.tfe_module_modules_versions"
		tagName := "module_repository_name"
		if isWorkspace {
			metric = "tfe.tfe_workspace_modules_versions"
			tagName = "workspace_name"
		}

		// Set the body of metric
		body := datadogV2.MetricPayload{
			Series: []datadogV2.MetricSeries{
				{
					Metric: metric,
					Type:   datadogV2.METRICINTAKETYPE_UNSPECIFIED.Ptr(), //nolint:golint,nosnakecase
					Points: []datadogV2.MetricPoint{
						{
							Timestamp: datadog.PtrInt64(time.Now().Unix()),
							Value:     datadog.PtrFloat64(1),
						},
					},
					Resources: []datadogV2.MetricResource{
						{
							Name: datadog.PtrString(name),
							Type: datadog.PtrString(tagName),
						},
						{
							Name: datadog.PtrString(m.ModuleName()),
							Type: datadog.PtrString("module_name"),
						},
						{
							Name: datadog.PtrString(m.Version),
							Type: datadog.PtrString("module_version"),
						},
						{
							Name: datadog.PtrString(m.Name),
							Type: datadog.PtrString("module_resource"),
						},
					},
				},
			},
		}
		configuration := datadog.NewConfiguration()
		apiClient := datadog.NewAPIClient(configuration)
		api := datadogV2.NewMetricsApi(apiClient)
		_, _, err := api.SubmitMetrics(ctx, body, *datadogV2.NewSubmitMetricsOptionalParameters())
		if err != nil {
			errRet = errDDMetricNotSent
		}
	} else {
		errRet = errDDEnvNotSet
	}
	return errRet
}

func sendProviderMetrics(name, providerName, providerVersion string, isWorkspace bool) error {
	var errRet error
	// Set datadog creds in context
	apiKeyAuth := os.Getenv("DD_API_KEY")
	if apiKeyAuth != "" {
		ctx := context.WithValue(
			context.Background(),
			datadog.ContextAPIKeys,
			map[string]datadog.APIKey{
				"apiKeyAuth": {
					Key: apiKeyAuth,
				},
			},
		)

		metric := "tfe.tfe_module_providers_versions"
		tagName := "module_repository_name"
		if isWorkspace {
			metric = "tfe.tfe_workspace_providers_versions"
			tagName = "workspace_name"
		}

		// Set the body of metric
		body := datadogV2.MetricPayload{
			Series: []datadogV2.MetricSeries{
				{
					Metric: metric,
					Type:   datadogV2.METRICINTAKETYPE_UNSPECIFIED.Ptr(), //nolint:golint,nosnakecase
					Points: []datadogV2.MetricPoint{
						{
							Timestamp: datadog.PtrInt64(time.Now().Unix()),
							Value:     datadog.PtrFloat64(1),
						},
					},
					Resources: []datadogV2.MetricResource{
						{
							Name: datadog.PtrString(name),
							Type: datadog.PtrString(tagName),
						},
						{
							Name: datadog.PtrString(providerName),
							Type: datadog.PtrString("provider_name"),
						},
						{
							Name: datadog.PtrString(providerVersion),
							Type: datadog.PtrString("provider_version"),
						},
					},
				},
			},
		}
		configuration := datadog.NewConfiguration()
		apiClient := datadog.NewAPIClient(configuration)
		api := datadogV2.NewMetricsApi(apiClient)
		_, _, err := api.SubmitMetrics(ctx, body, *datadogV2.NewSubmitMetricsOptionalParameters())
		if err != nil {
			errRet = errDDMetricNotSent
		}
	} else {
		errRet = errDDEnvNotSet
	}
	return errRet
}

func init() {
	reportCmd.Flags().BoolVar(&providedFlags.dryRun, "dry-run", false, "generate the report without sending data to datadog")
}
