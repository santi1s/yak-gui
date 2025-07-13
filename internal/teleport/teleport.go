package teleport

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	"github.com/go-ini/ini"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
)

const (
	teleportConfigPath                = "/configs/teleport/config.yml"
	AWSConfigType                     = "aws"
	DBConfigType                      = "db"
	minimumSessionDuration            = 900
	minimumTeleportRecommendedSession = 30
	SSOStartURL                       = "https://doctolib.awsapps.com/start"
	SSORegion                         = "eu-central-1"
)

func TshLogin(tConfig Config) error {
	v, err := checkVersion(tConfig.Version)
	if err != nil {
		return err
	}
	if !v {
		return fmt.Errorf("please update your tsh version to %s", tConfig.Version)
	}
	cmd := tshExecHelper("login", "--proxy", tConfig.Cluster, tConfig.Cluster)
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("%s", cmd.Stderr)
	}
	return nil
}

func GetTshStatus() (ProfileStatus, error) {
	cmd := tshExecHelper("status", "-f", "json")
	out, err := cmd.Output()
	if err != nil {
		return ProfileStatus{}, fmt.Errorf("%s", cmd.Stderr)
	}

	var profile ActiveProfileStatus
	err = json.Unmarshal(out, &profile)
	if err != nil {
		return ProfileStatus{}, err
	}

	return profile.Active, nil
}

func TshRequestNeededRoles(rolesToRequest []Role, username string, targets string, expire string, reviewers []string, rType string) error {
	approvedRequests, err := GetTshApprovedRequests(rolesToRequest, username, rType)
	if err != nil {
		return fmt.Errorf("error getting Teleport approved requests: %s", err)
	}

	requestsToAssume := AskForRequestsToAssumes(approvedRequests)
	if len(requestsToAssume) > 0 {
		err = TshAssumeRoles(requestsToAssume)
		if err != nil {
			return fmt.Errorf("error assuming Teleport roles: %s", err)
		}
	}

	assumed := []string{}
	for _, r := range requestsToAssume {
		assumed = append(assumed, r.Spec.Roles...)
	}

	var missingToRequest []Role
	for _, r := range rolesToRequest {
		if !helper.ListContainsString(assumed, r.Name) {
			missingToRequest = append(missingToRequest, r)
		}
	}

	if len(missingToRequest) > 0 {
		err = RequestTshPrivilegedAccess(targets, missingToRequest, expire, reviewers, rType)
		if err != nil {
			return fmt.Errorf("error requesting Teleport privileged access: %s", err)
		}
	}
	return nil
}

func checkVersion(target string) (bool, error) {
	cmd := tshExecHelper("version", "-f", "json")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("%s", cmd.Stderr)
	}

	var tVersion map[string]interface{}
	err = json.Unmarshal([]byte(out), &tVersion)
	if err != nil {
		return false, err
	}

	tshVersion := tVersion["version"].(string)
	if tshVersion == "" {
		return false, fmt.Errorf("error parsing tsh version JSON: version is empty")
	}

	isEqual := tshVersion == target

	if !isEqual {
		log.Debugf("tsh detected version: %s", tshVersion)
		log.Debugf("tsh target version: %s", target)
		return false, nil
	}

	return true, nil
}

func CheckIfNeedToLogout(sessionTTL string) (bool, error) {
	ttl, err := time.Parse(time.RFC3339, sessionTTL)
	if err != nil {
		return false, err
	}

	if time.Until(ttl) < minimumTeleportRecommendedSession*time.Minute {
		cli.Println(fmt.Sprintf("Session TTL too low %s, we recommend you to  re-login to Teleport.", time.Until(ttl).Round(time.Minute).String()))
		return true, nil
	}
	return false, nil
}

