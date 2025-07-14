package helper

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-ini/ini"
	log "github.com/sirupsen/logrus"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/constant"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/google/go-github/v73/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/vault/api"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
)

type Secret struct {
	Path      string
	Version   int
	Data      map[string]interface{}
	Metadata  map[string]interface{}
	Deleted   bool
	Destroyed bool
}

type VaultConfig struct {
	Endpoints      []string
	AwsProfile     string
	AwsRegion      string
	VaultRole      string
	SecretPrefix   string
	VaultNamespace string
}

type SecretExpectedError struct {
	Path     string
	Version  int
	Name     string
	Error    error
	Contains string
}

var (
	errPlatformNotFound                         = errors.New("platform does not exist in configuration file")
	errCantReadAwsProfile                       = errors.New("can't read awsProfile")
	errCantReadAwsRegion                        = errors.New("can't read awsRegion")
	errCantReadVaultRole                        = errors.New("can't read vaultRole")
	errCantReadClusterEndpoint                  = errors.New("can't read endpoint for cluster from config")
	errCantReadCluster                          = errors.New("can't read provided cluster")
	errEnvironmentNotFound                      = errors.New("environment does not exist in configuration file")
	ErrEnvironmentCantBeSetWithoutPlatform      = errors.New("environment can't be set without platform being set")
	errEnvironmentCantBeSetWhenPlatformIsCommon = errors.New("environment can't be set when platform is common")
	ErrSecretNotFound                           = errors.New("secret does not exist")
	ErrSecretPathNotFound                       = errors.New("secret path does not exist")
	ErrSecretDataKeyNotFound                    = errors.New("secret data key not found")
	ErrGithubTokenNotFound                      = errors.New("no GITHUB_TOKEN or HOMEBREW_GITHUB_API_TOKEN found in environment")
	ErrListSecretNotSecretPath                  = errors.New("can't list a secret, you must provide a path")
)

func GetVaultClient(vaultAddr string) *api.Client {
	client, err := api.NewClient(&api.Config{
		Address: vaultAddr,
	})

	client.AddHeader("X-Yak-Client", "true")
	client.AddHeader("X-Yak-Client-Version", constant.Version)
	client.AddHeader("X-Yak-Client-Control", constant.VersionControl)

	if err != nil {
		panic(err)
	}

	return client
}

var VaultLoginWithAwsAndGetClients = internalVaultLoginWithAwsAndGetClients

func internalVaultLoginWithAwsAndGetClients(config *VaultConfig) ([]*api.Client, error) {
	var clients []*api.Client
	for _, endpoint := range config.Endpoints {
		client := GetVaultClient(endpoint)
		client.SetNamespace(config.VaultNamespace)
		secretToken, err := VaultLoginWithAws(client, config.AwsProfile, config.AwsRegion, config.VaultRole)
		if err != nil {
			return nil, err
		}
		client.SetToken(secretToken.ClientToken)

		clients = append(clients, client)
	}

	return clients, nil
}

func VaultLoginWithAws(vaultClient *api.Client, awsProfile string, awsRegion string, vaultRole string) (*api.SecretAuth, error) {
	sess := GetAwsSession(awsProfile, awsRegion)
	data, err := generateLoginData(sess)
	if err != nil {
		return nil, err
	}

	data["role"] = vaultRole
	path := fmt.Sprintf("auth/aws-%s/login", awsRegion)
	secret, err := vaultClient.Logical().Write(path, data)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		return nil, fmt.Errorf("empty response from credential provider")
	}

	return secret.Auth, nil
}

func generateLoginData(awsSession *session.Session) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	var params *sts.GetCallerIdentityInput
	svc := sts.New(awsSession)
	stsRequest, _ := svc.GetCallerIdentityRequest(params)

	err := stsRequest.Sign()
	if err != nil {
		return nil, err
	}

	// extract out the relevant parts of the request
	headersJSON, err := json.Marshal(stsRequest.HTTPRequest.Header)
	if err != nil {
		return nil, err
	}
	requestBody, err := io.ReadAll(stsRequest.HTTPRequest.Body)
	if err != nil {
		return nil, err
	}

	data["iam_http_request_method"] = stsRequest.HTTPRequest.Method
	data["iam_request_url"] = base64.StdEncoding.EncodeToString([]byte(stsRequest.HTTPRequest.URL.String()))
	data["iam_request_headers"] = base64.StdEncoding.EncodeToString(headersJSON)
	data["iam_request_body"] = base64.StdEncoding.EncodeToString(requestBody)

	return data, nil
}

