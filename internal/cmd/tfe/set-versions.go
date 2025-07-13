package tfe

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/doctolib/yak/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hashicorp/go-tfe"
)

type versionFlags struct {
	setVersionFile string
}

type actionResult struct {
	Version string `json:"version"`
	Action  string `json:"action"`
	Error   string `json:"error,omitempty"`
}

var (
	deprecationMessage = "Automatic deprecation"

	providedVersionFlags versionFlags
	tfeVersionCmd        = &cobra.Command{
		Use:   "set-versions",
		Short: "manage terraform versions in TFE",
		RunE:  tfeVersion,
	}

	errConfigFileDoesNotExist = errors.New("provided file does not exist")
)

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func tfeVersion(cmd *cobra.Command, args []string) error {
	if _, err := os.Stat(providedVersionFlags.setVersionFile); err != nil {
		return errConfigFileDoesNotExist
	}

	terraformVersions := viper.New()
	terraformVersions.SetConfigFile(providedVersionFlags.setVersionFile)
	cobra.CheckErr(terraformVersions.ReadInConfig())

	config, err := getTfeConfig()
	if err != nil {
		return err
	}

	client, err := getTfeClient(config)
	if err != nil {
		return err
	}

	deprecatedVersions := []string{terraformVersions.GetString("deprecated")}
	acceptedVersions := []string{terraformVersions.GetString("stable"), terraformVersions.GetString("next")}

	gotError := false

	ctx := context.Background()

	var resultArray []actionResult

	page := 1
	for {
		tfList, err := client.Admin.TerraformVersions.List(ctx, &tfe.AdminTerraformVersionsListOptions{
			ListOptions: tfe.ListOptions{
				PageNumber: page,
			},
		})

		if err != nil {
			return fmt.Errorf("wasn't able to list terraform versions: %s", err)
		}

		for _, v := range tfList.Items {
			ar := actionResult{}
			ar.Version = v.Version
			if contains(deprecatedVersions, v.Version) {
				if !v.Deprecated || !v.Enabled {
					ar.Action = "deprecate"
					_, err := client.Admin.TerraformVersions.Update(ctx, v.ID, tfe.AdminTerraformVersionUpdateOptions{
						Type:             "terraform-versions",
						Deprecated:       tfe.Bool(true),
						DeprecatedReason: &deprecationMessage,
						Enabled:          tfe.Bool(true),
					})
					if err != nil {
						ar.Error = err.Error()
						gotError = true
					}
					resultArray = append(resultArray, ar)
				}
			} else if contains(acceptedVersions, v.Version) {
				if !v.Enabled || v.Deprecated || v.DeprecatedReason != nil {
					ar.Action = "enable"
					_, err := client.Admin.TerraformVersions.Update(ctx, v.ID, tfe.AdminTerraformVersionUpdateOptions{
						Type:             "terraform-versions",
						Deprecated:       tfe.Bool(false),
						DeprecatedReason: nil,
						Enabled:          tfe.Bool(true),
					})
					if err != nil {
						ar.Error = err.Error()
						gotError = true
					}
					resultArray = append(resultArray, ar)
				}
			} else if v.Enabled {
				ar.Action = "disable"
				_, err := client.Admin.TerraformVersions.Update(ctx, v.ID, tfe.AdminTerraformVersionUpdateOptions{
					Type:             "terraform-versions",
					DeprecatedReason: v.DeprecatedReason, // keep it, otherwise TFE will set to its default message :(
					Enabled:          tfe.Bool(false),
				})
				if err != nil {
					ar.Error = err.Error()
					gotError = true
				}
				resultArray = append(resultArray, ar)
			}
		}

		page++
		if tfList.CurrentPage == tfList.TotalPages || tfList.TotalPages == 0 {
			break
		}
	}
	err = cli.PrintYAML(resultArray)
	if err != nil {
		return err
	}

	if gotError {
		return fmt.Errorf("got errors during the execution, check the logs")
	}
	return nil
}

func init() {
	tfeVersionCmd.Flags().StringVarP(&providedVersionFlags.setVersionFile, "file", "f", "", "yaml file containing deprecated, stable and next versions")
	err := tfeVersionCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}
}
