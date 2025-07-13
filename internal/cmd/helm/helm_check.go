package helm

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var (
	helmCheckCmd = &cobra.Command{
		Use:     "check",
		Short:   "Check if external Helm charts used in kube repository are allowed",
		RunE:    checkHelmCharts,
		Example: "yak helm check",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if os.Getenv("YAK_HELM_WORKDIR") == "" {
				os.Setenv("YAK_HELM_WORKDIR", os.Getenv("KUBE_REPOSITORY_PATH"))
			}
			if os.Getenv("YAK_HELM_EXPORTDIR") == "" {
				os.Setenv("YAK_HELM_EXPORTDIR", os.Getenv("KUBE_REPOSITORY_PATH"))
			}
			if cmd.Parent().PreRunE != nil {
				return cmd.Parent().PreRunE(cmd, args)
			}
			return nil
		},
	}

	AwsEcrRegexp = regexp.MustCompile(`(?P<account>[0-9]+)\.dkr\.ecr\.(?P<region>[a-z0-9-]+)\.amazonaws\.com`)

	errHelmChartsProblemFound = errors.New("found at least one problem in Helm chart definitions (see above for more details)")
)

func checkHelmCharts(cmd *cobra.Command, args []string) error {
	workDir := os.Getenv("YAK_HELM_WORKDIR")
	//exportDir := os.Getenv("YAK_HELM_EXPORTDIR")
	helmConfigPath := filepath.Join(workDir, "configs/helm")
	viper.SetConfigFile(providedHelmFlags.cfgFile)
	cobra.CheckErr(viper.MergeInConfig())
	whitelist := viper.GetStringSlice("external_helm_charts")
	platform := "*"
	env := "*"
	numberOfErrors := 0

	if providedHelmFlags.env != "" {
		env = providedHelmFlags.env
	}
	if providedHelmFlags.platform != "" {
		platform = providedHelmFlags.platform
	}

	registriesConfigFile := filepath.Join(workDir, "configs/helm/registries.yaml")
	registries, err := NewRegistries(registriesConfigFile)
	if err != nil {
		return err
	}

	if providedHelmFlags.checkChart {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		chartFile, err := os.ReadFile(filepath.Join(cwd, "Chart.yaml"))
		if err != nil {
			return err
		}
		var chart *Chart
		err = yaml.Unmarshal(chartFile, &chart)
		if err != nil {
			return err
		}

		if chart.Dependencies == nil {
			return nil
		}

		for _, dep := range *chart.Dependencies {
			u, err := url.Parse(dep.Repository)
			if err != nil {
				return err
			}
			chartURL, err := url.JoinPath(dep.Repository, dep.Name)
			if err != nil {
				return err
			}

			if u.Scheme == "oci" && IsInternalOciRepository(dep.Repository, registries) {
				v, err := semver.NewVersion(dep.Version)
				if err != nil {
					log.Errorf("Version must follow semver for internal Helm chart %s in chart dependencies\n", chartURL)
					numberOfErrors++
				} else if v.Prerelease() != "" {
					log.Errorf("Version must not be a prerelease for internal Helm chart %s in chart dependencies\n", chartURL)
					numberOfErrors++
				}
				continue
			}

			if !IsExternalChartAllowed(chartURL, whitelist) {
				log.Errorf("Forbidden external Helm chart %s found in chart dependencies\n", chartURL)
				numberOfErrors++
			}
		}
	} else {
		helmExportConfigFiles, err := filepath.Glob(filepath.Join(
			helmConfigPath,
			platform,
			env,
			"helm-export.yaml"))
		if err != nil {
			return err
		}
		for _, helmExportConfigFile := range helmExportConfigFiles {
			log.Infof("Reading config file %s", helmExportConfigFile)
			applications, err := NewExportConfig(helmExportConfigFile)
			if err != nil {
				return err
			}
			for _, application := range applications.Applications {
				for _, source := range application.Sources {
					if source.IsExternalChart(registries) {
						if !IsExternalChartAllowed(source.RepoURL, whitelist) {
							log.Errorf("Forbidden external Helm chart %s found in %s\n", source.RepoURL, helmExportConfigFile)
							numberOfErrors++
						}
					} else if source.IsHelmOCI() {
						v, err := semver.NewVersion(source.TargetRevision)
						if err != nil {
							log.Errorf("Version must follow semver for internal Helm chart %s found in %s\n", source.RepoURL, helmExportConfigFile)
							numberOfErrors++
						} else if v.Prerelease() != "" {
							log.Errorf("Version must not be a prerelease for internal Helm chart %s found in %s\n", source.RepoURL, helmExportConfigFile)
							numberOfErrors++
						}
					}
				}
			}
		}
	}

	if numberOfErrors > 0 {
		return errHelmChartsProblemFound
	}
	return nil
}

func IsInternalOciRepository(repoURL string, registries *HelmRegistries) bool {
	if u, err := url.Parse(repoURL); err == nil {
		if AwsEcrRegexp.Match([]byte(u.Host)) {
			matches := AwsEcrRegexp.FindStringSubmatch(u.Host)
			account := matches[AwsEcrRegexp.SubexpIndex("account")]
			region := matches[AwsEcrRegexp.SubexpIndex("region")]

			for _, registry := range registries.ECR {
				if account == registry.Account && region == registry.Region {
					return true
				}
			}
		}
	}
	return false
}

func IsExternalChartAllowed(repoURL string, whitelist []string) bool {
	for _, u := range whitelist {
		if repoURL == u {
			return true
		}
	}

	return false
}

func init() {
	helmCheckCmd.Flags().BoolVarP(&providedHelmFlags.checkChart, "chart", "c", false, "check a helm chart (instead of argocd applications) for external helm charts usage")
	helmCheckCmd.Flags().StringVarP(&providedHelmFlags.cfgFile, "file", "f", "", "yaml file containing whitelisted external helm charts")
	err := helmCheckCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}
}
