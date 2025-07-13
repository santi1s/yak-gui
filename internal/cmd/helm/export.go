package helm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"

	"github.com/doctolib/yak/internal/helper"
)

const (
	RenderPrefix = "render://"
)

type ExportChart struct {
	Source    *ApplicationSource
	appName   string
	namespace string
}

func HelmDependencyUpdate(dir string) error {
	cmd := exec.Command(
		"helm",
		"dependency",
		"update",
	) // #nosec G204
	cmd.Dir = dir
	_, err := helper.ExecLogger(cmd, true)
	if err != nil {
		return fmt.Errorf("failed to run helm dependency update in %s: %s", dir, err.Error())
	}
	return nil
}

func (opts *ExportChart) ExportHelmChart(workDir string) ([]byte, error) {
	var chart, chdir string

	if err := opts.Source.Validate(); err != nil {
		return nil, err
	}

	// local chart
	if opts.Source.Chart == "." && opts.Source.Path != "" {
		chdir = filepath.Join(workDir, opts.Source.Path)
		if err := HelmDependencyUpdate(chdir); err != nil {
			return nil, err
		}
	}

	if opts.Source.IsHelmOCI() {
		chart = opts.Source.RepoURL
	} else {
		chart = opts.Source.Chart
	}

	args := []string{
		"template",
		opts.appName,
		chart,
		"--namespace",
		opts.namespace,
	}
	if opts.Source.Chart != "" && opts.Source.RepoURL != "" {
		args = append(args, "--repo", opts.Source.RepoURL)
	}

	if opts.Source.TargetRevision != "" {
		args = append(args, "--version", opts.Source.TargetRevision)
	}
	if opts.Source.Helm != nil {
		if opts.Source.Helm.SkipCrds {
			args = append(args, "--skip-crds")
		}

		if opts.Source.Helm.IncludeCrds {
			args = append(args, "--include-crds")
		}

		if !opts.Source.Helm.DontSkipTests {
			args = append(args, "--skip-tests")
		}

		for _, f := range opts.Source.Helm.ValueFiles {
			var helmRenderPluginValues bool
			if strings.HasPrefix(f, RenderPrefix) {
				helmRenderPluginValues = true
				f = strings.TrimPrefix(f, RenderPrefix)
			}
			f = filepath.Join(workDir, f)
			if _, err := os.Stat(f); os.IsNotExist(err) {
				return nil, fmt.Errorf("value file %s does not exist", f)
			}
			if helmRenderPluginValues {
				f = RenderPrefix + f
			}
			args = append(args, "--values", f)
		}

		if opts.Source.Helm.ValuesObject != nil {
			valuesObjectFile, err := os.CreateTemp("", "helm-values-object-*.yaml")
			if err != nil {
				return nil, err
			}
			defer os.Remove(valuesObjectFile.Name())
			enc := yaml.NewEncoder(valuesObjectFile)
			if err := enc.Encode(opts.Source.Helm.ValuesObject); err != nil {
				return nil, err
			}
			enc.Close()
			args = append(args, "--values", valuesObjectFile.Name())
		}

		if opts.Source.Helm.PostRenderer != nil {
			log.Infof("Using post-renderer %s", opts.Source.Helm.PostRenderer.ExecPath)
			f := filepath.Join(workDir, opts.Source.Helm.PostRenderer.ExecPath)
			args = append(args, "--post-renderer", f)
			if len(opts.Source.Helm.PostRenderer.Args) > 0 {
				args = append(args, "--post-renderer-args")
				args = append(args, opts.Source.Helm.PostRenderer.Args...)
			}
		}
	}
	cmd := exec.Command(
		"helm",
		args...,
	) // #nosec G204
	if chdir != "" {
		cmd.Dir = chdir
	}
	buf, err := helper.ExecLogger(cmd, false)
	if err != nil {
		return nil, fmt.Errorf("failed to run helm template: %s", err.Error())
	}
	return buf.Bytes(), nil
}

func writeManifestsToFile(manifests []byte, path string) error {
	err := os.WriteFile(path, manifests, 0644) // #nosec G306
	if err != nil {
		return fmt.Errorf("failed to write manifests to file: %s", err.Error())
	}
	log.Infof("Manifests written to %s", path)
	return nil
}

func (applications *ApplicationsValues) Export(workDir string, exportDir string) (int, error) {
	errors := 0
	var managedFiles []string
	env := applications.ExtractEnv()
	for appName, config := range applications.Applications {
		if providedHelmFlags.application != "" && appName != providedHelmFlags.application {
			continue
		}
		exports := config.Export(appName)
		if len(exports) == 0 {
			log.Infof("No helm chart to export in %s", appName)
			continue
		}

		log.Infof("Generating manifests for %s in %s", appName, env)

		manifestDir := filepath.Join(exportDir, "envs", env, "helm", appName)
		if err := os.MkdirAll(manifestDir, 0755); err != nil {
			log.Errorf("failed to create directory: %s", err.Error())
			errors++
			continue
		}
		exportErrors := 0
		for _, export := range exports {
			manifest, err := export.ExportHelmChart(workDir)
			if err != nil {
				log.Errorf("failed to export helm chart: %s", err.Error())
				errors++
				exportErrors++
				continue
			}
			manifestFile := filepath.Join(
				manifestDir,
				"export_"+export.Source.ExportName+".yml")
			if err := writeManifestsToFile(manifest, manifestFile); err != nil {
				log.Errorf("failed to write manifests to file: %s", err.Error())
				errors++
				continue
			}
			managedFiles = append(managedFiles, manifestFile)
		}
		if exportErrors > 0 {
			continue
		}
	}
	// Search for unmanaged files
	if providedHelmFlags.application != "" {
		log.Warn("Skipping file checks because an application is specified")
		return errors, nil
	}
	allFiles, err := filepath.Glob(filepath.Join(
		exportDir,
		"envs",
		env,
		"helm",
		"*",
		"*.*"))
	if err != nil {
		errors++
		return errors, fmt.Errorf("failed to list files: %s", err.Error())
	}
	for _, file := range allFiles {
		found := false
		for _, managedFile := range managedFiles {
			if managedFile == file {
				found = true
				continue
			}
		}
		if !found {
			if os.Getenv("REMOVE_UNMANAGED") == "true" {
				if err := os.Remove(file); err != nil {
					errors++
					log.Errorf("failed to remove file: %s", err.Error())
				}
				log.Infof("Removed unmanaged file %s", file)
			} else {
				log.Errorf("Found unmanaged file %s", file)
				errors++
			}
		}
	}
	return errors, nil
}
