package helper

import (
	"os"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type getVaultConfigTest struct {
	initial     string
	platform    string
	environment string
	expected    interface{}
}

func TestGetVaultConfig(t *testing.T) {
	var testScenarios = map[string]getVaultConfigTest{
		// Everything is fine ; one endpoint defined
		"nominal": {
			initial: `clusters:
  shared:
    endpoint: "https://vault.local:8200"
platforms:
  dev:
    clusters: [shared]
    awsProfile: dev-sso
    awsRegion: eu-central-1
    vaultRole: administrator
    environments:
      de: dev-aws-de-fra-1`,
			platform:    "dev",
			environment: "de",
			expected: &VaultConfig{
				AwsProfile:     "dev-sso",
				AwsRegion:      "eu-central-1",
				VaultRole:      "administrator",
				SecretPrefix:   "dev-aws-de-fra-1",
				VaultNamespace: "doctolib/dev",
				Endpoints:      []string{"https://vault.local:8200"},
			},
		},
		// Everything is fine ; two endpoints defined
		"nominal_multiple_endpoints": {
			initial: `clusters:
  shared:
    endpoint: "https://vault.local:8200"
  production:
    endpoint: "https://vault-2.local:8200"
platforms:
  common:
    clusters: [production, shared]
    awsProfile: dev-sso
    awsRegion: eu-central-1
    vaultRole: administrator`,
			platform: "common",
			expected: &VaultConfig{
				AwsProfile:     "dev-sso",
				AwsRegion:      "eu-central-1",
				VaultRole:      "administrator",
				SecretPrefix:   "common",
				VaultNamespace: "doctolib/common",
				Endpoints:      []string{"https://vault-2.local:8200", "https://vault.local:8200"},
			},
		},
		"nominal_parent_namespace_defined": {
			initial: `clusters:
  shared:
    endpoint: "https://vault.local:8200"
platforms:
  dev:
    clusters: [shared]
    awsProfile: dev-sso
    awsRegion: eu-central-1
    vaultRole: administrator
    vaultParentNamespace: uhdp
    environments:
      de: dev-aws-de-fra-1`,
			platform:    "dev",
			environment: "de",
			expected: &VaultConfig{
				AwsProfile:     "dev-sso",
				AwsRegion:      "eu-central-1",
				VaultRole:      "administrator",
				SecretPrefix:   "dev-aws-de-fra-1",
				VaultNamespace: "uhdp/dev",
				Endpoints:      []string{"https://vault.local:8200"},
			},
		},
		// environment is set but not platform
		"environment_without_platform": {
			initial: `clusters:
  shared:
    endpoint: "https://vault.local:8200"
platforms:
  dev:
    clusters: [shared]
    awsProfile: dev-sso
    awsRegion: eu-central-1
    vaultRole: administrator
    environments:
      de: dev-aws-de-fra-1`,
			platform:    "",
			environment: "de",
			expected:    SecretExpectedError{Error: ErrEnvironmentCantBeSetWithoutPlatform, Name: "errEnvironmentCantBeSetWithoutPlatform"},
		},
		// environment is set and platform is explicitely set to common
		"environment_with_platform_common": {
			initial: `clusters:
  shared:
    endpoint: "https://vault.local:8200"
platforms:
  common:
    clusters: [shared]
    awsProfile: dev-sso
    awsRegion: eu-central-1
    vaultRole: administrator`,
			platform:    "common",
			environment: "de",
			expected:    SecretExpectedError{Error: errEnvironmentCantBeSetWhenPlatformIsCommon, Name: "errEnvironmentCantBeSetWhenPlatformIsCommon"},
		},
		// platform set is invalid
		"environment_with_invalid_platform": {
			initial: `clusters:
  shared:
    endpoint: "https://vault.local:8200"
platforms:
  dev:
    clusters: [shared]
    awsProfile: dev-sso
    awsRegion: eu-central-1
    vaultRole: administrator
    environments:
      de: dev-aws-de-fra-1`,
			platform:    "invalid",
			environment: "de",
			expected:    SecretExpectedError{Error: errPlatformNotFound, Name: "errPlatformNotFound"},
		},
		// awsProfile is not defined in config for the provided platform
		"cant_read_aws_profile": {
			initial: `clusters:
  shared:
    endpoint: "https://vault.local:8200"
platforms:
  dev:
    clusters: [shared]
    awsRegion: eu-central-1
    vaultRole: administrator
    environments:
      de: dev-aws-de-fra-1`,
			platform:    "dev",
			environment: "de",
			expected:    SecretExpectedError{Error: errCantReadAwsProfile, Name: "errCantReadAwsProfile"},
		},
		// awsRegion is not defined in config for the provided platform
		"cant_read_aws_region": {
			initial: `clusters:
  shared:
    endpoint: "https://vault.local:8200"
platforms:
  dev:
    clusters: [shared]
    awsProfile: dev-sso
    vaultRole: administrator
    environments:
      de: dev-aws-de-fra-1`,
			platform:    "dev",
			environment: "de",
			expected:    SecretExpectedError{Error: errCantReadAwsRegion, Name: "errCantReadAwsRegion"},
		},
		// vaultRole is not defined in config for the provided platform
		"cant_read_vault_role": {
			initial: `clusters:
  shared:
    endpoint: "https://vault.local:8200"
platforms:
  dev:
    clusters: [shared]
    awsProfile: dev-sso
    awsRegion: eu-central-1
    environments:
      de: dev-aws-de-fra-1`,
			platform:    "dev",
			environment: "de",
			expected:    SecretExpectedError{Error: errCantReadVaultRole, Name: "errCantReadVaultRole"},
		},
		// environment set is not defined in config for the provided platform
		"cant_read_specific_environment": {
			initial: `clusters:
  shared:
    endpoint: "https://vault.local:8200"
platforms:
  dev:
    clusters: [shared]
    awsProfile: dev-sso
    awsRegion: eu-central-1
    vaultRole: administrator
    environments:
      fr: dev-aws-fr-par-1`,
			platform:    "dev",
			environment: "de",
			expected:    SecretExpectedError{Error: errEnvironmentNotFound, Name: "errEnvironmentNotFound"},
		},
		// environments attribute is missing in config for the provided platform
		"cant_read_environments": {
			initial: `clusters:
  shared:
    endpoint: "https://vault.local:8200"
platforms:
  dev:
    clusters: [shared]
    awsProfile: dev-sso
    awsRegion: eu-central-1
    vaultRole: administrator`,
			platform:    "dev",
			environment: "de",
			expected:    SecretExpectedError{Error: errEnvironmentNotFound, Name: "errEnvironmentNotFound"},
		},
		// clusters attribute is empty in config for the provided platform
		"cant_read_endpoint_list_empty": {
			initial: `clusters:
  shared:
    endpoint: "https://vault.local:8200"
platforms:
  dev:
    clusters: []
    awsProfile: dev-sso
    awsRegion: eu-central-1
    vaultRole: administrator`,
			platform: "dev",
			expected: SecretExpectedError{Error: errCantReadCluster, Name: "errCantReadCluster"},
		},
		// clusters attribute is missing in config for the provided platform
		"cant_read_endpoint_list_non_present": {
			initial: `clusters:
  shared:
    endpoint: "https://vault.local:8200"
platforms:
  dev:
    awsProfile: dev-sso
    awsRegion: eu-central-1
    vaultRole: administrator`,
			platform: "dev",
			expected: SecretExpectedError{Error: errCantReadCluster, Name: "errCantReadCluster"},
		},
		// endpoint value is missing in config for the provided cluster
		"cant_read_endpoint": {
			initial: `platforms:
  dev:
    clusters: [shared]
    awsProfile: dev-sso
    awsRegion: eu-central-1
    vaultRole: administrator`,
			platform: "dev",
			expected: SecretExpectedError{Error: errCantReadClusterEndpoint, Name: "errCantReadClusterEndpoint"},
		},
	}
	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			InitViper(v.initial)

			config, err := GetVaultConfig(v.platform, v.environment)

			if expectedConfig, ok := v.expected.(*VaultConfig); ok {
				// if we expect a secret to be returned
				assert.NoError(t, err, "no error should be returned")
				assert.EqualValues(t, expectedConfig, config)
			} else if expectedError, ok := v.expected.(SecretExpectedError); ok {
				// if we expect an error to be returned
				assert.ErrorIsf(t, err, expectedError.Error, "error should be %s", expectedError.Name)
				assert.Nil(t, config)
			} else {
				// if something else is configured in the test
				t.Fatal("expected value is of unsupported type")
			}
		})
	}
}

