package terraform

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	"github.com/google/go-github/v73/github"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

var (
	moduleBumpCmd = &cobra.Command{
		Use:   "bump",
		Short: "bump a module version",
		RunE:  moduleBump,
	}

	// Parameter errors
	errMissingModuleName           = errors.New("a module name is needed <module_name>/<module_main_provider>")
	errMissingTargetVersion        = errors.New("a target version is needed X.X.X")
	errVersionFormat               = errors.New("version is not compliant with SemVer format X.X.X")
	errMissingRelease              = errors.New("missing release type")
	errAskConfirmationNotConfirmed = errors.New("action not confirmed by user")
)

func init() {
	moduleBumpCmd.Flags().StringVarP(&providedFlags.name, "name", "n", "", "name of the module")
	moduleBumpCmd.Flags().StringVarP(&providedFlags.version, "version", "v", "", "target version for the bump")
	moduleBumpCmd.Flags().BoolVar(&providedFlags.check, "check", false, "run a version check after bump")
	moduleBumpCmd.Flags().StringVarP(&providedFlags.release, "release", "R", "", "release type (patch, minor or major)")
	moduleBumpCmd.Flags().StringVarP(&providedFlags.description, "description", "d", "", "pull-request description")
	moduleBumpCmd.Flags().StringSliceVarP(&providedFlags.repository, "repository", "r", nil, "bump module on specific module repositories list")
	moduleBumpCmd.Flags().BoolVar(&providedFlags.allRepositories, "all-repos", false, "bump module in all terraform-* repositories")
	moduleBumpCmd.Flags().BoolVar(&providedFlags.skipConfirm, "skip-confirm", false, "will skip confirmation for bump in specific module repositories")
	moduleBumpCmd.Flags().IntVarP(&providedFlags.parallelism, "parallelism", "p", 8, "number of parallel tasks when --repository or --all-repos is used")
	moduleBumpCmd.Flags().StringVarP(&providedFlags.branch, "branch", "b", "", "specify a custom branch name for the pull-request creation when using --repository or --all-repos")

	moduleBumpCmd.MarkFlagsMutuallyExclusive("repository", "all-repos")

	err := moduleBumpCmd.MarkFlagRequired("name")
	if err != nil {
		panic(err)
	}
	err = moduleBumpCmd.MarkFlagRequired("version")
	if err != nil {
		panic(err)
	}
}

func getLatestModuleVersion(module string) (string, error) {
	gh, err := helper.GetGithubClient()
	if err != nil {
		return "", err
	}

	repositoryName := fmt.Sprintf("terraform-%s-%s", strings.Split(module, "/")[1], strings.Split(module, "/")[0])

	opt := &github.ListOptions{PerPage: 10}

	// get all pages of results
	var allTags []*github.RepositoryTag
	for {
		tags, resp, err := gh.Repositories.ListTags(context.Background(), "doctolib", repositoryName, opt)
		if err != nil {
			return "", err
		}
		allTags = append(allTags, tags...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	var allTagNames []string
	for _, t := range allTags {
		allTagNames = append(allTagNames, *t.Name)
	}

	semver.Sort(allTagNames)

	return allTagNames[len(allTagNames)-1], nil
}

func moduleBump(cmd *cobra.Command, args []string) error {
	if strings.ToLower(providedFlags.version) != "latest" && !semver.IsValid(fmt.Sprintf("v%s", providedFlags.version)) {
		return errVersionFormat
	}

	var err error
	if len(providedFlags.repository) > 0 || providedFlags.allRepositories {
		err = providerAndModuleBumpWorkflow("module", bumpModuleInTerraformFiles)
	} else {
		err = bumpModuleInTerraformFiles()
	}

	if err != nil {
		return err
	}

	return nil
}

func bumpModuleInTerraformFiles(dir ...string) error {
	re := regexp.MustCompile(fmt.Sprintf(`(?U)(source\s*=\s*"tfe\.doctolib\.net\/doctolib\/%v"[^}]*\bversion\s*=\s*")[^"]*(")`, regexp.QuoteMeta(providedFlags.name)))

	tfFiles, err := getTerraformFiles(dir...)
	if err != nil {
		return err
	}
	version := providedFlags.version
	if strings.ToLower(version) == "latest" {
		version, err = getLatestModuleVersion(providedFlags.name)
		if err != nil {
			return err
		}
	}

	repl := []byte("${1}" + version + "${2}")

	for _, file := range tfFiles {
		readFile, err := os.ReadFile(file)
		if err != nil {
			_, _ = cli.PrintlnErr(err)
			continue
		}
		content := re.ReplaceAll(readFile, repl)
		if !bytes.Equal(content, readFile) {
			err = os.WriteFile(file, content, 0600)
			if err != nil {
				_, _ = cli.PrintlnErr(err)
				continue
			}
			if len(dir) == 0 {
				cli.Printf("module bumped in %s\n", file)
			}
		}
	}

	if providedFlags.check {
		_, koDirs := checkModuleUsage(tfFiles, providedFlags.name, providedFlags.version)
		if len(koDirs) > 0 {
			_, _ = cli.PrintfErr("Issues were detected in the following %v directories:\n", len(koDirs))
			for _, dir := range koDirs {
				_, _ = cli.PrintlnErr(dir)
			}
		}
	}

	return nil
}

/*
This function parse the hcl and checks that the given module is at the given version
will return 2 lists of directories the ones at the given version, and all the other ones
*/
func checkModuleUsage(tfFiles []string, moduleName, version string) ([]string, []string) {
	var bumpedDirs, issueDirs []string
	tfDirs := uniqDir(tfFiles)
	for _, tfDir := range tfDirs {
		module, _ := tfconfig.LoadModule(tfDir)
		for _, v := range module.ModuleCalls {
			if v.Source == fmt.Sprintf("tfe.doctolib.net/doctolib/%s", moduleName) {
				if version != v.Version {
					issueDirs = append(issueDirs, tfDir)
				} else {
					bumpedDirs = append(bumpedDirs, tfDir)
				}
			}
		}
	}

	return bumpedDirs, issueDirs
}

// Is given a list of files return a list of dirs with no duplicate
func uniqDir(fileList []string) []string {
	m := make(map[string]bool)
	var result []string
	for _, str := range fileList {
		dir := filepath.Dir(str)
		if m[dir] {
			continue // Already in the map
		}
		result = append(result, dir)
		m[dir] = true
	}
	return result
}
