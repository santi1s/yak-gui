package version

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/constant"
	"github.com/google/go-github/v73/github"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

type versionFlags struct {
	short   bool
	noCheck bool
}

var (
	providedFlags = versionFlags{}
	versionCmd    = &cobra.Command{
		Use:              "version",
		Short:            "show version",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {},
		RunE:             version,
	}
)

var (
	errCantGetGithubLatestRelease = fmt.Errorf("unable to get %s latest release version. Please check you have a valid GITHUB_TOKEN environment variable set", constant.CliName)
)

func GetRootCmd() *cobra.Command {
	return versionCmd
}

func checkLastVersion() (string, error) {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		githubToken = os.Getenv("HOMEBREW_GITHUB_API_TOKEN")
	}

	if githubToken != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})
		tc := oauth2.NewClient(context.Background(), ts)
		gc := github.NewClient(tc)
		release, _, err := gc.Repositories.GetLatestRelease(context.Background(), constant.GithubOrganization, constant.GithubRepository)

		if err != nil {
			return "", errCantGetGithubLatestRelease
		}

		return release.GetTagName(), nil
	}

	return "", errCantGetGithubLatestRelease
}

func PrintUpgradeMessage() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.WithError(err).Debugf("can't get home directory")
		return
	}

	var date time.Time
	versionCachePathStat, err := os.Stat(home + constant.VersionCacheFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.WithError(err).Debugf("error when doing stat on file %s", home+constant.VersionCacheFilePath)
			return
		}

		date = time.Now().AddDate(-1, 1, 1)
	} else {
		date = versionCachePathStat.ModTime()
	}

	now := time.Now()
	if date.Year() != now.Year() || date.YearDay() != now.YearDay() {
		err := os.MkdirAll(filepath.Dir(home+constant.VersionCacheFilePath), 0755)
		if err != nil {
			log.Debugf("can't create directory %s", filepath.Dir(home+constant.VersionCacheFilePath))
			return
		}
		_, err = os.Create(home + constant.VersionCacheFilePath)
		if err != nil {
			log.Debugf("can't touch file %s", home+constant.VersionCacheFilePath)
			return
		}

		latest, err := checkLastVersion()
		if err != nil {
			return
		}
		printUpgradeMessage(latest)
	}
}

func printUpgradeMessage(latestVersion string) {
	if constant.Version != latestVersion {
		_, _ = cli.PrintfErr("%s %s is available, please run 'brew upgrade %s'.\n", constant.CliName, latestVersion, constant.CliName)
	}
}

func version(cmd *cobra.Command, args []string) error {
	if providedFlags.short {
		cli.Printf("%s\n", constant.Version)
	} else {
		cli.Printf("%s %s\n", constant.CliName, constant.Version)
	}

	if !providedFlags.noCheck {
		latest, err := checkLastVersion()
		if err != nil {
			return err
		}

		printUpgradeMessage(latest)
	}

	return nil
}

func init() {
	versionCmd.Flags().BoolVarP(&providedFlags.short, "short", "s", false, "only print version")
	versionCmd.Flags().BoolVar(&providedFlags.noCheck, "no-check", false, "don't check for new version")
}
