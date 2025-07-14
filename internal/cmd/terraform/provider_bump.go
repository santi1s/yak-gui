package terraform

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Masterminds/semver"
	"github.com/santi1s/yak/cli"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/spf13/cobra"
)

var (
	providerBumpCmd = &cobra.Command{
		Use:   "bump",
		Short: "bump provider version",
		RunE:  providerBump,
	}
)

func providerBump(cmd *cobra.Command, args []string) error {
	_, err := semver.NewConstraint(providedFlags.version)
	if err != nil {
		return fmt.Errorf("could not parse version constraint: %s", err)
	}

	if len(providedFlags.repository) > 0 || providedFlags.allRepositories {
		err = providerAndModuleBumpWorkflow("provider", bumpProviderInTerraformFiles)
	} else {
		err = bumpProviderInTerraformFiles()
	}

	if err != nil {
		return err
	}

	return nil
}

func bumpProviderInTerraformFiles(dir ...string) error {
	moduleDirs := make(map[string]bool)

	// Get unique list of directories containing *.tf files
	pathsToCheck, err := getTerraformFiles(dir...)
	if err != nil {
		return (err)
	}

	// This regex tries to match, case insensitive mode,
	// a line with `source = "<provider_name>"`
	// and a line `version = "<anything>"`
	// with any caracter between them, outside of `}`, which could be a block end
	// Capture what's before and after version defined, to be used in replacement function
	reSrcVer := regexp.MustCompile(fmt.Sprintf(`(?i)(source\s*=\s*"%v"[^}]*version\s*=\s*")[^"]*(")`, regexp.QuoteMeta(providedFlags.name)))
	// This regex tries to match the opposite
	reVerSrc := regexp.MustCompile(fmt.Sprintf(`(?i)(\bversion\s*=\s*")[^"]*("[^"}]*source\s*=\s*"%v")`, regexp.QuoteMeta(providedFlags.name)))
	repl := []byte("${1}" + providedFlags.version + "${2}")

	for _, path := range pathsToCheck {
		readFile, err := os.ReadFile(path)
		if err != nil {
			_, _ = cli.PrintlnErr(err)
		}
		content := reVerSrc.ReplaceAll(reSrcVer.ReplaceAll(readFile, repl), repl)
		if !bytes.Equal(content, readFile) {
			err = os.WriteFile(path, content, 0600)
			if err != nil {
				_, _ = cli.PrintlnErr(err)
			}
			if len(dir) == 0 {
				cli.Printf("provider bumped in %s\n", path)
			}
			moduleDirs[filepath.Dir(path)] = true
		}
	}

	if providedFlags.check {
		_, koDirs, err := checkProviderUsage(moduleDirs, providedFlags.name, providedFlags.version)
		if err != nil {
			return err
		}
		if len(koDirs) > 0 {
			_, _ = cli.PrintfErr("issues were detected in the following %v directories:\n", len(koDirs))
			for _, dir := range koDirs {
				_, _ = cli.PrintlnErr(dir)
			}
			return errors.New("could not bump all provider definitions")
		}
	}

	return nil
}

// Parse the hcl and check that the given provider has the given version constraint defined
// Return 2 lists of directories the ones that match version constraint, and the ones that don't
func checkProviderUsage(moduleDirs map[string]bool, providerName, version string) ([]string, []string, error) {
	var bumpedDirs, issueDirs []string
	for path := range moduleDirs {
		module, diags := tfconfig.LoadModule(path)
		if diags.HasErrors() {
			return nil, nil, fmt.Errorf("error reading module %s: %s", path, diags)
		}
		providerFound := false
		for _, v := range module.RequiredProviders {
			if v.Source == providerName {
				providerFound = true
				foundInVersionConstraints := false
				for _, vc := range v.VersionConstraints {
					if vc == version {
						foundInVersionConstraints = true
					}
				}
				if !foundInVersionConstraints {
					issueDirs = append(issueDirs, path)
				} else {
					bumpedDirs = append(bumpedDirs, path)
				}
			}
		}
		if !providerFound {
			issueDirs = append(issueDirs, path)
		}
	}

	return bumpedDirs, issueDirs, nil
}

func init() {
	providerBumpCmd.Flags().StringVarP(&providedFlags.name, "name", "n", "", "name of the provider to bump")
	providerBumpCmd.Flags().StringVarP(&providedFlags.version, "version", "v", "", "version constraint to use for the provider")
	providerBumpCmd.Flags().BoolVar(&providedFlags.check, "check", false, "run a version check after bump")
	providerBumpCmd.Flags().StringVarP(&providedFlags.release, "release", "R", "", "release type (patch, minor or major)")
	providerBumpCmd.Flags().StringVarP(&providedFlags.description, "description", "d", "", "pull-request description")
	providerBumpCmd.Flags().StringSliceVarP(&providedFlags.repository, "repository", "r", nil, "bump module on specific module repositories list")
	providerBumpCmd.Flags().BoolVar(&providedFlags.allRepositories, "all-repos", false, "bump module in all terraform-* repositories")
	providerBumpCmd.Flags().BoolVar(&providedFlags.skipConfirm, "skip-confirm", false, "will skip confirmation for bump on specfici module repositories")
	providerBumpCmd.Flags().IntVarP(&providedFlags.parallelism, "parallelism", "p", 8, "number of parallel tasks when --repository or --all-repos is used")
	providerBumpCmd.Flags().StringVarP(&providedFlags.branch, "branch", "b", "", "specify a custom branch name for the pull-request creation when using --repository or --all-repos")

	providerBumpCmd.MarkFlagsMutuallyExclusive("repository", "all-repos")

	err := providerBumpCmd.MarkFlagRequired("name")
	if err != nil {
		panic(err)
	}
	err = providerBumpCmd.MarkFlagRequired("version")
	if err != nil {
		panic(err)
	}
}
