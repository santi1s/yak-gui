package terraform

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/google/go-github/v73/github"
)

type BumpResult struct {
	Repository string `json:"-"`
	Error      string `json:"error,omitempty"`
	URL        string `json:"url,omitempty"`
}

func getTerraformModulesRepositoriesNameList(gh *github.Client) ([]string, error) {
	repositories := []*github.Repository{}
	page := 1
	for {
		r, _, err := gh.Search.Repositories(context.Background(), "org:doctolib in:name terraform", &github.SearchOptions{
			ListOptions: github.ListOptions{Page: page, PerPage: 100},
		})

		if err != nil {
			return nil, fmt.Errorf("error while searching repositories: %v", err)
		}

		repositories = append(repositories, r.Repositories...)
		page++

		if len(repositories) >= *r.Total {
			break
		}
	}

	modulesRepositories := []string{}
	for _, r := range repositories {
		// Ignore some repositories because they are not terraform modules but are returned by Github API
		// when we search for terraform and the API does not support complex search
		switch *r.Name {
		case "cookbook-terraform", "terraform-data-team":
			continue
		default:
			modulesRepositories = append(modulesRepositories, *r.Name)
		}
	}

	return modulesRepositories, nil
}

func cloneRepositoryAndGetWorktree(gh *github.Client, agent *ssh.PublicKeysCallback, repo string, branch string, dir string) (*git.Repository, *git.Worktree, error) {
	var err error

	r, err := helper.CloneGitRepository(repo, dir)
	if err != nil {
		return nil, nil, err
	}

	wt, err := helper.CheckoutGitBranch(r, defaultBranch, false, false)
	if err != nil {
		return nil, nil, err
	}

	err = helper.PullGitBranch(wt)
	if err != nil {
		return nil, nil, err
	}

	wt, err = helper.CheckoutGitBranch(r, branch, true, false)
	if err != nil {
		wt, err = helper.CheckoutGitBranch(r, branch, false, false)

		if err != nil {
			return r, nil, err
		}
	}

	return r, wt, nil
}

func commitChangesAndCreatePullRequest(repository *git.Repository, worktree *git.Worktree, gh *github.Client, repo string, branch string, dir string, bumpType string) (int, error) {
	status, err := worktree.Status()
	if err != nil {
		return -1, err
	}

	if status.IsClean() {
		return -1, nil
	}

	cmd := exec.Command("terraform-docs", path.Join(dir, repo)) //#nosec
	_, err = cmd.CombinedOutput()
	if err != nil {
		return -1, err
	}

	err = worktree.AddGlob("*")
	if err != nil {
		return -1, err
	}

	_, err = worktree.Commit(fmt.Sprintf("chore: bump %s %s to version %s", bumpType, providedFlags.name, providedFlags.version), &git.CommitOptions{All: true})
	if err != nil {
		return -1, err
	}

	err = repository.Push(&git.PushOptions{})
	if err != nil {
		return -1, err
	}

	pr, _, err := gh.PullRequests.Create(context.Background(), "doctolib", repo, &github.NewPullRequest{
		Title:               github.Ptr(fmt.Sprintf("chore: bump %s %s to version %s", bumpType, providedFlags.name, providedFlags.version)),
		Head:                github.Ptr(branch),
		Base:                github.Ptr("main"),
		Body:                github.Ptr(fmt.Sprintf("# Description\nBump %s %s to version %s\n\n%s\n# Context\n%s bumped with YAK CLI\n\n# Dependencies\nno", bumpType, providedFlags.name, providedFlags.version, providedFlags.description, bumpType)),
		MaintainerCanModify: github.Ptr(true),
		Draft:               github.Ptr(false),
	})
	if err != nil {
		return -1, err
	}

	_, _, _ = gh.Issues.AddLabelsToIssue(context.Background(), "doctolib", repo, pr.GetNumber(), []string{"release:" + providedFlags.release})
	return pr.GetNumber(), err
}

func providerAndModuleBumpWorkflow(bumpType string, bumpFunc func(dir ...string) error) error {
	switch providedFlags.release {
	case "patch", "minor", "major":
		break
	default:
		return errMissingRelease
	}

	var err error

	gh, err := helper.GetGithubClient()
	if err != nil {
		return err
	}

	var repos = []string{}
	if len(providedFlags.repository) > 0 {
		repos = providedFlags.repository
	} else {
		repos, err = getTerraformModulesRepositoriesNameList(gh)
		if err != nil {
			return err
		}
	}
	sort.Strings(repos)

	agent, err := ssh.NewSSHAgentAuth("git")
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := fmt.Sprintf("%s/%s", home, repositoriesCacheDir)

	err = helper.CreateDirectory(dir)
	if err != nil {
		return err
	}

	if !providedFlags.skipConfirm {
		cli.Printf("You are about to bump %s %s to version '%s' in the following repositories:\n", bumpType, providedFlags.name, providedFlags.version)
		for _, r := range repos {
			cli.Printf("  - %s\n", r)
		}
		cli.Print("\n")
	}

	cli.SetSkipConfirmation(providedFlags.skipConfirm)
	if !cli.AskConfirmation("Do you want to confirm this action?") {
		return errAskConfirmationNotConfirmed
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(repos))
	results := make(chan BumpResult, len(repos))
	parallelism := make(chan int, providedFlags.parallelism)

	for _, repo := range repos {
		go func(repo string) {
			defer wg.Done()

			parallelism <- 1
			result := BumpResult{}
			result.Repository = repo

			branch := providedFlags.branch
			if branch == "" {
				branch = fmt.Sprintf("yak_terraform_%s_bump-%s-%d", bumpType, providedFlags.name, time.Now().UnixMilli())
			}

			repository, worktree, err := cloneRepositoryAndGetWorktree(gh, agent, repo, branch, dir)
			if err != nil {
				result.Error = err.Error()
				results <- result
				<-parallelism
				return
			}

			err = bumpFunc(path.Join(dir, repo))
			if err != nil {
				result.Error = err.Error()
				results <- result
				<-parallelism
				return
			}

			prNumber, err := commitChangesAndCreatePullRequest(repository, worktree, gh, repo, branch, dir, bumpType)
			if err != nil {
				result.Error = err.Error()
				results <- result
				<-parallelism
				return
			}

			if prNumber != -1 {
				result.URL = fmt.Sprintf("https://github.com/doctolib/%s/pull/%d\n", repo, prNumber)
			}
			results <- result
			<-parallelism
		}(repo)
	}

	wg.Wait()
	close(results)

	output := map[string]interface{}{}
	for result := range results {
		output[result.Repository] = result
	}
	_ = cli.PrintYAML(output)
	return nil
}
