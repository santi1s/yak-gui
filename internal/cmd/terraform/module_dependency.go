package terraform

import (
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/spf13/cobra"
)

const defaultBranch = "main"

var moduleDependencyCmd = &cobra.Command{
	Use:   "dependency",
	Short: "check modules dependencies",
	RunE:  moduleDependency,
}

func init() {
	moduleDependencyCmd.Flags().IntVarP(&providedFlags.depth, "depth", "d", 10, "max depth")
}

func moduleDependency(_ *cobra.Command, _ []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := fmt.Sprintf("%s/%s", home, repositoriesCacheDir)
	// Get unique list of directories containing *.tf files
	pathsToCheck, err := GetTerraformDirs(".")
	if err != nil {
		return fmt.Errorf("error getting directories containing terraform files")
	}

	for ptc := range pathsToCheck {
		err = ModuleDependenciesSearch(ptc, "", dir, 0)
		if err != nil {
			return err
		}
	}

	return nil
}

func ModuleDependenciesSearch(dir string, repo string, cacheDir string, depth int) error {
	var checkPath string
	if repo == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		cli.Println("module dependencies in: " + path.Join(wd, dir))
		checkPath = dir
	} else {
		checkPath = path.Join(cacheDir, repo)
	}

	modules, err := GetModulesDependencies(checkPath)
	if err != nil {
		return err
	}

	sort.SliceStable(modules, func(i, j int) bool {
		return modules[i].ModuleName()+"@"+modules[i].Version < modules[j].ModuleName()+"@"+modules[j].Version
	})
	for _, m := range modules {
		for i := 0; i < depth; i++ {
			cli.Printf("  ")
		}

		// if module is not a TFE module, it means it's a not a module in TFE registry
		// so we will just print the module but we will not check the dependencies
		if !m.IsTfeModule() {
			cli.Printf(" - %s (%s) [dependencies check are not supported for modules not in TFE registry]\n", m.Source, m.Name)
		} else {
			if depth == providedFlags.depth {
				return nil
			}

			cli.Printf(" - %s@%s (%s)\n", m.ModuleName(), m.Version, m.Name)
			r, err := helper.CloneGitRepository(m.GithubRepository(), cacheDir)
			if err != nil {
				return err
			}
			_, err = helper.CheckoutGitBranch(r, defaultBranch, false, false)
			if err != nil {
				return err
			}

			err = ModuleDependenciesSearch(dir, m.GithubRepository(), cacheDir, depth+1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func GetModulesDependencies(dir string) ([]Module, error) {
	modules := []Module{}
	module, diags := tfconfig.LoadModule(dir)
	if diags.HasErrors() {
		return nil, fmt.Errorf("error reading module %s: %s", dir, diags)
	}
	for _, moduleSpec := range module.ModuleCalls {
		m := Module{
			Name:    moduleSpec.Name,
			Source:  moduleSpec.Source,
			Version: moduleSpec.Version,
		}
		modules = append(modules, m)
	}

	return modules, nil
}
