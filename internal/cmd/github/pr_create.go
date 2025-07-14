package github

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	c "github.com/fatih/color"
	"github.com/nexidian/gocliselect"
	log "github.com/sirupsen/logrus"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"golang.org/x/crypto/ssh"

	"github.com/go-git/go-git/v5"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/google/go-github/v73/github"
	"github.com/spf13/cobra"
)

// flags used by the create pull request command
type tasksCreatePullRequestFlags struct {
	// draft is a flag to create a draft pull request
	draft bool
	// adhocScope is a flag to provide a scope for an adhoc pull request
	adhocScope string
	// adhocDescription is a flag to provide a description for an adhoc pull request
	adhocDescription string
	// prType is a flag to provide a type for a pull request
	prType string
	// breakingChanges is a flag to indicate if the pull request has breaking changes
	breakingChanges bool
	// silent is a flag to create a pull request silently
	silent bool
}

// PullRequestTitle represents the title of a GitHub Pull Request
type PullRequestTitle struct {
	Type            string // e.g. "feat", "fix", "refactor", etc.
	Scope           string // scope of the PR, e.g. "api", "ui", "backend", etc.
	BreakingChanges bool   // does the PR have breaking changes?
	Description     string // description of the PR
}

// PullRequestBody represents the body of a GitHub Pull Request
type PullRequestBody struct {
	// Description is the description of the PR
	Description string `json:"description"`
	// Context is the context of the PR (e.g. "Fixes #123")
	Context string `json:"context,omitempty"`
	// Dependencies is a boolean to indicate if the PR has dependencies
	Dependencies bool `json:"dependencies,omitempty"`
	// Tests is the output of the tests ran during the PR
	Tests string `json:"tests,omitempty"`
}

// PullRequest represents a GitHub Pull Request
type PullRequest struct {
	LocalBranch   string
	RepoName      string
	DefaultBranch string
	Draft         bool
	Title         PullRequestTitle
	Body          PullRequestBody

	// BaseRepo is the repository that the Pull Request is against
	BaseRepo *github.Repository
	// HeadRepo is the repository that the Pull Request is from
	HeadRepo *github.Repository
}

const (
	JiraBranchNamePattern = `^[A-Za-z]+-\d+/[A-Za-z0-9_-]+$`
)

var (
	providedPullRequestFlags tasksCreatePullRequestFlags
	browser                  cli.Browser

	createPullRequestCmd = &cobra.Command{
		Use:     "create",
		Aliases: []string{"cpr"},
		Short:   "Create pull request for current branch",
		RunE:    createPullRequest,
		Example: "yak pr create -a cluster-1 d 'Increases version to 1.0.0' -w\nyak pr create -b -s",
	}

	PRTypes = []string{"fix", "feat", "chore", "perf", "docs", "style", "refactor", "ci", "test"}
)

// isJiraTaskLocalBranch checks if the local branch is a Jira task branch
//
// Jira task branches have the following format: <JIRA_KEY>-<NUMBER>/<BRANCH_NAME>
//
// regex: ^[A-Za-z]+-\d+/[A-Za-z0-9_-]+$
func isJiraTaskLocalBranch(branchName string) bool {
	match, err := regexp.MatchString(JiraBranchNamePattern, branchName)
	if err != nil {
		log.Printf("Error in MatchString for JiraBranchNamePattern: %v", err)
		return false
	}
	return match
}

// ExtractRepoName extracts the repository name from a remote URL
//
// Supported formats:
// - git@github.com:user/repo.git
// - https://github.com/user/repo.git
//
// Returns an empty string if the URL is not in a supported format
func ExtractRepoName(remoteURL string) string {
	// git@github.com:user/repo.git
	if strings.HasPrefix(remoteURL, "git@") {
		parts := strings.Split(remoteURL, ":")
		if len(parts) > 1 {
			path := parts[1] // user/repo.git
			return strings.TrimSuffix(filepath.Base(path), ".git")
		}
	}

	// https://github.com/user/repo.git
	if strings.HasPrefix(remoteURL, "http://") || strings.HasPrefix(remoteURL, "https://") {
		parsedURL, err := url.Parse(remoteURL)
		if err != nil {
			log.Printf("Error parsing URL '%s': %v", remoteURL, err)
			return ""
		}
		return strings.TrimSuffix(filepath.Base(parsedURL.Path), ".git")
	}

	return ""
}

