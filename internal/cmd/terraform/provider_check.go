package terraform

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type providerCheckFlags struct {
	cfgFile string
}

var (
	providerInError            int
	providedProviderCheckFlags providerCheckFlags
	providerCheckCmd           = &cobra.Command{
		Use:   "check",
		Short: "check provider usage",
		RunE:  providerCheck,
	}
)

var vcExp = regexp.MustCompile(`^\s*(?P<operator>[<>!=~]{0,2})\s?(?P<version>.*)\s*$`)

func providerCheck(cmd *cobra.Command, args []string) error {
	type versionConstraint struct {
		operator string
		version  *semver.Version
	}

	// Load list of allowed providers
	viper.SetConfigFile(providedProviderCheckFlags.cfgFile)
	cobra.CheckErr(viper.MergeInConfig())

	// Get unique list of directories containing *.tf files
	pathsToCheck := make(map[string]bool)
	providerInError = 0
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

				if err := checkProviderDeclarationFile(path); err != nil {
					return err
				}

				if err := checkRequiredProviderDeclaration(path); err != nil {
					return err
				}
			}
			return nil
		})
	if err != nil {
		return (err)
	}

	// Loop over the list of directories and try to load them as Terraform modules to see which providers are defined
	for path := range pathsToCheck {
		module, diags := tfconfig.LoadModule(path)
		if diags.HasErrors() {
			return fmt.Errorf("error reading module %s: %s", path, diags)
		}

		for providerName, providerReq := range module.RequiredProviders {
			// states using terraform_remote_state datasources happen to be seen with a provider terraform, we don't want to check that...
			if providerName == "terraform" && providerReq.Source == "" && len(providerReq.VersionConstraints) == 0 {
				annotations, err := helper.FindStringInPathAndGetLineNumber(path, "terraform_remote_state", -1)
				if err != nil {
					return err
				}
				if len(annotations) != 0 {
					continue
				}

				annotations, err = helper.FindStringInPathAndGetLineNumber(path, "terraform_data", -1)
				if err != nil {
					return err
				}
				if len(annotations) != 0 {
					continue
				}
			}
			// states using tfe_outputs datasources happen to be seen with a provider tfe, we don't want to check that...
			if providerName == "tfe" && providerReq.Source == "" && len(providerReq.VersionConstraints) == 0 {
				annotations, err := helper.FindStringInPathAndGetLineNumber(path, "tfe_outputs", -1)
				if err != nil {
					return err
				}
				if len(annotations) != 0 {
					continue
				}
			}

			if providerReq.Source == "" {
				// Provider is defined without any source

				err := printAnnotationsForModule(path, fmt.Sprintf("%s = {", providerName), fmt.Sprintf("No source defined for provider %s in module %s", providerName, path), -1)
				providerInError++
				if err != nil {
					return err
				}
			} else if len(viper.GetStringMap("terraform_providers."+providerReq.Source)) == 0 && len(viper.GetStringMap("terraform_providers.hashicorp/"+providerReq.Source)) == 0 {
				// Provider is not found in allowed providers list
				err := printAnnotationsForModule(path, providerReq.Source, fmt.Sprintf("Forbidden source %s for provider %s in module %s", providerReq.Source, providerName, path), -1)
				providerInError++
				if err != nil {
					return err
				}
			} else if len(providerReq.VersionConstraints) == 0 || (len(providerReq.VersionConstraints) == 1 && providerReq.VersionConstraints[0] == "") {
				// Provider has no version constraint defined
				err := printAnnotationsForModule(path, fmt.Sprintf("%s = {", providerName), fmt.Sprintf("No version defined for provider %s in module %s", providerName, path), -1)
				providerInError++
				if err != nil {
					return err
				}
			} else {
				// For clarity and ease of check, all version constraint strings should be compliant to rules ;
				// string containing only exclusions (!= operator) are tolerated if not the only string defined for the provider.
				// Rules checked are:
				// - specific version is defined
				// or
				// - range is defined according to level of trust (to next major if trusted, to next minor otherwise)

				// providerReq.VersionConstraints: [">= 1.2.0, != 1.2.3, < 2.0.0", ">= 1.2.1"]
				// vcString: ">= 1.2.0, != 1.2.3, < 2.0.0"
				// versionConstraint: ">= 1.2.0"
				// versionConstraint.operator: ">="
				// versionConstraint.version: semver.Version("1.2.0")
				onlyExclusions := true
				constraintIsRange := false
				for _, vcString := range providerReq.VersionConstraints {
					_, err := semver.NewConstraint(vcString)
					if err != nil {
						err := printAnnotationsForModule(path, fmt.Sprintf("%s = {", providerName), fmt.Sprintf("Version constraint empty or malformed for provider %s in module %s: %s", providerName, path, err), -1)
						providerInError++
						if err != nil {
							return err
						}
						onlyExclusions = false
						continue
					}
					var minBound, maxBound versionConstraint

					constraintStrings := strings.Split(vcString, ",")
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
								// Else, if specific version is found, it should be the only one defined in the VersionConstraint string
								// and we should have only one VersionConstraint string because other conditions will not be taken into account
								if len(providerReq.VersionConstraints) > 1 || len(constraintStrings) > 1 {
									err := printAnnotationsForModule(path, fmt.Sprintf("%s = {", providerName), fmt.Sprintf("Version ranges and exclusions should not be defined alongside specific version for provider %s in module %s", providerName, path), -1, "warning")
									if err != nil {
										return err
									}
									constraintIsRange = false
								}
								break
							} else if len(constraintStrings) == 1 {
								// Else, if we only have a single constraint, that means the range is too large
								err := printAnnotationsForModule(path, fmt.Sprintf("%s = {", providerName), fmt.Sprintf("Version range must be defined with two bounds for provider %s in module %s", providerName, path), -1)
								providerInError++
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
										err := printAnnotationsForModule(path, fmt.Sprintf("%s = {", providerName), fmt.Sprintf("Version range should define only one minimum bound for provider %s in module %s", providerName, path), -1, "warning")
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
										err := printAnnotationsForModule(path, fmt.Sprintf("%s = {", providerName), fmt.Sprintf("Version range should define only one maximum bound for provider %s in module %s", providerName, path), -1, "warning")
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
						err := printAnnotationsForModule(path, fmt.Sprintf("%s = {", providerName), fmt.Sprintf("Version constraints cannot contain only exclusions for provider %s in module %s", providerName, path), -1)
						providerInError++
						if err != nil {
							return err
						}
					} else if constraintIsRange {
						// In any case, if the range is more than 2 major wide, we have an error (ex: ">= 1.2.3, < 3.0.0")
						nextMajor := minBound.version.IncMajor().IncMajor()
						if maxBound.version.GreaterThan(&nextMajor) {
							err := printAnnotationsForModule(path, fmt.Sprintf("%s = {", providerName), fmt.Sprintf("Maximum bound cannot be more than two Major versions above minimum bound for provider %s in module %s", providerName, path), -1)
							providerInError++
							if err != nil {
								return err
							}
						} else {
							// Let's get the trusted level of the provider
							trusted := viper.GetBool("terraform_providers." + providerReq.Source + ".trusted")
							if trusted {
								// If operator is not "<" we might end with a range too wide as we want to exclude next major version for trusted providers (ex: ">= 1.2.3, <= 2.0.0")
								if maxBound.operator != "<" {
									err := printAnnotationsForModule(path, fmt.Sprintf("%s = {", providerName), fmt.Sprintf("Version constraint can only be defined to two next Major version - excluded - for trusted provider %s in module %s", providerName, path), -1)
									providerInError++
									if err != nil {
										return err
									}
								}
							} else {
								// If major digits are different, it means the range is too wide for untrusted providers (ex: ">= 1.2.3, < 2.0.0")
								if minBound.version.Major() != maxBound.version.Major() {
									err := printAnnotationsForModule(path, fmt.Sprintf("%s = {", providerName), fmt.Sprintf("Version constraint must use the same Major digit for untrusted provider %s in module %s", providerName, path), -1)
									providerInError++
									if err != nil {
										return err
									}
								}
							}
						}
						// }
					}
				}
				log.Debugf("provider %s (source: \"%s\") in module %s: OK\n", providerName, providerReq.Source, path)
			}
		}
	}

	// Check if we had errors and print the results
	if providerInError != 0 {
		return fmt.Errorf("found %d errors in providers, please fix them", providerInError)
	}
	return nil
}

