package tfe

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type checkVersionFlags struct {
	teamsFile        string
	setVersionFile   string
	organization     string
	allOrganizations bool
	sendEmail        bool
	json             bool
	yaml             bool
}

type WorkspaceDetails struct {
	WorkspaceName                   string
	OwnerEmail                      string
	TerraformVersion                string
	UsingDeprecatedTerraformVersion bool
}

type TerraformVersions struct {
	Deprecated string `yaml:"deprecated"`
	Stable     string `yaml:"stable"`
	Next       string `yaml:"next"`
}

type Team struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

type Workspace struct {
	OwnerEmail           string   `json:"owner_email"`
	DeprecatedWorkspaces []string `json:"deprecated_workspaces"`
}

var (
	tfeCheckVersionCmd = &cobra.Command{
		Use:   "check-versions",
		Short: "check workspaces using a deprecated terraform version",
		RunE:  tfeCheckVersion,
	}

	providedCheckVersionFlags                    checkVersionFlags
	errTeamsFileDoesNotExist                     = errors.New("provided teams mapping file does not exist")
	errOrganizationCantBeSetWithAllOrganizations = errors.New("organization and all-organizations parameters are mutually exclusive")
)

func cleanOwnerTag(ownerTag string) string {
	return strings.ReplaceAll(ownerTag, "owner:", "")
}

func getOwnerEmail(teamName string, teams []Team) string {
	for _, attributes := range teams {
		if teamName == attributes.Name {
			return attributes.Email
		}
	}

	// fallback on SRE Green
	return "team-peer-sre-green+fixowner@doctolib.com"
}

func isTerraformVersionDeprecated(currentTerraformVersion string, supportedVersions *TerraformVersions) bool {
	if currentTerraformVersion != supportedVersions.Deprecated &&
		currentTerraformVersion != supportedVersions.Stable &&
		currentTerraformVersion != supportedVersions.Next {
		return true
	}

	return currentTerraformVersion == supportedVersions.Deprecated
}

func getWorkspacesUsingDeprecatedTerraformVersion(workspacesMap map[string][]WorkspaceDetails, supportedVersions *TerraformVersions) error {
	var dw []Workspace
	for team, workspaces := range workspacesMap {
		var w Workspace
		w.OwnerEmail = team
		if len(workspaces) == 1 && !workspaces[0].UsingDeprecatedTerraformVersion {
			w.DeprecatedWorkspaces = append(w.DeprecatedWorkspaces, "")
			dw = append(dw, w)
			continue
		}
		for _, workspace := range workspaces {
			if workspace.UsingDeprecatedTerraformVersion {
				w.DeprecatedWorkspaces = append(w.DeprecatedWorkspaces, workspace.WorkspaceName)
			}
		}
		dw = append(dw, w)
	}

	if len(dw) > 0 && providedCheckVersionFlags.sendEmail {
		profile := viper.GetString("email.awsProfile")
		if profile == "" {
			return errCantReadAwsProfile
		}

		region := viper.GetString("email.awsRegion")
		if region == "" {
			return errCantReadAwsRegion
		}

		for i, wsl := range dw {
			if len(wsl.DeprecatedWorkspaces) > 0 {
				body := fmt.Sprintf("Hello, your team have %d terraform workspaces using deprecated version of terraform (%s).\nYou need to upgrade to stable (%s) or next (%s) version before the end of the quarter.\n\nImpacted workspaces:\n", len(dw[i].DeprecatedWorkspaces), supportedVersions.Deprecated, supportedVersions.Stable, supportedVersions.Next)
				for _, ws := range wsl.DeprecatedWorkspaces {
					body += ws + "\n"
				}

				err := helper.SendEmail(wsl.OwnerEmail, "[FA][TFE] Workspaces using deprecated terraform version to update", body)
				if err != nil {
					_, _ = cli.PrintfErr("Failed to send email to %s: %s", wsl.OwnerEmail, err)
				}
			}
		}

		err := helper.SendEmail("team-peer-sre-green@doctolib.com", "[TFE] Workspaces using deprecated terraform version to update", cli.SprintYAML(dw))
		if err != nil {
			return err
		}
	}

	switch {
	case providedCheckVersionFlags.json:
		return cli.PrintJSON(dw)
	case providedCheckVersionFlags.yaml:
		return cli.PrintYAML(dw)
	default:
		return cli.PrintYAML(dw)
	}
}

