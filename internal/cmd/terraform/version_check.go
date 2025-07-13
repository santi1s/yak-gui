package terraform

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type versionCheckFlags struct {
	cfgFile string
}

var (
	errorCount                int
	providedVersionCheckFlags versionCheckFlags
	versionCheckCmd           = &cobra.Command{
		Use:   "check",
		Short: "Checks terraform version",
		RunE:  versionCheck,
	}
)

func init() {
	versionCheckCmd.Flags().StringVarP(&providedVersionCheckFlags.cfgFile, "file", "f", "", "yaml file containing terraform allowed versions")
	err := versionCheckCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}
}

func versionCheck(cmd *cobra.Command, args []string) error {
	type versionConstraint struct {
		operator string
		version  *semver.Version
	}

	// Get terraform version constraint file and load yaml
	viper.SetConfigFile(providedVersionCheckFlags.cfgFile)
	cobra.CheckErr(viper.MergeInConfig())

	versionDeprecated, _ := semver.NewVersion(viper.GetString("deprecated"))
	versionStable, _ := semver.NewVersion(viper.GetString("stable"))
	versionNext, _ := semver.NewVersion(viper.GetString("next"))

	// Get unique list of directories containing *.tf files
	pathsToCheck := make(map[string]bool)
	err := filepath.WalkDir(".",
		func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() &&
				strings.HasSuffix(path, ".tf") &&
				!strings.Contains(path, ".terraform/") &&
				!strings.Contains(path, "terraform/ci") &&
				!strings.Contains(path, "test/") &&
				!strings.Contains(path, ".git") {
				pathsToCheck[filepath.Dir(path)] = true
			}
			return nil
		})
	if err != nil {
		return (err)
	}

	errorCount = 0

	// Loop over the list of directories and try to load them as Terraform modules to see which terraform version is defined
	for path := range pathsToCheck {
		module, diags := tfconfig.LoadModule(path)
		if diags.HasErrors() {
			return fmt.Errorf("error reading module %s: %s", path, diags)
		}

		onlyExclusions := true
		constraintIsRange := false

		if len(module.RequiredCore) == 0 {
			err := printAnnotationsForModule(path, "terraform {", fmt.Sprintf("Terraform version is not defined in module %s", path), -1)
			errorCount++
			if err != nil {
				return err
			}
		}

		for _, rcString := range module.RequiredCore {
			c, err := semver.NewConstraint(rcString)
			if err != nil {
				err := printAnnotationsForModule(path, "required_version =", fmt.Sprintf("Version constraint empty or malformed for terraform version in module %s: %s", path, err), -1)
				errorCount++
				if err != nil {
					return err
				}
				onlyExclusions = false
				continue
			}
			if !c.Check(versionStable) && !c.Check(versionNext) {
				if !c.Check(versionDeprecated) {
					err := printAnnotationsForModule(path, "required_version =", fmt.Sprintf("Terraform version defined does not match any allowed version in module %s", path), -1, "warning")
					if err != nil {
						return err
					}
				} else {
					err := printAnnotationsForModule(path, "required_version =", fmt.Sprintf("Terraform version defined only matches a deprecated version in module %s", path), -1, "warning")
					if err != nil {
						return err
					}
				}
			}

			var minBound, maxBound versionConstraint

			constraintStrings := strings.Split(rcString, ",")
			for _, constraintString := range constraintStrings {
				var constraint versionConstraint
				matches := vcExp.FindStringSubmatch(constraintString)
				constraint.operator = matches[vcExp.SubexpIndex("operator")]
				constraint.version, _ = semver.NewVersion(matches[vcExp.SubexpIndex("version")])

				if constraint.operator == "!=" {
					// If exclusion is defined, we skip ; onlyExclusions keeps its value to check it is not the only rule defined
					continue
				} else {
					onlyExclusions = false
					if constraint.operator == "~>" {
						// "rightmost" operator is always allowed
						break
					} else if constraint.operator == "" || constraint.operator == "=" {
						// Else, if specific version is found, it's not allowed
						err := printAnnotationsForModule(path, "required_version =", fmt.Sprintf("Fixed terraform version are not allowed in module %s", path), -1)
						errorCount++
						if err != nil {
							return err
						}
						constraintIsRange = false
						break
					} else if len(constraintStrings) == 1 {
						// Else, if we only have a single constraint, that means the range is too large
						err := printAnnotationsForModule(path, "required_version =", fmt.Sprintf("Version range must be defined with two bounds for terraform version in module %s", path), -1)
						errorCount++
						if err != nil {
							return err
						}
						constraintIsRange = false
						break
					} else {
						constraintIsRange = true
						// Let's find min and max bounds to determine the range width
						if constraint.operator == ">" || constraint.operator == ">=" {
							if (minBound != versionConstraint{}) {
								// We found two > / >= bounds in the same VersionConstraint string
								err := printAnnotationsForModule(path, "required_version =", fmt.Sprintf("Version range should define only one minimum bound for terraform version in module %s", path), -1, "warning")
								errorCount++
								if err != nil {
									return err
								}
								// This is a range but we want to stop evaluating it and return only the previous error
								constraintIsRange = false
								break
							}
							minBound = constraint
						} else if constraint.operator == "<" || constraint.operator == "<=" {
							if (maxBound != versionConstraint{}) {
								// We found two < / <= bounds in the same VersionConstraint string
								err := printAnnotationsForModule(path, "required_version =", fmt.Sprintf("Version range should define only one maximum bound for terraform version in module %s", path), -1, "warning")
								errorCount++
								if err != nil {
									return err
								}
								// This is a range but we want to stop evaluating it and return only the previous error
								constraintIsRange = false
								break
							}
							maxBound = constraint
						}
					}
				}
			}
			if onlyExclusions {
				err := printAnnotationsForModule(path, "required_version =", fmt.Sprintf("Version constraints cannot contain only exclusions for terraform version in module %s", path), -1)
				errorCount++
				if err != nil {
					return err
				}
			} else if constraintIsRange {
				// In any case, if the range is more than 2 major wide, we have an error (ex: ">= 1.2.3, < 3.0.0")
				nextMajor := minBound.version.IncMajor().IncMajor()
				if maxBound.version.GreaterThan(&nextMajor) {
					err := printAnnotationsForModule(path, "required_version =", fmt.Sprintf("Maximum bound cannot be more than two Major versions above minimum bound for terraform version in module %s", path), -1)
					errorCount++
					if err != nil {
						return err
					}
				} else if maxBound.operator != "<" {
					// If operator is not "<" we might end with a range too wide as we want to exclude next major version for terraform versions (ex: ">= 1.2.3, <= 2.0.0")
					err := printAnnotationsForModule(path, "required_version =", fmt.Sprintf("Version constraint can only be defined to two next Major version - excluded - for terraform version in module %s", path), -1)
					errorCount++
					if err != nil {
						return err
					}
				}
			}
		}
		log.Debugf("terraform version in module %s: OK\n", path)
	}

	// Check if we had errors and print the results
	if errorCount != 0 {
		return fmt.Errorf("found %d errors for terraform version definition, please fix them", errorCount)
	}
	return nil
}
