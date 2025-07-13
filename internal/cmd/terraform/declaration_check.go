package terraform

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	declarationErrors   int
	declarationCheckCmd = &cobra.Command{
		Use:   "check",
		Short: "Check declaration rules",
		RunE:  declarationCheck,
	}
)

func declarationCheck(cmd *cobra.Command, args []string) error {
	declarationErrors = 0

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
				if err := checkCloudDeclarationFile(path); err != nil {
					return err
				}
			}
			return nil
		})
	if err != nil {
		return (err)
	}

	if declarationErrors != 0 {
		return fmt.Errorf("found %d errors, please fix them", declarationErrors)
	}
	return nil
}

func checkCloudDeclarationFile(path string) error {
	reg := regexp.MustCompile(`^\s+cloud\s+{`)

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
		if path_split[len(path_split)-1] != "backend.tf" {
			err := printAnnotationsForModule(path, "cloud", fmt.Sprintf("cloud declared in '%s' must be declared in `backend.tf`", path), -1)
			declarationErrors++
			if err != nil {
				return err
			}
		}
	}

	return nil
}