// TODO: Use a more elegant solution to get credentials from SSO
type awsCredentials struct {
	Version         int       `json:"Version"`
	AccessKeyID     string    `json:"AccessKeyId"`
	SecretAccessKey string    `json:"SecretAccessKey"`
	SessionToken    string    `json:"SessionToken"`
	Expires         time.Time `json:"Expiration"`
}

// TODO: Use a more elegant solution to get credentials from SSO
func getAwsCredentials(profile string) awsCredentials {
	cmd := exec.Command("aws-vault", "exec", profile, "--json")
	os.Unsetenv("AWS_VAULT") // avoid nested session error

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()

	if err != nil {
		cli.Println(stderr.String())
		panic(err)
	}

	credentials := awsCredentials{}
	err = json.Unmarshal(output, &credentials)
	if err != nil {
		panic(err)
	}

	return credentials
}

func stsSigningResolver(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
	defaultEndpoint, err := endpoints.DefaultResolver().EndpointFor(service, region, optFns...)
	if err != nil {
		return defaultEndpoint, err
	}
	defaultEndpoint.SigningRegion = region

	return defaultEndpoint, nil
}

func GetAwsSession(profile string, region string) *session.Session {
	var sess *session.Session

	if profile != "" {
		// For CI, rely on instance role and assume role configured in profile
		if _, exists := os.LookupEnv("GITHUB_ACTIONS"); exists {
			sess = session.Must(session.NewSessionWithOptions(
				session.Options{
					Profile: profile,
					Config: aws.Config{
						Region:              aws.String(region),
						STSRegionalEndpoint: endpoints.RegionalSTSEndpoint,
						EndpointResolver:    endpoints.ResolverFunc(stsSigningResolver),
					},
					SharedConfigState: session.SharedConfigEnable,
				},
			))
		} else { // If executed from a laptop, rely on aws-vault and local aws config
			creds := getAwsCredentials(profile)
			sess = session.Must(session.NewSessionWithOptions(
				session.Options{
					Config: aws.Config{
						Region:           aws.String(region),
						Credentials:      credentials.NewStaticCredentials(creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken),
						EndpointResolver: endpoints.ResolverFunc(stsSigningResolver),
					},
				},
			))
		}
	} else {
		sess = session.Must(session.NewSession(&aws.Config{
			Region:           aws.String(region),
			EndpointResolver: endpoints.ResolverFunc(stsSigningResolver),
		}))
	}

	return sess
}

func InitViper(c string) {
	viper.Reset()
	viper.SetConfigType("yaml")
	r := strings.NewReader(c)
	err := viper.ReadConfig(r)
	if err != nil {
		panic(err)
	}
}

func CreateDirectory(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func WriteFile(content string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte(content))
	if err != nil {
		return err
	}
	file.Close()
	return nil
}

func ExecuteCobraCommand(cmd *cobra.Command, args []string, input ...io.Reader) (string, string, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cli.SetOut(stdout)
	cli.SetErr(stderr)
	if len(input) > 0 {
		cli.SetIn(input[0])
		cli.SetPasswordReader(&cli.IoReaderPasswordReader{Reader: input[0]})
	}
	cli.SetCobraCmdOutErr(cmd)
	cmd.SilenceErrors = true
	cmd.SetArgs(args)
	err := cmd.Execute()

	return stdout.String(), stderr.String(), err
}

func getFeatureTeamSSOProfile(team string) string {
	return fmt.Sprintf("shared-%s-sso", team)
}

func getFeatureTeamSSORoleName(team string) string {
	return fmt.Sprintf("FT-%s", team)
}

func getFeatureTeamVaultRole(team string) string {
	return fmt.Sprintf("ft-%s", team)
}

