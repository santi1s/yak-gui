package repo

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/doctolib/yak/cli"
	"github.com/spf13/cobra"
)

var errFilesNotCorrectlyFormatted = errors.New("files are not correctly formatted")

func secretFmt(cmd *cobra.Command, args []string) error {
	err := checkRepositoryPath()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	ymlReferenceFiles, err := filepath.Glob(cwd + "/configs/vault-secrets/*/*.yml")
	if err != nil {
		return err
	}

	if len(ymlReferenceFiles) == 0 {
		return errNoReferenceFilesFound
	}

	wd, _ := os.Getwd()

	success := true
	for _, f := range ymlReferenceFiles {
		yml, err := os.ReadFile(f)
		if err != nil {
			return err
		}

		b1 := []byte(yml)

		logicalSecret := &LogicalSecret{}
		err = UnmarshalRefFile(f, logicalSecret)
		if err != nil {
			return err
		}

		for path := range logicalSecret.Secrets {
			var secret Secret

			duplicates := make(map[string]bool)
			keys := []string{}

			for _, v := range logicalSecret.Secrets[path].Keys {
				if _, value := duplicates[v]; !value {
					duplicates[v] = true
					keys = append(keys, v)
				}
			}

			secret.Keys = keys
			sort.Strings(secret.Keys)

			switch logicalSecret.Secrets[path].Version.(type) {
			case string:
				secret.Version, err = strconv.Atoi(logicalSecret.Secrets[path].Version.(string))
				if err != nil {
					secret.Version = logicalSecret.Secrets[path].Version.(string)
				}
			case int:
				secret.Version = logicalSecret.Secrets[path].Version.(int)
			}

			logicalSecret.Secrets[path] = secret
		}

		b2 := bytes.Buffer{}
		err = EncodeLogicalSecretToYmlReferenceFile(*logicalSecret, &b2)
		if err != nil {
			return err
		}

		f1, err := os.CreateTemp("", "yak-kube-secret-fmt-diff")
		if err != nil {
			return err
		}
		defer os.Remove(f1.Name())
		defer f1.Close()

		f2, err := os.CreateTemp("", "yak-kube-secret-fmt-diff")
		if err != nil {
			return err
		}
		defer os.Remove(f2.Name())
		defer f2.Close()

		_, err = f1.Write(b1)
		if err != nil {
			return err
		}

		_, err = f2.Write(b2.Bytes())
		if err != nil {
			return err
		}

		if !bytes.Equal(b1, b2.Bytes()) {
			success = false
			filename := strings.Replace(f, wd+"/", "", 1)
			data, err := exec.Command("diff", "--label=old/"+filename, "--label=new/"+filename, "-u", f1.Name(), f2.Name()).CombinedOutput() //#nosec
			if len(data) > 0 {
				cli.Println(filename)             // always print the filename where a diff is found (and fix if not in check mode)
				if providedRepoSecretFlags.diff { // only print the diff if diff flag is enabled
					cli.Println(string(data))
				}
			} else if err != nil {
				return err
			}

			if !providedRepoSecretFlags.check {
				err = writeYmlFile(*logicalSecret, f)
				if err != nil {
					return err
				}
			}
		}
	}

	if !success && providedRepoSecretFlags.check {
		return errFilesNotCorrectlyFormatted
	}

	return nil
}

var secretFmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "reformat logical secret resources files",
	RunE:  secretFmt,
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true
	},
	Args: cobra.ExactArgs(0),
}

func init() {
	secretFmtCmd.Flags().BoolVar(&providedRepoSecretFlags.diff, "diff", false, "display diffs of formatting changes")
	secretFmtCmd.Flags().BoolVar(&providedRepoSecretFlags.check, "check", false, "check if the files are formatted. Exit status will be 0 if all input is properly formatted and non-zero otherwise.")
}