type findStringInFileAndGetLineNumber struct {
	Content       string
	ErrorExpected bool
	LineExpected  int
	Search        string
}

func TestFindStringInFileAndGetLineNumber(t *testing.T) {
	var testScenarios = map[string]findStringInFileAndGetLineNumber{
		"found":          {Content: "foo\nbar\nbaz", ErrorExpected: false, LineExpected: 3, Search: "baz"},
		"not_found":      {Content: "foo\nbar\nbaz", ErrorExpected: false, LineExpected: -1, Search: "nothing"},
		"cant_open_file": {Content: "", ErrorExpected: true, LineExpected: -1, Search: "nothing"},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			file, err := os.CreateTemp("", "TestFindStringInFileAndGetLineNumber")
			if err != nil {
				t.Fatal(err)
			}

			_, err = file.Write([]byte(v.Content))
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(file.Name())

			if v.ErrorExpected {
				os.Remove(file.Name())
			}

			result, err := FindStringInFileAndGetLineNumber(file.Name(), v.Search)

			if v.ErrorExpected {
				assert.Error(t, err, "there should be an error")
			} else {
				assert.NoError(t, err, "error should be nil")
			}

			assert.Equalf(t, result, v.LineExpected, "result should be equals to %d", v.LineExpected)
		})
	}
}