func GetVaultConfig(platform string, environment string, team ...string) (*VaultConfig, error) {
	var err error
	vaultConfig := VaultConfig{}

	if platform == "" && environment != "" {
		err = ErrEnvironmentCantBeSetWithoutPlatform
	} else if platform == "common" && environment != "" && environment != "terraform" {
		err = errEnvironmentCantBeSetWhenPlatformIsCommon
	} else {
		if platform == "" {
			platform = "common"
		}
		config := viper.GetStringMap("platforms." + platform)
		if len(config) == 0 {
			err = errPlatformNotFound
		} else {
			if len(team) > 0 && team[0] != "" {
				vaultConfig.AwsProfile = getFeatureTeamSSOProfile(team[0])
				vaultConfig.VaultRole = getFeatureTeamVaultRole(team[0])
			} else {
				if val, ok := config["awsprofile"].(string); ok {
					vaultConfig.AwsProfile = val
				} else {
					err = errCantReadAwsProfile
				}
				if val, ok := config["vaultrole"].(string); ok {
					vaultConfig.VaultRole = val
				} else {
					err = errCantReadVaultRole
				}
			}
			if val, ok := config["awsregion"].(string); ok {
				vaultConfig.AwsRegion = val
			} else {
				err = errCantReadAwsRegion
			}
			if clusters, ok := config["clusters"].([]interface{}); ok && len(clusters) != 0 {
				for _, cluster := range clusters {
					if val := viper.GetString("clusters." + fmt.Sprint(cluster) + ".endpoint"); len(val) != 0 {
						vaultConfig.Endpoints = append(vaultConfig.Endpoints, val)
					} else {
						return nil, errCantReadClusterEndpoint
					}
				}
			} else {
				err = errCantReadCluster
			}

			parentNamespace := "doctolib"
			if val, ok := config["vaultparentnamespace"].(string); ok {
				parentNamespace = val
			}

			vaultConfig.VaultNamespace = parentNamespace + "/" + platform
			if environment == "common" || environment == "" {
				vaultConfig.SecretPrefix = "common"
			} else {
				if config["environments"] != nil {
					if _, ok := config["environments"].(map[string]interface{})[environment]; ok {
						vaultConfig.SecretPrefix = config["environments"].(map[string]interface{})[environment].(string)
					} else {
						err = errEnvironmentNotFound
					}
				} else {
					err = errEnvironmentNotFound
				}
			}
		}
	}

	if err != nil {
		return nil, err
	}
	return &vaultConfig, err
}

type Annotation struct {
	Path string
	Line int
}

// Find a specific string in files under the given path
// Return a list with the files and line numbers and an error.
func FindStringInPathAndGetLineNumber(checkPath string, find string, maxDepth int) ([]Annotation, error) {
	var a []Annotation
	err := filepath.WalkDir(checkPath,
		func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if maxDepth != -1 {
				if strings.Count(path, string(os.PathSeparator))-strings.Count(checkPath, string(os.PathSeparator)) > maxDepth {
					return fs.SkipDir
				}
			}
			if !info.IsDir() &&
				strings.Contains(path, ".tf") &&
				!strings.Contains(path, ".terraform/") {
				line, err := FindStringInFileAndGetLineNumber(path, find)
				if err != nil {
					return err
				}
				if line != -1 {
					a = append(a, Annotation{Path: path, Line: line})
				}
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// Find a specific string in the given file
// Return the line number where it has been found and an error.
func FindStringInFileAndGetLineNumber(file string, find string) (int, error) {
	f, err := os.Open(file)
	if err != nil {
		return -1, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	line := 1

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), find) {
			return line, nil
		}
		line++
	}

	return -1, nil
}

// Send email using Amazon SES
func SendEmail(to string, subject string, body string) error {
	profile := viper.GetString("email.awsProfile")
	if profile == "" {
		return errCantReadAwsProfile
	}

	region := viper.GetString("email.awsRegion")
	if region == "" {
		return errCantReadAwsRegion
	}

	sess := GetAwsSession(profile, region)
	svc := ses.New(sess)
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(to),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(body),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(constant.EmailDefaultSender),
	}

	_, err := svc.SendEmail(input)
	if err != nil {
		return err
	}
	return nil
}