func RequestTshPrivilegedAccess(env string, roles []Role, sessionTTL string, reviewers []string, rType string) error {
	cli.Printf("Why are you requesting privileged %s access on %s ? ", rType, env)
	reason, err := cli.ReadLine()
	if err != nil {
		return err
	}

	sessionTTLTime, err := time.Parse(time.RFC3339, sessionTTL)
	if err != nil {
		fmt.Println("Error parsing timestamp:", err)
		return err
	}

	maxDuration := time.Until(sessionTTLTime).Round(time.Minute)
	rolesStr := []string{}
	for _, role := range roles {
		rolesStr = append(rolesStr, role.Name)
		d, err := time.ParseDuration(role.MaxDuration)
		if err != nil {
			return err
		}
		if d < maxDuration {
			maxDuration = d
		}
	}

	tDuration := AskForRequestDuration(rolesStr, maxDuration, rType)

	if len(reviewers) == 0 {
		for _, role := range roles {
			for _, ug := range role.SlackReviewersUserGroups {
				if !helper.ListContainsString(reviewers, ug) {
					reviewers = append(reviewers, ug)
				}
			}
		}
	}

	cmdArray := []string{"request", "create", "--roles", strings.Join(rolesStr, ","), "--reason", reason, "--session-ttl", tDuration}
	if len(reviewers) > 0 {
		cmdArray = append(cmdArray, "--reviewers="+strings.Join(reviewers, ","))
	}
	cmd := tshExecHelper(cmdArray...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func ReviewersToSlice(reviewers string) []string {
	if reviewers == "" {
		return []string{}
	}
	return strings.Split(reviewers, ",")
}

func AskForRequestDuration(roles []string, maxDuration time.Duration, rType string) string {
	for {
		cli.Printf("For what duration you need access to %s role(s) (max %s)? ", strings.Join(roles, ","), maxDuration.String())
		duration, err := cli.Read()
		if err != nil {
			log.Errorf("Error reading duration: %s", err)
			continue
		}

		tDuration, err := time.ParseDuration(duration)
		if err != nil {
			log.Errorf("Error parsing duration: %s", err)
			continue
		}

		if tDuration > maxDuration {
			log.Errorf("duration exceeds max duration for role (max: %s)", maxDuration.String())
			continue
		}

		if rType == AWSConfigType && tDuration < minimumTeleportRecommendedSession*time.Minute && tDuration < maxDuration {
			duration = (minimumTeleportRecommendedSession * time.Minute).String()
		}

		return duration
	}
}
func GetFQDNNLevelDomain(fqdn string, level int) string {
	parts := strings.Split(fqdn, ".")
	if len(parts) >= level+1 {
		return parts[level]
	}
	return ""
}

func GetTeleportDBResourceFromURI(endpoint string) (Resource, error) {
	cmd := tshExecHelper("db", "ls", "-f", "json")
	out, err := cmd.Output()
	if err != nil {
		return Resource{}, fmt.Errorf("%s", cmd.Stderr)
	}

	var resources []Resource
	err = json.Unmarshal(out, &resources)
	if err != nil {
		return Resource{}, err
	}

	for _, resource := range resources {
		if strings.Contains(resource.Spec.URI, GetFQDNNLevelDomain(endpoint, 1)) && strings.Contains(GetFQDNNLevelDomain(endpoint, 0), GetFQDNNLevelDomain(resource.Spec.URI, 0)) {
			return resource, nil
		}
	}
	return Resource{}, errors.New("could not find DB resource in Teleport")
}

func TshDBLogin(tEndpoint string, dbuser string, dbname string) error {
	cmd := tshExecHelper("db", "login", "--db-user", dbuser, "--db-name", dbname, tEndpoint)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s", cmd.Stderr)
	}
	return nil
}

func TshAwsLogin(roles []Role) error {
	errs, _ := errgroup.WithContext(context.Background())

	for _, role := range roles {
		r := role
		errs.Go(func() error {
			log.Infof("Logging in to AWS application %s ...", r.AWSApplicationName)
			cmd := tshExecHelper("apps", "login", r.AWSApplicationName, "--aws-role", "teleport-"+r.Name)
			_, err := cmd.Output()
			if err != nil {
				return fmt.Errorf("%s", cmd.Stderr)
			}
			return nil
		})
	}

	return errs.Wait()
}

func TshAwsProxy(roles []Role) ([]ProxyConfig, []*exec.Cmd, error) {
	proxyConfig := []ProxyConfig{}
	var childs []*exec.Cmd
	for _, role := range roles {
		cmd := tshExecHelper("proxy", "aws", "--app", role.AWSApplicationName, "-e")
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			return nil, childs, err
		}
		if err := cmd.Start(); err != nil {
			return nil, childs, fmt.Errorf("%s", cmd.Stderr)
		}
		childs = append(childs, cmd)

		var lines []string
		scanner := bufio.NewScanner(stdoutPipe)
		go func() {
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
		}()
		time.Sleep(3 * time.Second)

		if len(cmd.Stderr.(*bytes.Buffer).Bytes()) != 0 {
			return nil, childs, fmt.Errorf("%s", cmd.Stderr)
		}

		config := parseAWSProxyTextConfig(strings.Join(lines, "\n"))
		config.TargetProfiles = role.AWSConfigProfiles
		proxyConfig = append(proxyConfig, config)
	}
	return proxyConfig, childs, nil
}

func TshLogout() error {
	cmd := exec.Command("sh", "-c", "env -u KUBECONFIG tsh logout")
	log.Debugf("executing %s", cmd.String())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("%s", cmd.Stderr)
	}
	return nil
}