type initKubeClusterConfig struct {
	file                    *testFile
	kubeconfigEnvSet        bool
	correctKubeconfigEnv    bool
	expectedError           bool
	multiplePathsKubeConfig bool
}

type testFile struct {
	path    string
	content string
}

var sampleKubeConfig string = `apiVersion: v1
clusters:
- cluster:
    server: https://CA05F37FDEAFEC44E0567EF1632622AA.gr7.eu-central-1.eks.amazonaws.com
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM1ekNDQWMrZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeU1EUXdOekUxTkRJME5Gb1hEVE15TURRd05ERTFOREkwTkZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTkR6ClB1VmFUY3lQTUJOVGhwMS8xSTh1SGIxQ0VaeDBWd0t6T3BUbWZ6OEl2Q2JML2xQNnE4YWsxRWFQRzlyaWRMZjAKR21zaUtBOUFocVhTL2xwUG5WcjhXWUE2V3FUSGVhVXZRUHpqd0lPSVo2YlhiUSthSFk3aEJKdVhYRGEyWHRtZQpDR2gwTVNkU2JZNHUyQ1lUMXQxOU0xamJmSFd5SXpVWGVCVy8wckZ4NFJ6eW9WWVJOTnNtYUp0dVd6T0RpWlViCkhTT0luK25zaWN1UUd2NExOQnVpbFNJUHFiNENLdDFZL254WXNncnUyblplSEhHTWptaU5ualplWWRKa3hDZloKN21mNTBOSGE0Y2ZHMC9nenVEUk5mWTIrYWI1QlZzazE5VCs0VjFzbGltT1hFWW1Xa0t6ZHdaVjJCQVdzZkVVWQpyanB5dFBBSUlkTzY2MGtHWEQ4Q0F3RUFBYU5DTUVBd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZPSU85dW4vMEM0a0RseVp1bUIzSnFyaDhQZ2ZNQTBHQ1NxR1NJYjMKRFFFQkN3VUFBNElCQVFBT2lCaVU2WjBhVUV6RWZ5M0ZPdzB0N1B2TEpHUnVhL2FCUUc2KzdIanlVZk1JRVNqNApuT3pQWjRwUk5KeDZVL2dsbTlsTlA1SldXSVRJWi93TzAybzNiV1h1N081UFovdSt1MzB1TGFHYXdkSU5weGF2Ck5TOFJqVU41YXlHSjY5N2JRdGVJMHJFWmVxYXdDRno1NlpkYlJlSUk2bmJNWmNrTXFmd0oyQWFNaEE5cnljR0kKYmU1aFFzQ1dUR0Ruamh1ZllQRGVXQXJzMDQ4MExRU05DVGpoL1ZHMGVWdFoxVkhpZ25kL2g0c2pWT1lqY0NEcQpVUDl4UXkxMzdVUmwxdW9PcEV0bGJhaDZPRkt2VDBtZG0weEpiN3N0QkpZT29Kc2hlOFZCcUw3cENETlJLNG1hCnhUcGtpZjRIbUx5d1hCVTNXa3NSRVNtNDJwbThkY2V5eW05WQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
  name: dev-aws-de-fra-1
contexts:
- context:
    cluster: dev-aws-de-fra-1
    namespace: dev-aws-de-fra-1
    user: dev-aws-de-fra-1
  name: dev-aws-de-fra-1
current-context: dev-aws-de-fra-1
kind: Config
preferences: {}
users:
- name: dev-aws-de-fra-1
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: aws-iam-authenticator
      args:
        - "token"
        - "-i"
        - "dev-aws-de-fra-1"
`

