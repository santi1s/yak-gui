package kafka

import (
	"github.com/santi1s/yak/internal/constant"

	"github.com/spf13/cobra"
)

type KafkaFlags struct {
	cfgFile       string
	roleBasedAuth bool
}

var ProvidedKafkaFlags KafkaFlags

var (
	kafkaCmd = &cobra.Command{
		Use:   "kafka",
		Short: "tools to manage kafka",
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
	return kafkaCmd
}

func init() {
	kafkaCmd.PersistentFlags().StringVarP(&ProvidedKafkaFlags.cfgFile, "configurationFile", "f", "~/.kaf/config", "path to configuration file in system")
	kafkaCmd.AddCommand(kafkaReplicationCmd)
	kafkaCmd.AddCommand(kafkaWorkerResetCmd)
	kafkaCmd.AddCommand(kafkaMonitorConsumerLagCmd)
}
