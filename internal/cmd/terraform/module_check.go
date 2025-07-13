package terraform

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

type ModuleCheckFlags struct {
	allowRelativeSources bool
}

var (
	moduleCheckFlags ModuleCheckFlags
	moduleCheckCmd   = &cobra.Command{
		Use:   "check",
		Short: "check modules",
		RunE:  ModuleCheck,
	}
	errModuleVersionMsg = errors.New("module version is not compliant")
	/* Allowed modules to not contain a version*/
	whitelistedModuleSources = []string{
		"../../../../terraform/modules/datadog-monitors",
		"github-aws-runners/github-runner/aws//modules/multi-runner",
		"philips-labs/github-runner/aws//modules/multi-runner",
		"philips-labs/github-runner/aws//modules/webhook-github-app",
	}
	whitelistedPaths = []string{
		"envs/shared/terraform/02_tfe_staging",
		"envs/shared/terraform/02_tfe",
	}
)

func init() {
	moduleCheckCmd.Flags().BoolVar(&moduleCheckFlags.allowRelativeSources, "allow-relative-sources", false, "Allow modules with relative paths as sources (../ and ./)")
}

func ModuleCheck(cmd *cobra.Command, args []string) error {
	/*Regex to identify a module as a TFE module*/
	reModuleTfeSource := regexp.MustCompile(`tfe\.doctolib\.net\/[A-Za-z_1-9-]+\/[A-Za-z_1-9-]+\/[A-Za-z_1-9_]+$`)
	/*Regex to identify a version is a prerelease version*/
	reModulePrereleaseVersion := regexp.MustCompile(`0.0.0\-pr\d+$`)
	/*Regex to identify a version is an expected semver version*/
	reModuleExpectedVersion := regexp.MustCompile(`\d+\.\d+\.\d+$`)
	// Regex to identify a module from github
	reModuleGithubSource := regexp.MustCompile(`git@github\.com\:[A-Za-z]+\/[A-Za-z_1-9-]+\.git\?[A-Za-z]+\=\d+\.\d+\.\d+$`)
	moduleInError := 0

	// Get unique list of directories containing *.tf files
	pathsToCheck, err := GetTerraformDirs(".")
	if err != nil {
		return fmt.Errorf("error getting directories containing terraform files")
	}

	// Loop over the list of directories and try to load them as Terraform modules to extract the defined modules
	for path := range pathsToCheck {
		module, diags := tfconfig.LoadModule(path)
		if diags.HasErrors() {
			return fmt.Errorf("error reading module %s: %s", path, diags)
		}
		for moduleName, moduleSpec := range module.ModuleCalls {
			err := ModuleVersionCheck(moduleName, path, moduleSpec, reModuleExpectedVersion, reModulePrereleaseVersion, reModuleGithubSource)
			if err != nil {
				moduleInError++
			}
			err = ModuleInsideTFECheck(moduleName, path, moduleSpec, reModuleTfeSource, reModuleGithubSource)
			if err != nil {
				moduleInError++
			}
		}
	}
	// Check if we had errors and print the results
	if moduleInError != 0 {
		return fmt.Errorf("found %d errors in modules, please fix them", moduleInError)
	}
	return nil
}

func ModuleInsideTFECheck(moduleName string, path string, moduleSpec *tfconfig.ModuleCall, reModuleTfeSource *regexp.Regexp, reModuleGithubSource *regexp.Regexp) error {
	if IsRelativePath(moduleSpec) && !moduleCheckFlags.allowRelativeSources && !IsWhitelistedModule(moduleSpec) {
		err := printAnnotationsForModule(path, fmt.Sprintf("module \"%s\"", moduleName), fmt.Sprintf("Module %s relative sources are forbidden, please use --allow-relative-sources", moduleName), 1)
		if err != nil {
			return err
		}
		return errModuleVersionMsg
	}

	if !IsTfeModule(reModuleTfeSource, moduleSpec) &&
		!IsWhitelistedModule(moduleSpec) &&
		!IsWhitelistedPathFromGithub(reModuleGithubSource, moduleSpec, path) &&
		!IsRelativePath(moduleSpec) {
		/*Throw an error because the terraform module is not inside of TFE registry*/
		err := printAnnotationsForModule(path, fmt.Sprintf("module \"%s\"", moduleName), fmt.Sprintf("Module %s is not part of the TFE registry", moduleName), 1)
		if err != nil {
			return err
		}
		return errModuleVersionMsg
	}
	return nil
}