var malformedSampleKubeConfig string = `apiVersion: v1
clusters:
- cluster:
    server: https://CA05F37FDEAFEC44E0567EF1632622AA.gr7.eu-central-1.eks.amazonaws.com
    certificate-authority-data: LS0XXXXXX
  name: dev-aws-de-fra-1
contexts:
- context:
    cluster: dev-aws-de-fra-1
    namespace: dev-aws-de-fra-1
    user: dev-aws-de-fra-1
  name: dev-aws-de-fra-1
current-context: dev-aws-de-fra-1
kind: Config
preferences: {}
users:
- name: dev-aws-de-fra-1
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: aws-iam-authenticator
      args:
        - "token"
        - "-i"
        - "dev-aws-de-fra-1"
`

func TestInitKubeClusterConfig(t *testing.T) {
	var testScenarios = map[string]initKubeClusterConfig{
		"nominal": {
			kubeconfigEnvSet:     true,
			correctKubeconfigEnv: true,
			file: &testFile{
				path:    "kubeconfig",
				content: sampleKubeConfig,
			},
			expectedError:           false,
			multiplePathsKubeConfig: false,
		},
		"malformed_config": {
			kubeconfigEnvSet: true,
			file: &testFile{
				path:    "kubeconfig",
				content: malformedSampleKubeConfig,
			},
			expectedError:           true,
			multiplePathsKubeConfig: false,
		},
		"wrong_env_var": {
			kubeconfigEnvSet:     true,
			correctKubeconfigEnv: false,
			file: &testFile{
				path:    "kubeconfig",
				content: sampleKubeConfig,
			},
			expectedError:           true,
			multiplePathsKubeConfig: false,
		},
		"no_env_var": {
			kubeconfigEnvSet: false,
			file: &testFile{
				path:    "kubeconfig",
				content: sampleKubeConfig,
			},
			expectedError:           true,
			multiplePathsKubeConfig: false,
		},
		"multiple_configs_in_env_var": {
			kubeconfigEnvSet:     true,
			correctKubeconfigEnv: true,
			file: &testFile{
				path:    "kubeconfig",
				content: sampleKubeConfig,
			},
			expectedError:           false,
			multiplePathsKubeConfig: true,
		},
	}

	for k, v := range testScenarios {
		t.Run(k, func(t *testing.T) {
			os.Unsetenv("KUBERNETES_SERVICE_HOST")
			os.Unsetenv("KUBERNETES_SERVICE_PORT")

			dir, err := os.MkdirTemp("", "TestInitKubeClusterConfig")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)

			err = CreateDirectory(dir + "/" + k)
			if err != nil {
				t.Fatal(err)
			}

			if v.file != nil {
				file, err := os.Create(dir + "/" + k + "/" + v.file.path)
				if err != nil {
					t.Fatal(err)
				}

				_, err = file.Write([]byte(v.file.content))
				if err != nil {
					t.Fatal(err)
				}
			}
			if v.kubeconfigEnvSet {
				if v.correctKubeconfigEnv {
					err = os.Setenv("KUBECONFIG", dir+"/"+k+"/"+v.file.path)
					if err != nil {
						t.Fatal(err)
					}
					if v.multiplePathsKubeConfig {
						err = os.Setenv("KUBECONFIG", os.Getenv("KUBECONFIG")+":"+dir+"/"+k+"/"+v.file.path)
						if err != nil {
							t.Fatal(err)
						}
					}
				} else {
					err = os.Setenv("KUBECONFIG", "not/the/right/path")
					if err != nil {
						t.Fatal(err)
					}
				}
				defer os.Unsetenv("KUBECONFIG")
			}

			config, clientset, err := InitKubeClusterConfig()

			if v.expectedError {
				assert.Error(t, err, "should have an error")
				assert.Nil(t, config)
				assert.Nil(t, clientset)
			} else {
				assert.NoError(t, err, "error should be nil")
				assert.IsType(t, &rest.Config{}, config)
				assert.NotNil(t, config)
				assert.IsType(t, &kubernetes.Clientset{}, clientset)
				assert.NotNil(t, clientset)
			}
		})
	}
}