func GetGithubClient() (*github.Client, error) {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		githubToken = os.Getenv("HOMEBREW_GITHUB_API_TOKEN")
	}

	if githubToken == "" {
		return nil, ErrGithubTokenNotFound
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})
	tc := oauth2.NewClient(context.Background(), ts)

	return github.NewClient(tc), nil
}

func CloneGitRepository(repo string, dir string) (*git.Repository, error) {
	return cloneGitRepository(repo, dir, 0)
}

func CloneGitRepositoryFast(repo string, dir string) (*git.Repository, error) {
	return cloneGitRepository(repo, dir, 1)
}

func cloneGitRepository(repo string, dir string, depth int) (*git.Repository, error) {
	agent, err := ssh.NewSSHAgentAuth("git")
	if err != nil {
		return nil, err
	}

	err = CreateDirectory(path.Join(dir, repo))
	if err != nil {
		return nil, err
	}

	var r *git.Repository
	opts := &git.CloneOptions{
		URL:  fmt.Sprintf("git@github.com:doctolib/%s.git", repo),
		Auth: agent,
	}
	if depth > 0 {
		opts.Depth = depth
		opts.SingleBranch = true
	}

	r, err = git.PlainOpen(path.Join(dir, repo))
	if err == git.ErrRepositoryNotExists {
		r, err = git.PlainClone(path.Join(dir, repo), false, opts)

		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return r, nil
}

func CheckoutGitBranch(repo *git.Repository, branch string, create, keepLocalIndex bool) (*git.Worktree, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Create: create,
		Force:  false,
		Keep:   keepLocalIndex, // default: false
		Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
	})

	if err != nil {
		return nil, err
	}

	return worktree, nil
}

func PullGitBranch(worktree *git.Worktree) error {
	err := worktree.Pull(&git.PullOptions{})
	if err != nil {
		if err != git.NoErrAlreadyUpToDate {
			return err
		}
	}
	return nil
}

func CloneRepositoryAndGetWorktree(repo string, branchToCreate string, dir string, mainBranch string, depth int) (*git.Repository, *git.Worktree, error) {
	var err error

	r, err := cloneGitRepository(repo, dir, depth)
	if err != nil {
		return nil, nil, err
	}

	wt, err := CheckoutGitBranch(r, mainBranch, false, false)
	if err != nil {
		return nil, nil, err
	}

	err = PullGitBranch(wt)
	if err != nil {
		return nil, nil, err
	}

	wt, err = CheckoutGitBranch(r, branchToCreate, true, false)
	if err != nil {
		wt, err = CheckoutGitBranch(r, branchToCreate, false, false)

		if err != nil {
			return r, nil, err
		}
	}
	return r, wt, nil
}

func HasStagedGitChanges(repo *git.Repository) (bool, error) {
	w, err := repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	// Iterate through all file statuses.  A staged file will have a
	// non-zero Staging value *and* won't be Untracked.
	for _, fileStatus := range status {
		if fileStatus.Staging != git.Unmodified && fileStatus.Worktree != git.Untracked {
			return true, nil // Found a staged change
		}
	}

	return false, nil // No staged changes found
}

func CreateGitCommit(repo *git.Repository, message string) error {
	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	_, err = w.Status()
	if err != nil {
		return err
	}

	// Unfortunately, go-git doesn't support commit signing in a GPG/SSH-agnostic way without making things
	// too complex, so we use the `git` command-line tool to leverage .gitconfig for signing
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = w.Filesystem.Root() // Set the working directory to the repository root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed: %w\nOutput: %s", err, output)
	}
	return nil
}

func CommitAll(repository *git.Repository, worktree *git.Worktree, message string) error {
	status, err := worktree.Status()
	if err != nil {
		return err
	}

	if status.IsClean() {
		return nil
	}

	err = worktree.AddGlob("*")
	if err != nil {
		return err
	}

	_, err = worktree.Commit(message, &git.CommitOptions{All: true})
	if err != nil {
		return err
	}

	err = repository.Push(&git.PushOptions{})
	if err != nil {
		return err
	}
	return nil
}