func parseAWSProxyTextConfig(text string) ProxyConfig {
	endpointURLRegex := regexp.MustCompile(`AWS endpoint URL at (.+?)\.\n`)
	accessKeyRegex := regexp.MustCompile(`AWS_ACCESS_KEY_ID=(.+)`)
	secretKeyRegex := regexp.MustCompile(`AWS_SECRET_ACCESS_KEY=(.+)`)
	caBundleRegex := regexp.MustCompile(`AWS_CA_BUNDLE=(.+)`)

	awsConfig := ProxyConfig{
		EndpointURL:     endpointURLRegex.FindStringSubmatch(text)[1],
		AccessKeyID:     accessKeyRegex.FindStringSubmatch(text)[1],
		SecretAccessKey: secretKeyRegex.FindStringSubmatch(text)[1],
		AWSCaBundle:     caBundleRegex.FindStringSubmatch(text)[1],
	}
	return awsConfig
}

func TshGeneratePsqlArgsPrefix(tEndpoint string) (string, error) {
	cmd := tshExecHelper("db", "config", "--format", "json", tEndpoint)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s", cmd.Stderr)
	}

	var config TeleportDBConnexionConfig
	if err := json.Unmarshal(out, &config); err != nil {
		return "", err
	}

	prefix := fmt.Sprintf("user=%s host=%s port=%d dbname=%s sslrootcert=%s sslcert=%s sslkey=%s sslmode=verify-full",
		config.User, config.Host, config.Port, config.Database, config.CA, config.Cert, config.Key)

	return prefix, nil
}

func ReadTeleportConfig() (Config, error) {
	envPath := os.Getenv("TFINFRA_REPOSITORY_PATH")
	if envPath == "" {
		return Config{}, errors.New("TFINFRA_REPOSITORY_PATH is not set")
	}
	ymlPath := filepath.Join(envPath, teleportConfigPath)

	file, err := os.ReadFile(ymlPath)
	if err != nil {
		return Config{}, err
	}

	var tConfig Config
	err = yaml.Unmarshal(file, &tConfig)
	if err != nil {
		return Config{}, err
	}
	return tConfig, nil
}

func TeleportRolesToRequest(tConfig Config, permission string, tshRoles, requestedAccounts []string, roleType string) ([]Role, []Role) {
	rolesToRequest, rolesAlreadyAssumed := []Role{}, []Role{}
	for _, account := range requestedAccounts {
		for accountNo, config := range tConfig.Accounts {
			if config.Name != account {
				continue
			}

			for _, role := range config.Roles {
				if role.Type != roleType || role.Permission != permission {
					continue
				}

				if helper.ListContainsString(tshRoles, role.Name) {
					role.AccountNo = accountNo
					rolesAlreadyAssumed = append(rolesAlreadyAssumed, role)
					continue
				}
				role.AccountNo = accountNo
				rolesToRequest = append(rolesToRequest, role)
			}
		}
	}
	return rolesToRequest, rolesAlreadyAssumed
}

func BackupRolesToRequest(tConfig Config, permission string, requestedAccounts []string, roleType string) []Role {
	rolesToRequest := []Role{}
	for _, account := range requestedAccounts {
		for accountNo, config := range tConfig.Accounts {
			if config.Name != account {
				continue
			}

			for _, role := range config.Roles {
				if role.Type != roleType || role.Permission != permission {
					continue
				}
				role.AccountNo = accountNo
				rolesToRequest = append(rolesToRequest, role)
			}
		}
	}
	return rolesToRequest
}