// extractBranchName extracts the branch name from a branch reference
//
// Branch reference format: <remote>/<branch_name>
//
// Example: origin/feature/JIRA-1234-my-awesome-branch
//
// Returns the branch name
func extractBranchName(branchRef string) string {
	// Split the branch reference by '/'
	parts := strings.Split(branchRef, "/")
	// Return the last part of the split (which is the branch name)
	return parts[len(parts)-1]
}

// generatePRTitleStr generates the title of the Pull Request
//
// The format of the title is:
// <type>(<scope>): <description>
//
// <type> is one of:
// - feat: A new feature
// - fix: A bug fix
// - docs: Documentation only changes
// - style: Changes that do not affect the meaning of the code
// - refactor: A code change that neither fixes a bug nor adds a feature
// - perf: A code change that improves performance
// - test: Adding missing or correcting existing tests
// - chore: Changes to the build process or auxiliary tools and libraries such as documentation generation
//
// <scope> is a brief description of the scope of the change
//
// <description> is a brief description of the change
//
// Example: feat(api): Add new endpoint to retrieve all repos
//
// If the PR has breaking changes, the title will have a "!" at the end
//
// Returns the generated title
func generatePRTitleStr(pr *PullRequest) string {
	titleStr := pr.Title.Type
	if pr.Title.Scope != "" {
		titleStr = titleStr + fmt.Sprintf("(%s)", pr.Title.Scope)
	}
	if pr.Title.BreakingChanges {
		titleStr = titleStr + "!"
	}
	titleStr = titleStr + ": " + pr.Title.Description
	return titleStr
}

func NewPullRequest(cmd *cobra.Command, flags *tasksCreatePullRequestFlags, adhoc bool) (*PullRequest, error) {
	yellow := c.New(c.FgYellow).SprintFunc()
	p := &PullRequest{LocalBranch: "",
		RepoName:      "",
		DefaultBranch: "",
		Draft:         flags.draft,
		Title: PullRequestTitle{
			Type:            "",
			Scope:           "",
			BreakingChanges: false,
			Description:     "",
		},
		Body: PullRequestBody{
			Description:  "",
			Context:      "",
			Dependencies: false,
			Tests:        "",
		},
	}

	// Get the current branch
	path, err := os.Getwd()
	if err != nil {
		log.Errorf("Error getting current working directory: %v", err)
		return nil, err
	}
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		log.Errorf("Error opening Git repository: %v", err)
		return nil, err
	}
	headRef, err := repo.Head()
	if err != nil {
		log.Errorf("Error getting HEAD reference: %v", err)
		return nil, err
	}
	if headRef.Name().IsBranch() {
		p.LocalBranch = headRef.Name().Short()
	} else {
		log.Errorf("HEAD is detached at %s\n", headRef.Hash())
		return nil, errors.New("detached HEAD")
	}

	if adhoc {
		if cmd.Flags().Changed("adhoc-scope") {
			p.Title.Scope = flags.adhocScope
		}
		if cmd.Flags().Changed("adhoc-description") {
			p.Title.Description = flags.adhocDescription
			p.Body.Description = flags.adhocDescription
		}
		p.Body.Context = ""
	} else {
		if !isJiraTaskLocalBranch(headRef.Name().Short()) {
			fmt.Printf("HEAD %s is not a Jira task branch. Please use adhoc mode\n", yellow(headRef.Name().Short()))
			return nil, errors.New("not a Jira task branch")
		}
		p.Title.Scope = strings.Split(headRef.Name().Short(), "/")[0]
		p.Title.Description = strings.ReplaceAll(strings.Split(headRef.Name().Short(), "/")[1], "_", " ")
		p.Body.Description = strings.ReplaceAll(strings.Split(headRef.Name().Short(), "/")[1], "_", " ")
		p.Body.Context = p.Title.Scope
	}

	origin, err := repo.Remote("origin")
	if err != nil {
		log.Errorf("Error getting remote 'origin': %v", err)
		return nil, err
	}

	// Get origin info, repo name and default branch
	p.RepoName = ExtractRepoName(origin.Config().URLs[0])

	sshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
	sshKey, err := os.ReadFile(sshKeyPath)

	if err != nil {
		log.Errorf("Error reading SSH key: %v", err)
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(sshKey)
	if err != nil {
		log.Errorf("Error parsing SSH key: %v", err)
		return nil, err
	}

	auth := &gitssh.PublicKeys{User: "git", Signer: signer}

	refList, err := origin.List(&git.ListOptions{Auth: auth})

	if err != nil {
		log.Errorf("Error listing references from 'origin': %v", err)
		return nil, err
	}

	for _, ref := range refList {
		if ref.Name() == "HEAD" {
			p.DefaultBranch = extractBranchName(ref.Target().String())
			break
		}
	}

	p.Title.Type = providedPullRequestFlags.prType
	p.Title.BreakingChanges = providedPullRequestFlags.breakingChanges
	return p, nil
}