/*Check that the version of a module is semver valid and not a prerelease version*/
func ModuleVersionCheck(moduleName string, path string, moduleSpec *tfconfig.ModuleCall, reModuleExpectedVersion *regexp.Regexp, reModulePrereleaseVersion *regexp.Regexp, reModuleGithubSource *regexp.Regexp) error {
	if moduleSpec.Version == "" {
		/* The module is allowed to not contain a version if it is whitelisted */
		if !IsWhitelistedModule(moduleSpec) && !IsWhitelistedPathFromGithub(reModuleGithubSource, moduleSpec, path) && !IsRelativePath(moduleSpec) {
			err := printAnnotationsForModule(path, fmt.Sprintf("module \"%s\"", moduleName), fmt.Sprintf("Module %s does not have a version", moduleName), 1)
			if err != nil {
				return err
			}
			return errModuleVersionMsg
		}
	} else {
		if IsModuleUsingSemverVersion(moduleSpec) {
			if IsPrereleaseModule(reModulePrereleaseVersion, moduleSpec) {
				err := printAnnotationsForModule(path, fmt.Sprintf("module \"%s\"", moduleName), fmt.Sprintf("Module %s with version %s has a prerelease version", moduleName, moduleSpec.Version), 1)
				if err != nil {
					return err
				}
				return errModuleVersionMsg
			}
			if HasModuleExpectedVersionFormat(reModuleExpectedVersion, moduleSpec) {
				return nil
			}
			err := printAnnotationsForModule(path, fmt.Sprintf("module \"%s\"", moduleName), fmt.Sprintf("Module %s does not have a valid version %s", moduleName, moduleSpec.Version), 1)
			if err != nil {
				return err
			}
			return errModuleVersionMsg
		}
		err := printAnnotationsForModule(path, fmt.Sprintf("module \"%s\"", moduleName), fmt.Sprintf("Module %s with version %s does not have a semver valid version", moduleName, moduleSpec.Version), 1)
		if err != nil {
			return err
		}
		return errModuleVersionMsg
	}
	return nil
}

func IsPrereleaseModule(reModulePrereleaseVersion *regexp.Regexp, moduleSpec *tfconfig.ModuleCall) bool {
	return reModulePrereleaseVersion.MatchString(moduleSpec.Version)
}

func IsTfeModule(reModuleTfeSource *regexp.Regexp, moduleSpec *tfconfig.ModuleCall) bool {
	return reModuleTfeSource.MatchString(moduleSpec.Source)
}

func HasModuleExpectedVersionFormat(reModuleExpectedVersion *regexp.Regexp, moduleSpec *tfconfig.ModuleCall) bool {
	return reModuleExpectedVersion.MatchString(moduleSpec.Version)
}

func IsModuleUsingSemverVersion(moduleSpec *tfconfig.ModuleCall) bool {
	var semverModuleVersion string
	if string(moduleSpec.Version[0]) == "v" {
		semverModuleVersion = moduleSpec.Version
	} else {
		semverModuleVersion = "v" + moduleSpec.Version
	}
	if semver.IsValid(semverModuleVersion) {
		return true
	}
	return false
}

func IsWhitelistedModule(moduleSpec *tfconfig.ModuleCall) bool {
	for _, whitelistedModuleSource := range whitelistedModuleSources {
		if whitelistedModuleSource == moduleSpec.Source {
			return true
		}
	}
	return false
}

func IsWhitelistedPathFromGithub(reModuleExpectedVersion *regexp.Regexp, moduleSpec *tfconfig.ModuleCall, path string) bool {
	for _, whitelistedPath := range whitelistedPaths {
		if whitelistedPath == path && reModuleExpectedVersion.MatchString(moduleSpec.Source) {
			return true
		}
	}
	return false
}

func IsRelativePath(moduleSpec *tfconfig.ModuleCall) bool {
	return strings.HasPrefix(moduleSpec.Source, "../") || strings.HasPrefix(moduleSpec.Source, "./")
}