func getWorkspaces(ctx context.Context, client *tfe.Client, organizationName string, options tfe.WorkspaceListOptions) (*tfe.WorkspaceList, error) {
	w, err := client.Workspaces.List(ctx, organizationName, &options)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func getTags(ctx context.Context, client *tfe.Client, workspaceID string, options tfe.WorkspaceTagListOptions) *tfe.TagList {
	t, err := client.Workspaces.ListTags(ctx, workspaceID, &options)
	if err != nil {
		log.Fatal(err)
	}

	return t
}

func tfeCheckVersion(cmd *cobra.Command, args []string) error {
	if providedCheckVersionFlags.organization != "" && providedCheckVersionFlags.allOrganizations {
		return errOrganizationCantBeSetWithAllOrganizations
	}

	var currentPage int
	var ownerEmail string
	var usingDeprecatedTerraformVersion bool
	var workspaceListOptions tfe.WorkspaceListOptions
	var tagListOptions tfe.WorkspaceTagListOptions
	var currentWorkspace WorkspaceDetails
	teams := []Team{}
	workspacesMap := make(map[string][]WorkspaceDetails)

	workspaceListOptions.PageSize = 50
	workspaceListOptions.Include = []tfe.WSIncludeOpt{
		"current_run",
	}
	query := "owner"

	if _, err := os.Stat(providedCheckVersionFlags.teamsFile); err != nil {
		return errTeamsFileDoesNotExist
	}
	viper.SetConfigFile(providedCheckVersionFlags.teamsFile)

	if _, err := os.Stat(providedCheckVersionFlags.setVersionFile); err != nil {
		return errConfigFileDoesNotExist
	}
	viper.SetConfigFile(providedCheckVersionFlags.setVersionFile)
	cobra.CheckErr(viper.MergeInConfig())

	supportedTerraformVersions := &TerraformVersions{}
	err := viper.Unmarshal(supportedTerraformVersions)
	if err != nil {
		log.Fatal(err)
	}

	yfile, err := os.ReadFile(providedCheckVersionFlags.teamsFile)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(yfile, &teams)
	if err != nil {
		log.Fatal(err)
	}

	config, err := getTfeConfig()
	if err != nil {
		return err
	}

	client, err := getTfeClient(config)
	if err != nil {
		return err
	}

	ctx := context.Background()

	orgs := []string{}
	if providedCheckVersionFlags.allOrganizations {
		organizationsList, err := client.Admin.Organizations.List(ctx, &tfe.AdminOrganizationListOptions{})
		if err != nil {
			return fmt.Errorf("impossible to list TFE organizations: %v", err)
		}

		for _, o := range organizationsList.Items {
			orgs = append(orgs, o.Name)
		}
	} else {
		orgs = append(orgs, providedCheckVersionFlags.organization)
	}

	for _, org := range orgs {
		currentPage = 1
		for {
			workspaceListOptions.PageNumber = currentPage
			workspaces, err := getWorkspaces(ctx, client, org, workspaceListOptions)
			if err != nil {
				return err
			}
			for _, workspace := range workspaces.Items {
				w := *workspace
				currentWorkspace = WorkspaceDetails{}
				currentWorkspace.WorkspaceName = fmt.Sprintf("%s / %s", w.Organization.Name, w.Name)
				currentWorkspace.TerraformVersion = w.TerraformVersion
				usingDeprecatedTerraformVersion = isTerraformVersionDeprecated(w.TerraformVersion, supportedTerraformVersions)
				currentWorkspace.UsingDeprecatedTerraformVersion = usingDeprecatedTerraformVersion

				tagListOptions.Query = &query
				tags := getTags(ctx, client, w.ID, tagListOptions)
				found := false
				for _, tag := range tags.Items {
					if strings.Contains(tag.Name, "owner") {
						ownerEmail = getOwnerEmail(cleanOwnerTag(tag.Name), teams)
						found = true
					}
				}

				if !found {
					ownerEmail = "team-peer-sre-green+noownertag@doctolib.com"
				}
				currentWorkspace.OwnerEmail = ownerEmail

				if usingDeprecatedTerraformVersion {
					workspacesMap[ownerEmail] = append(workspacesMap[ownerEmail], currentWorkspace)
				}
			}
			currentPage++
			if workspaces.CurrentPage == workspaces.TotalPages || workspaces.TotalPages == 0 {
				break
			}
		}
	}
	return getWorkspacesUsingDeprecatedTerraformVersion(workspacesMap, supportedTerraformVersions)
}

func init() {
	tfeCheckVersionCmd.Flags().StringVarP(&providedCheckVersionFlags.setVersionFile, "file", "f", "", "yaml file containing deprecated, stable and next versions")
	tfeCheckVersionCmd.Flags().StringVarP(&providedCheckVersionFlags.teamsFile, "teams", "t", "", "yaml file containing mapping between owners and team emails")
	tfeCheckVersionCmd.Flags().StringVarP(&providedCheckVersionFlags.organization, "organization", "o", "", "organization name to check")
	tfeCheckVersionCmd.Flags().BoolVar(&providedCheckVersionFlags.allOrganizations, "all-organizations", false, "check all organizations (require an admin TFE token)")
	tfeCheckVersionCmd.Flags().BoolVar(&providedCheckVersionFlags.json, "json", false, "format output in JSON")
	tfeCheckVersionCmd.Flags().BoolVar(&providedCheckVersionFlags.yaml, "yaml", false, "format output in YAML")
	tfeCheckVersionCmd.Flags().BoolVar(&providedCheckVersionFlags.sendEmail, "send-email", false, "send email to each team and a global one to sre-green")
	tfeCheckVersionCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	err := tfeCheckVersionCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}
	err = tfeCheckVersionCmd.MarkFlagRequired("teams")
	if err != nil {
		panic(err)
	}
}
