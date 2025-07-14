package aws

import (
	"strings"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/santi1s/yak/internal/teleport"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	providedRequestRoleFlags requestRoleFlags
)

const (
	defaultOutputPath = "~/.aws/config.teleport"
)

type requestRoleFlags struct {
	targets    string
	permission string
	reviewers  string
	outputPath string
	bypassTsh  bool
}

var (
	requestConfigCmd = &cobra.Command{
		Use:   "request",
		Short: "request AWS config file with Teleport",
		Run:   requestAWSConfigWithTeleport,
	}
	rolesToOverride []teleport.Role
)

func requestAWSConfigWithTeleport(cmd *cobra.Command, args []string) {
	tConfig, err := teleport.ReadTeleportConfig()
	if err != nil {
		log.Errorf("Error reading Teleport config: %s", err)
		return
	}
	if providedRequestRoleFlags.targets == "" {
		cli.Println("Here's a list of available targets to request:")
		for _, account := range tConfig.Accounts {
			for _, role := range account.Roles {
				if role.Type == teleport.AWSConfigType {
					cli.Println("- name       :", account.Name)
					cli.Println("  permission :", role.Permission)
				}
			}
		}
		return
	}
	requestedAccounts := strings.Split(providedRequestRoleFlags.targets, ",")

	if providedRequestRoleFlags.bypassTsh {
		if !cli.AskConfirmation("Bypassing tsh can trigger a crisis on production. Continue ?") {
			return
		}
		rolesToOverride = teleport.BackupRolesToRequest(tConfig, providedRequestRoleFlags.permission, requestedAccounts, teleport.AWSConfigType)
	} else {
		if err := teleport.TshLogin(tConfig); err != nil {
			log.Errorf("Error on tsh login: %s", err)
			return
		}

		tshStatus, err := teleport.GetTshStatus()
		if err != nil {
			log.Errorf("Error getting Teleport active roles: %s", err)
			return
		}

		lo, err := teleport.CheckIfNeedToLogout(tshStatus.ValidUntil)
		if err != nil {
			log.Errorf("Error checking if need to logout: %s", err)
			return
		}
		if lo && cli.AskConfirmation("Do you want to logout to refresh credentials TTL ?") {
			err = teleport.TshLogout()
			if err != nil {
				log.Errorf("Error on logout: %s", err)
			}
			return
		}

		rolesToRequest, rolesAlreadyAssumed := teleport.TeleportRolesToRequest(tConfig, providedRequestRoleFlags.permission, tshStatus.Roles, requestedAccounts, teleport.AWSConfigType)
		if len(rolesToRequest) > 0 {
			err = teleport.TshRequestNeededRoles(rolesToRequest, tshStatus.Name, providedRequestRoleFlags.targets, tshStatus.ValidUntil, teleport.ReviewersToSlice(providedRequestRoleFlags.reviewers), teleport.AWSConfigType)
			if err != nil {
				log.Errorf("Error requesting needed roles: %s", err)
				return
			}
		}

		err = teleport.TshAwsLogin(append(rolesToRequest, rolesAlreadyAssumed...))
		if err != nil {
			log.Errorf("Error on login to a Teleport AWS application: %s", err)
			return
		}
		rolesToOverride = append(rolesToRequest, rolesAlreadyAssumed...)
	}

	config, err := helper.LoadAWSConfig()
	if err != nil {
		log.Errorf("Error loading AWS config file: %s", err)
		return
	}

	err = teleport.OverrideAWSConfigProfiles(config, rolesToOverride, []string{"region"}, providedRequestRoleFlags.bypassTsh)
	if err != nil {
		log.Errorf("Error overriding AWS config profiles: %s", err)
		return
	}

	err = helper.SaveIniFile(config, providedRequestRoleFlags.outputPath)
	if err != nil {
		log.Errorf("Error saving AWS config file: %s", err)
		return
	}
	cli.Println("You can now use the AWS CLI with Teleport profiles by exporting this environment variable:")
	cli.Println("")
	cli.Println("export AWS_CONFIG_FILE=" + providedRequestRoleFlags.outputPath)
}

func init() {
	output, _ := homedir.Expand(defaultOutputPath)
	requestConfigCmd.Flags().StringVarP(&providedRequestRoleFlags.targets, "targets", "t", "", "target accounts (comma separated)")
	requestConfigCmd.Flags().StringVarP(&providedRequestRoleFlags.permission, "permission", "p", "superadministrator", "permissions type to request")
	requestConfigCmd.Flags().StringVarP(&providedRequestRoleFlags.reviewers, "reviewers", "r", "", "override default reviewers of the request (comma separated)")
	requestConfigCmd.Flags().StringVarP(&providedRequestRoleFlags.outputPath, "output", "o", output, "config output path")
	requestConfigCmd.Flags().BoolVar(&providedRequestRoleFlags.bypassTsh, "bypass-tsh", false, "disable connection with tsh (it can trigger crisis on production)")
}
