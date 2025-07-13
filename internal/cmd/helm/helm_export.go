package helm

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	helmExportCmd = &cobra.Command{
		Use:     "export",
		Short:   "Export Helm charts to YML manifests in kube repository",
		RunE:    exportHelmCharts,
		Example: "yak helm export",
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
)

func exportHelmCharts(cmd *cobra.Command, args []string) error {
	workDir := os.Getenv("YAK_HELM_WORKDIR")
	exportDir := os.Getenv("YAK_HELM_EXPORTDIR")
	helmConfigPath := filepath.Join(workDir, "configs/helm")
	errors := 0
	// Registries authentication
	registriesConfigFile := filepath.Join(workDir, "configs/helm/registries.yaml")
	registries, err := NewRegistries(registriesConfigFile)
	if err != nil {
		return err
	}
	if err := registries.Authenticate(); err != nil {
		return err
	}

	platform := "*"
	env := "*"
	if providedHelmFlags.env != "" {
		env = providedHelmFlags.env
	}
	if providedHelmFlags.platform != "" {
		platform = providedHelmFlags.platform
	}

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
		exportErrors, err := applications.Export(workDir, exportDir)
		if err != nil {
			return err
		}
		errors += exportErrors
	}

	if errors > 0 {
		return fmt.Errorf("%d errors when generating manifests", errors)
	}
	return nil
}

func init() {
	helmExportCmd.Flags().StringVarP(&providedHelmFlags.env, "env", "e", "", "Select an environment to export helm charts")
	helmExportCmd.Flags().StringVarP(&providedHelmFlags.platform, "platform", "P", "", "Select a platform to export helm charts")
	helmExportCmd.Flags().StringVarP(&providedHelmFlags.application, "app", "a", "", "Select an application to export helm charts")
}