func CheckRemoteBranchExist(repo string, branch string) (bool, error) {
	gh, err := GetGithubClient()
	if err != nil {
		return false, err
	}

	_, resp, err := gh.Repositories.GetBranch(context.Background(), "doctolib", repo, branch, 0)

	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, err
		} else {
			log.Fatal(err)
		}
	} else {
		return true, nil
	}
	return false, nil
}

func ListPullRequests(repo string) ([]*github.PullRequest, error) {
	gh, err := GetGithubClient()
	if err != nil {
		return nil, err
	}
	var prsList []*github.PullRequest
	user, _, err := gh.Users.Get(context.Background(), "")
	if err != nil {
		return nil, err
	}

	prs, _, err := gh.PullRequests.List(context.Background(), "doctolib", repo, &github.PullRequestListOptions{
		State:     "open",
		Sort:      "created",
		Direction: "desc",
	})

	if err != nil {
		return nil, err
	}

	for _, pr := range prs {
		if pr.GetUser().GetLogin() == user.GetLogin() {
			prsList = append(prsList, pr)
		}
	}
	return prsList, nil
}

func CreatePullRequest(repo string, branchFrom string, branchTo string, title string, descriptionStr string, contextStr string, dependenciesStr string, wip bool) (*github.PullRequest, error) {
	gh, err := GetGithubClient()
	if err != nil {
		return nil, err
	}

	newPR := &github.NewPullRequest{
		Title: github.Ptr(title),
		Head:  github.Ptr(branchFrom),
		Base:  github.Ptr(branchTo),
		Body: github.Ptr(
			strings.ReplaceAll("# Description\n"+descriptionStr+"\n\n# Context\n"+contextStr+"\n\n# Dependencies\nno\n", "\n", "\r\n")),
		MaintainerCanModify: github.Ptr(true),
		Draft:               github.Ptr(false),
	}

	if dependenciesStr != "" {
		newPR.Body = github.Ptr("# Description\n" + descriptionStr + "\n\n# Context\n" + contextStr + "\n\n# Dependencies\n" + dependenciesStr + "\n")
	}

	if wip {
		newPR.Draft = github.Ptr(true)
	}

	pr, _, err := gh.PullRequests.Create(context.Background(), "doctolib", repo, newPR)
	if err != nil {
		return nil, err
	}
	return pr, err
}

func CreateComment(repo string, pr *github.PullRequest, comment string) error {
	gh, err := GetGithubClient()
	if err != nil {
		return err
	}

	// Works for PRs, a pull request is also an issue
	_, _, err = gh.Issues.CreateComment(context.Background(), "doctolib", repo, pr.GetNumber(),
		&github.IssueComment{
			Body: github.Ptr(comment),
		})

	if err != nil {
		return err
	}

	return nil
}

func ListContainsString(list []string, element string) bool {
	for _, v := range list {
		if v == element {
			return true
		}
	}
	return false
}

func SerializeVaultDataMapToString(data map[string]interface{}) map[string]string {
	// Since our "data" map can only contain string values, we
	// will take strings from Data and write them in as-is,
	// and write everything else in as a JSON serialization of
	// whatever value we get so that complex types can be
	// passed around and processed elsewhere if desired.
	// Note: This is a different map to jsonData, as this can only
	// contain strings
	dataMap := map[string]string{}
	for k, v := range data {
		if vs, ok := v.(string); ok {
			dataMap[k] = vs
		} else {
			// Again ignoring error because we know this value
			// came from JSON in the first place and so must be valid.
			vBytes, _ := json.Marshal(v)
			dataMap[k] = string(vBytes)
		}
	}
	return dataMap
}

func WalkVaultPath(clients []*api.Client, path string) ([]string, error) {
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	result := []string{}

	secrets, err := clients[0].Logical().List("kv/metadata/" + path)
	if err != nil {
		return result, err
	}

	if secrets == nil {
		return result, nil
	}

	for _, s := range secrets.Data["keys"].([]interface{}) {
		r := []string{}
		if strings.HasSuffix(s.(string), "/") {
			r, err = WalkVaultPath(clients, path+s.(string))
			if err != nil {
				return result, err
			}
		} else {
			if strings.HasPrefix(path+s.(string), "/") {
				r = append(r, strings.Replace(path+s.(string), "/", "", 1))
			} else {
				r = append(r, path+s.(string))
			}
		}

		result = append(result, r...)
	}

	return result, nil
}

