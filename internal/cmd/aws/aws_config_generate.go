package aws

import (
	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/teleport"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	providedGenerateCredentialsFlags generateCredentialsFlags
)

type generateCredentialsFlags struct {
	role string
	app  string
}

var (
	generateConfigCmd = &cobra.Command{
		Use:   "generate",
		Short: "generate AWS credentials process output",
		Run:   generateAWSCredentialProcessWithTeleport,
	}
)

func generateAWSCredentialProcessWithTeleport(cmd *cobra.Command, args []string) {
	if providedGenerateCredentialsFlags.role == "" {
		log.Error("--role is required")
		return
	}
	if providedGenerateCredentialsFlags.app == "" {
		log.Error("--app is required")
		return
	}

	tshStatus, err := teleport.GetTshStatus()
	if err != nil {
		log.Errorf("Error getting Teleport status: %s", err)
		return
	}

	credentials, err := teleport.GenerateAWSCredentialsProcess(providedGenerateCredentialsFlags.app, providedGenerateCredentialsFlags.role, tshStatus.Name, tshStatus.ValidUntil)
	if err != nil {
		log.Errorf("Error generating AWS credentials: %s", err)
		return
	}
	cli.Println(credentials)
}

func init() {
	generateConfigCmd.Flags().StringVarP(&providedGenerateCredentialsFlags.role, "role", "r", "", "target role to assume")
	generateConfigCmd.Flags().StringVarP(&providedGenerateCredentialsFlags.app, "app", "a", "", "target app")
}
