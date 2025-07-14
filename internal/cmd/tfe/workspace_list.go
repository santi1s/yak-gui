package tfe

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/constant"
	"github.com/santi1s/yak/internal/helper"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type listFlags struct {
	email        string
	not          bool
	organization string
	tag          string
	json         bool
	yaml         bool
}

var errResourceTypeDoesNotExist = errors.New("resource type does not exist")
var errCantReadAwsProfile = errors.New("can't read awsProfile from configuration")
var errCantReadAwsRegion = errors.New("can't read awsRegion from configuration")

func tfeListWorkspace(cmd *cobra.Command, args []string) error {
	config, err := getTfeConfig()
	if err != nil {
		return err
	}
	client, err := getTfeClient(config)
	if err != nil {
		return err
	}

	ctx := context.Background()

	result := []string{}
	page := 1
	for {
		options := &tfe.WorkspaceListOptions{
			ListOptions: tfe.ListOptions{
				PageSize:   20,
				PageNumber: page,
			},
		}

		wildcard := false
		if strings.HasSuffix(providedListFlags.tag, "*") {
			wildcard = true
		}

		if !wildcard {
			if providedListFlags.not {
				options.ExcludeTags = providedListFlags.tag
			} else {
				options.Tags = providedListFlags.tag
			}
		}

		items := []*tfe.Workspace{}
		wsList, err := client.Workspaces.List(ctx, providedListFlags.organization, options)
		if err != nil {
			return fmt.Errorf("error while listing workspaces for organization %s: %v", providedListFlags.organization, err)
		}
		items = append(items, wsList.Items...)

		for _, w := range items {
			if wildcard {
				found := false
				for _, t := range w.TagNames {
					if strings.HasPrefix(t, strings.Replace(providedListFlags.tag, "*", "", 1)) {
						found = true
						break
					}
				}

				if !providedListFlags.not && found {
					result = append(result, w.Name)
				} else if providedListFlags.not && !found {
					result = append(result, w.Name)
				}
			} else {
				result = append(result, w.Name)
			}
		}

		page++
		if wsList.CurrentPage == wsList.TotalPages || wsList.TotalPages == 0 {
			break
		}
	}

	if len(result) > 0 && providedListFlags.email != "" {
		profile := viper.GetString("email.awsProfile")
		if profile == "" {
			return errCantReadAwsProfile
		}

		region := viper.GetString("email.awsRegion")
		if region == "" {
			return errCantReadAwsRegion
		}

		err := helper.SendEmail(providedListFlags.email, fmt.Sprintf("[TFE] Detected workspaces without owner tag in organization %s", providedListFlags.organization), cli.SprintYAML(result))

		if err != nil {
			return err
		}
	}

	switch {
	case providedListFlags.json:
		return cli.PrintJSON(result)
	case providedListFlags.yaml:
		return cli.PrintYAML(result)
	default:
		return cli.PrintYAML(result)
	}
}

var (
	providedListFlags   listFlags
	tfeWorkspaceListCmd = &cobra.Command{
		Use:   "list",
		Short: "list workspaces",
		RunE:  tfeListWorkspace,
		Example: "List all workspaces which have tag owner:sre-iac:\n" +
			"list --tag owner:sre-iac\n\n" +
			"List all workspaces which don't have tag owner:sre-iac:\n" +
			"list --not --tag owner:sre-iac\n\n" +
			"Wildcard at the end of the tag is also accepted! (please not that it can takes a few seconds to get the answer)\n" +
			"List all workspaces which have a tag owner:*:\n" +
			"list --tag owner:*\n" +
			"list --not --tag owner:*\n\n" +
			"Using --email will also send a copy of the output to the provided email address:\n" +
			"list --tag owner:sre-iac --email example@doctolib.com",
	}
)

func init() {
	tfeWorkspaceListCmd.PersistentFlags().StringVar(&providedListFlags.email, "email", "", "send a copy of the output to the specified email address")
	tfeWorkspaceListCmd.PersistentFlags().BoolVar(&providedListFlags.not, "not", false, "reverse --tag condition")
	tfeWorkspaceListCmd.PersistentFlags().StringVarP(&providedListFlags.tag, "tag", "t", "", "filter workspaces having this tag (wildcard support, eg. owner:*)")
	tfeWorkspaceListCmd.PersistentFlags().StringVar(&providedListFlags.organization, "organization", constant.TfeDefaultOrganization, "execute command on given organization")
	tfeWorkspaceListCmd.PersistentFlags().BoolVar(&providedListFlags.json, "json", false, "format output in JSON")
	tfeWorkspaceListCmd.PersistentFlags().BoolVar(&providedListFlags.yaml, "yaml", false, "format output in YAML")
	tfeWorkspaceListCmd.MarkFlagsMutuallyExclusive("json", "yaml")
}