func getAwsConfigPath() string {
	if os.Getenv("AWS_CONFIG_FILE") != "" {
		return os.Getenv("AWS_CONFIG_FILE")
	}
	return os.Getenv("HOME") + "/.aws/config"
}

func LoadAWSConfig() (*ini.File, error) {
	cfg, err := ini.Load(getAwsConfigPath())
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func SaveIniFile(config *ini.File, configFilePath string) error {
	return config.SaveTo(configFilePath)
}

func overrideFeatureTeamAWSConfigProfile(config *ini.File, team string, awsProfileName string) error {
	roleName := getFeatureTeamSSORoleName(team)
	ssoProfileName := getFeatureTeamSSOProfile(team)
	mSection, err := config.GetSection("profile " + ssoProfileName)
	if err != nil {
		mSection, _ = config.NewSection("profile " + ssoProfileName)
	}
	for key := range mSection.KeysHash() {
		mSection.DeleteKey(key)
	}

	_, err = mSection.NewKey("region", constant.SSORegion)
	if err != nil {
		return err
	}
	_, err = mSection.NewKey("output", "json")
	if err != nil {
		return err
	}
	_, err = mSection.NewKey("sso_start_url", constant.SSOStartURL)
	if err != nil {
		return err
	}
	_, err = mSection.NewKey("sso_region", constant.SSORegion)
	if err != nil {
		return err
	}
	_, err = mSection.NewKey("sso_account_id", constant.AccountNo)
	if err != nil {
		return err
	}
	_, err = mSection.NewKey("sso_role_name", roleName)
	if err != nil {
		return err
	}
	section, err := config.GetSection("profile " + awsProfileName)
	if err != nil {
		section, _ = config.NewSection("profile " + awsProfileName)
	}
	for key := range section.KeysHash() {
		section.DeleteKey(key)
	}
	cmd := []string{"aws-vault", "exec", ssoProfileName, "--json"}
	_, err = section.NewKey("credential_process", strings.Join(cmd, " "))
	if err != nil {
		return err
	}
	_, err = section.NewKey("region", constant.SSORegion)
	if err != nil {
		return err
	}
	return nil
}

func createFileIfDoesnotExist(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0700)
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		newFile, err := os.Create(path)
		if err != nil {
			return err
		}
		newFile.Close()
		return nil
	}
	return nil
}

func AddAWSConfigProfileForFeatureTeam(team string) (string, error) {
	awsProfileName := fmt.Sprintf("shared-%s", team)
	awsConfigFilePath := os.Getenv("HOME") + "/.aws/config." + team

	// overload AWS_CONFIG_FILE in yak with a dedicated aws config file
	os.Setenv("AWS_CONFIG_FILE", awsConfigFilePath)

	err := createFileIfDoesnotExist(awsConfigFilePath)
	if err != nil {
		log.Errorf("Error creating AWS config file: %v", err)
		return "", err
	}
	// load aws config
	config, err := LoadAWSConfig()
	if err != nil {
		log.Errorf("Error loading AWS config file: %v", err)
		return "", err
	}

	// add the feature team profile in the aws config file
	err = overrideFeatureTeamAWSConfigProfile(config, team, awsProfileName)
	if err != nil {
		log.Errorf("Error overriding AWS config file for feature team: %v", err)
		return "", err
	}

	// save config back to the aws config file provided as a parameter
	err = SaveIniFile(config, getAwsConfigPath())
	if err != nil {
		log.Errorf("Error saving AWS config file: %v", err)
		return "", err
	}
	return awsConfigFilePath, nil
}

func HasAnyPrefix(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func GetEnvOrDefault(envKey, defaultValue string) string {
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	return defaultValue
}

func GetAge(t time.Time) string {
	age := time.Since(t)
	days := int(age.Hours() / 24)
	hours := int(age.Hours()) % 24
	minutes := int(age.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd", days)
	} else if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	} else {
		return fmt.Sprintf("%ds", int(age.Seconds()))
	}
}