func TestHasStagedGitChanges(t *testing.T) {
	// Create a temporary directory for the test repository
	dir, err := os.MkdirTemp("", "go-git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // Clean up the temporary directory

	// Initialize a Git repository
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatal(err)
	}

	// Test cases
	testCases := []struct {
		setup    func(*git.Worktree) error
		expected bool
		testName string
	}{
		{
			setup: func(w *git.Worktree) error {
				// No changes - empty repository
				return nil
			},
			expected: false,
			testName: "Empty repository",
		},
		{
			setup: func(w *git.Worktree) error {
				// Untracked file
				_, err := os.Create(dir + "/untracked.txt")
				return err
			},
			expected: false,
			testName: "Untracked file",
		},
		{
			setup: func(w *git.Worktree) error {
				// Staged file
				err := os.WriteFile(dir+"/staged.txt", []byte("content"), 0600)
				if err != nil {
					return err
				}

				_, err = w.Add("staged.txt")
				if err != nil {
					return err
				}
				return nil
			},
			expected: true,
			testName: "Staged file",
		},
		{
			setup: func(w *git.Worktree) error {
				// Staged and modified file
				err := os.WriteFile(dir+"/both.txt", []byte("initial content"), 0600)
				if err != nil {
					return err
				}
				_, err = w.Add("both.txt")
				if err != nil {
					return err
				}
				// Stage the file, then modify it further, like after a 'git add' but before commit
				err = os.WriteFile(dir+"/both.txt", []byte("modified content"), 0600)
				if err != nil {
					return err
				}

				return nil
			},
			expected: true,
			testName: "Staged and modified file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			w, err := repo.Worktree()
			if err != nil {
				t.Fatal(err)
			}

			// Setup for the test case
			err = tc.setup(w)
			if err != nil {
				t.Fatal(err)
			}

			hasStaged, err := HasStagedGitChanges(repo)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.expected, hasStaged, tc.testName)
		})
	}
}