// Check if required_provider are declared in versions.tf
func checkRequiredProviderDeclaration(path string) error {
	reg := regexp.MustCompile(`\s+required_providers\s+{`)

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !reg.MatchString(line) {
			continue
		}

		path_split := strings.Split(path, "/")
		if path_split[len(path_split)-1] != "versions.tf" {
			err := printAnnotationsForModule(path, "required_providers", fmt.Sprintf("required provider declared in '%s' must be declared in 'versions.tf'.", path), -1)
			providerInError++
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Check if a provider is declared in `provider_name.tf`
func checkProviderDeclarationFile(path string) error {
	reg := regexp.MustCompile(`^provider\s+"([^"]+)"`)

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !reg.MatchString(line) {
			continue
		}

		matches := reg.FindStringSubmatch(line)
		if len(matches) != 2 {
			return fmt.Errorf("too much provider name find in %s", path)
		}

		provider_name := matches[1]
		path_split := strings.Split(path, "/")
		filename := path_split[len(path_split)-1]
		if filename == "provider_"+provider_name+".tf" ||
			filename == "provider_"+provider_name+".autogenerated.tf" {
			return nil
		}

		err := printAnnotationsForModule(path, "provider", fmt.Sprintf("provider '%s' in '%s' should be declared in 'provider_%s.tf'", provider_name, path, provider_name), -1)
		providerInError++
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	providerCheckCmd.Flags().StringVarP(&providedProviderCheckFlags.cfgFile, "file", "f", "", "yaml file containing providers with trusted level")
	err := providerCheckCmd.MarkFlagRequired("file")
	if err != nil {
		panic(err)
	}
}
