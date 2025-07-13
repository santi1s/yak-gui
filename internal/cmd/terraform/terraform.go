package terraform

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/constant"
	"github.com/doctolib/yak/internal/helper"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type terraformFlags struct {
	cfgFile         string
	description     string
	name            string
	release         string
	version         string
	allRepositories bool
	check           bool
	skipConfirm     bool
	repository      []string
	parallelism     int
	dryRun          bool
	branch          string
	depth           int
}

var (
	providedFlags terraformFlags
	terraformCmd  = &cobra.Command{
		Use:   "terraform",
		Short: "manage terraform",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Root().Name() == constant.CliName && cmd.Root().PersistentPreRun != nil {
				cmd.Root().PersistentPreRun(cmd, args)
			}

			if cmd.HasParent() && cmd.Parent().Name() == "completion" {
				switch cmd.Name() {
				case "bash", "zsh", "fish", "powershell":
					cmd.ResetFlags()
				}
			}

			cmd.SilenceUsage = true
		},
	}
)

// Walk recursively through the current or provided directory
// Return a list of *.tf files found and an error
func getTerraformFiles(dir ...string) ([]string, error) {
	wd := "."
	if len(dir) > 0 {
		wd = dir[0]
	}
	var terraformFiles []string
	err := filepath.WalkDir(wd,
		func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() &&
				strings.HasSuffix(path, ".tf") &&
				!strings.Contains(path, ".terraform/") &&
				!strings.Contains(path, "terraform/ci") &&
				!strings.Contains(path, ".git") {
				terraformFiles = append(terraformFiles, path)
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return terraformFiles, nil
}

// Walk recursively through the provided directory
// Return a directories containing *.tf files and an error if any
func GetTerraformDirs(wd string) (map[string]bool, error) {
	// Get unique list of directories containing *.tf files
	pathsToCheck := make(map[string]bool)
	err := filepath.WalkDir(wd,
		func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() &&
				strings.HasSuffix(path, ".tf") &&
				!strings.Contains(path, ".terraform/") &&
				!strings.Contains(path, "terraform/ci") {
				pathsToCheck[filepath.Dir(path)] = true
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return pathsToCheck, nil
}

// Try to print GitHub annotations for a specific module
// File and line numbers are found using the pattern provided and searching under path
// In case a pattern is matched, print provided message (taking care of newlines)
// If the pattern is not matched, fallback to printing the message directly
func printAnnotationsForModule(path string, pattern string, message string, maxDepth int, level ...string) error {
	var lvl string
	if len(level) != 0 && (level[0] == "warning" || level[0] == "notice") {
		lvl = level[0]
	} else {
		lvl = "error"
	}
	annotations, err := helper.FindStringInPathAndGetLineNumber(path, pattern, maxDepth)
	if err != nil {
		return err
	}
	if len(annotations) == 0 {
		log.Debugf("Could not create annotation for %s: %s, pattern not matched: %s in path: %s\n", lvl, message, pattern, path)
		cli.Println(message)
	} else {
		for _, annotation := range annotations {
			_, err = cli.Printf("::%s file=%s,line=%d::%s\n", lvl, annotation.Path, annotation.Line, message)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func GetRootCmd() *cobra.Command {
	return terraformCmd
}

func init() {
	// Set config file to be read at command execution
	terraformCmd.PersistentFlags().StringVarP(&providedFlags.cfgFile, "config", "c", "", "config file")
	_ = viper.BindPFlag("config", terraformCmd.PersistentFlags().Lookup("config"))

	terraformCmd.AddCommand(moduleCmd)
	terraformCmd.AddCommand(providerCmd)
	terraformCmd.AddCommand(versionCmd)
	terraformCmd.AddCommand(reportCmd)
	terraformCmd.AddCommand(declarationCmd)
}
