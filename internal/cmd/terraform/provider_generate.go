package terraform

import (
	"fmt"
	"os"

	"github.com/doctolib/yak/cli"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/spf13/cobra"
)

var (
	providerGenerateCmd = &cobra.Command{
		Use:   "generate",
		Short: "generate empty provider config for CI",
		RunE:  providerGenerate,
	}
)

func providerGenerate(cmd *cobra.Command, args []string) error {
	module, diags := tfconfig.LoadModule(".")
	if diags.HasErrors() {
		return fmt.Errorf("error reading module from current directory: %s", diags)
	}
	for providerName, providerReq := range module.RequiredProviders {
		content := ""
		filename := fmt.Sprintf("provider_%s.tf", providerName)
		_, err := os.Stat(filename)
		if os.IsNotExist(err) {
			// File does not exist, we create it
			if len(providerReq.ConfigurationAliases) == 0 {
				content = fmt.Sprintf("provider \"%s\" {}\n", providerName)
			} else {
				// For each alias we create an empty definition block
				for _, providerRef := range providerReq.ConfigurationAliases {
					content += fmt.Sprintf("provider \"%s\" {\n  alias = \"%s\"\n}\n", providerRef.Name, providerRef.Alias)
				}
			}

			err = os.WriteFile(filename, []byte(content), 0600)
			if err != nil {
				_, _ = cli.PrintlnErr(err)
				continue
			}
		} else if err != nil {
			// Other errors
			return err
		}
	}
	return nil
}