func OverrideAWSConfigProfiles(config *ini.File, roles []Role, exceptKeys []string, bypassTsh bool) error {
	for _, role := range roles {
		mSection, err := config.GetSection("profile " + role.Name)
		if err != nil {
			mSection, _ = config.NewSection("profile " + role.Name)
		}
		if !bypassTsh {
			for key := range mSection.KeysHash() {
				mSection.DeleteKey(key)
			}
			cmd := []string{"yak", "aws", "config", "generate", "--app", role.AWSApplicationName, "--role", constructAWSRoleARN(role)}
			_, err = mSection.NewKey("credential_process", strings.Join(cmd, " "))
			if err != nil {
				return err
			}
			_, err = mSection.NewKey("region", "eu-central-1")
			if err != nil {
				return err
			}
		} else {
			for key := range mSection.KeysHash() {
				mSection.DeleteKey(key)
			}
			_, err = mSection.NewKey("sso_region", SSORegion)
			if err != nil {
				return err
			}
			_, err = mSection.NewKey("sso_start_url", SSOStartURL)
			if err != nil {
				return err
			}
			_, err = mSection.NewKey("sso_account_id", role.AccountNo)
			if err != nil {
				return err
			}
			if role.BackupSSORoleName == "" {
				return fmt.Errorf("backup_sso_role_name is not set in teleport configuration for role %s", role.Name)
			}
			_, err = mSection.NewKey("sso_role_name", role.BackupSSORoleName)
			if err != nil {
				return err
			}
			_, err = mSection.NewKey("output", "json")
			if err != nil {
				return err
			}
			_, err = mSection.NewKey("region", SSORegion)
			if err != nil {
				return err
			}
		}
		for _, profile := range role.AWSConfigProfiles {
			section, err := config.GetSection("profile " + profile)
			if err != nil {
				section, _ = config.NewSection("profile " + profile)
			}
			for key := range section.KeysHash() {
				if !helper.ListContainsString(exceptKeys, key) {
					section.DeleteKey(key)
				}
			}
			cmd := []string{"aws-vault", "exec", role.Name, "--json"}
			_, err = section.NewKey("credential_process", strings.Join(cmd, " "))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func constructAWSRoleARN(role Role) string {
	return fmt.Sprintf("arn:aws:iam::%s:role/teleport-%s", role.AccountNo, role.Name)
}

func GetTshApprovedRequests(roles []Role, username string, rType string) ([]AccessRequest, error) {
	cmd := tshExecHelper("request", "ls", "-f", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s", cmd.Stderr)
	}

	var requests []AccessRequest
	if string(out) == "null" {
		return nil, nil
	}
	err = json.Unmarshal(out, &requests)
	if err != nil {
		return nil, err
	}

	var approvedRequests []AccessRequest
	for _, request := range requests {
		for _, role := range roles {
			if request.Spec.State == 2 && request.Spec.User == username && helper.ListContainsString(request.Spec.Roles, role.Name) {
				expire, err := time.Parse(time.RFC3339, request.Metadata.Expires)
				if err != nil {
					return nil, err
				}
				if rType != AWSConfigType || time.Until(expire).Round(time.Minute) > minimumTeleportRecommendedSession*time.Minute {
					request.Metadata.Expires = time.Until(expire).Round(time.Minute).String()
					approvedRequests = append(approvedRequests, request)
					break
				}
			}
		}
	}

	return approvedRequests, nil
}

func AskForRequestsToAssumes(requests []AccessRequest) []AccessRequest {
	var requestToAssume []AccessRequest
	for _, request := range requests {
		cli.Println("You have a related approved request that you can assume:")
		cli.Println("\tID:", request.Metadata.Name)
		cli.Println("\tExpires:", request.Metadata.Expires)
		cli.Println("\tRoles:", request.Spec.Roles)
		cli.Println("\tRequest reason:", request.Spec.Reason)
		if cli.AskConfirmation("Do you want to assume it ?") {
			requestToAssume = append(requestToAssume, request)
		}
	}
	return requestToAssume
}

func TshAssumeRoles(requests []AccessRequest) error {
	for _, request := range requests {
		cmd := tshExecHelper("login", "--request-id", request.Metadata.Name)
		_, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("%s", cmd.Stderr)
		}
	}
	return nil
}

func GenerateAWSCredentialsProcess(app string, role string, name string, ttl string) (string, error) {
	t, err := time.Parse(time.RFC3339, ttl)
	if err != nil {
		return "", err
	}

	expires := time.Now().Add(minimumSessionDuration * time.Second)
	if !t.After(expires) {
		return "", errors.New("duration is less than minimum session duration of 15 minutes, please re-login to Teleport")
	}

	cmd := tshExecHelper("aws", "--app", app, "sts", "assume-role", "--role-arn", role, "--role-session-name", name, "--duration-seconds", strconv.Itoa(minimumSessionDuration))
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s", cmd.Stderr)
	}

	var resp AssumeRoleResponse
	err = json.Unmarshal(out, &resp)
	if err != nil {
		return "", err
	}

	j, err := json.MarshalIndent(CredentialProcessResponse{
		Version:         1,
		AccessKeyID:     resp.Credentials.AccessKeyID,
		SecretAccessKey: resp.Credentials.SecretAccessKey,
		SessionToken:    resp.Credentials.SessionToken,
		Expiration:      resp.Credentials.Expiration,
	}, "", "  ")
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func tshExecHelper(cmd ...string) *exec.Cmd {
	tshCmd := exec.Command("tsh", cmd...) //#nosec G204
	log.Debugf("executing %s", tshCmd.String())
	var stderr bytes.Buffer
	tshCmd.Stderr = &stderr
	return tshCmd
}
