package helm

import (
	"github.com/doctolib/yak/internal/constant"
	"github.com/spf13/cobra"
)

type HelmFlags struct {
	application string
	env         string
	platform    string
	cfgFile     string
	checkChart  bool
}

type Chart struct {
	AppVersion   string               `yaml:"appVersion"`
	Dependencies *[]ChartDependencies `yaml:"dependencies"`
	Name         string               `yaml:"name"`
	Version      string               `yaml:"version"`
}

type ChartDependencies struct {
	Name       string `yaml:"name"`
	Repository string `yaml:"repository"`
	Version    string `yaml:"version"`
}

var (
	providedHelmFlags HelmFlags
	helmCmd           = &cobra.Command{
		Use:   "helm",
		Short: "Helm related commands",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Root().Name() == constant.CliName && cmd.Root().PersistentPreRun != nil {
				cmd.Root().PersistentPreRun(cmd, args)
			}

			if cmd.HasParent() && cmd.Parent().Name() == "completion" {
				switch cmd.Name() {
				case "bash", "zsh", "fish", "powershell":
					cmd.ResetFlags()
				}
			}
			cmd.SilenceUsage = true
		},
	}
)

func GetRootCmd() *cobra.Command {
	return helmCmd
}

func init() {
	helmCmd.AddCommand(helmExportCmd)
	helmCmd.AddCommand(helmCheckCmd)
}