// createPullRequest creates a new Pull Request on GitHub
func createPullRequest(cmd *cobra.Command, _ []string) error {
	// Ask user for PR type
	menu := gocliselect.NewMenu("Select the type of PR to create")
	for _, PRType := range PRTypes {
		menu.AddItem(PRType, PRType)
	}
	menu.AddItem("None", "0")
	providedPullRequestFlags.prType = menu.Display()

	// If user selected "None", exit
	if providedPullRequestFlags.prType == "0" {
		return nil
	}

	// If creating an adhoc PR, make sure user provided both scope and description
	if cmd.Flags().Changed("adhoc-scope") {
		if !cmd.Flags().Changed("adhoc-description") {
			log.Fatal("Using hadoc-scope requires Hadhoc description")
		}
	}

	// Create new PullRequest struct
	adhocPR := cmd.Flags().Changed("adhoc-scope") || cmd.Flags().Changed("adhoc-description")
	PullRequest, err := NewPullRequest(cmd, &providedPullRequestFlags, adhocPR)
	if err != nil {
		return err
	}

	// Create Pull Request on GitHub
	var pr *github.PullRequest
	remoteExists, _ := helper.CheckRemoteBranchExist(PullRequest.RepoName, PullRequest.LocalBranch)
	if remoteExists {
		pr, err = helper.CreatePullRequest(PullRequest.RepoName,
			PullRequest.LocalBranch,
			PullRequest.DefaultBranch,
			generatePRTitleStr(PullRequest),
			PullRequest.Body.Description,
			PullRequest.Body.Context, "\\<placeholder for dependencies\\>\n\n# Tests\n\\<placeholder for tests\\>", PullRequest.Draft)

		if err != nil {
			log.Fatalf("Error creating Pull Request: %v", err)
		}

		// If --silent flag is not provided, open a browser tab to the newly created PR
		if providedPullRequestFlags.silent {
			fmt.Printf("Pull request %d created\n", pr.GetNumber())
		} else {
			err = browser.Go(fmt.Sprintf("https://github.com/doctolib/%s/pull/%d", PullRequest.RepoName, pr.GetNumber()))
			if err != nil {
				log.Fatalf("Error opening browser: %v", err)
			}
		}
	} else {
		fmt.Printf("Remote branch %s does not exist\n", PullRequest.LocalBranch)
	}
	return nil
}

func init() {
	createPullRequestCmd.Flags().BoolVarP(&providedPullRequestFlags.draft, "wip", "w", false, "Create draft Pull Request")
	createPullRequestCmd.Flags().StringVarP(&providedPullRequestFlags.adhocScope, "adhoc-scope", "a", "", "Provide Hadhoc scope")
	createPullRequestCmd.Flags().StringVarP(&providedPullRequestFlags.adhocDescription, "adhoc-description", "d", "", "Provide Hadhoc description")
	createPullRequestCmd.Flags().BoolVarP(&providedPullRequestFlags.breakingChanges, "breaking-changes", "b", false, "Pull request has breaking changes")
	createPullRequestCmd.Flags().BoolVarP(&providedPullRequestFlags.silent, "silent", "s", false, "Create Pull Request silently")

	switch runtime.GOOS {
	case "darwin":
		browser = &cli.MacosBrowser{}
	case "windows":
		browser = &cli.WinBrowser{}
	case "linux":
		browser = &cli.LinuxBrowser{}
	default:
		log.Fatalf("Unsupported operating system")
	}
}
