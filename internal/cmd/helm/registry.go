package helm

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/santi1s/yak/internal/cmd/aws"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ECRRegistry struct {
	Account string `yaml:"account"`
	Region  string `yaml:"region"`
	Profile string `yaml:"profile"`
}

type HelmRegistries struct {
	ECR []*ECRRegistry `yaml:"ecr"`
}

func NewRegistries(configFile string) (*HelmRegistries, error) {
	var registries *HelmRegistries
	file, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, &registries)
	if err != nil {
		return nil, err
	}
	return registries, nil
}

func (hr *HelmRegistries) Authenticate() error {
	for _, ecr := range hr.ECR {
		if err := ecr.Login(); err != nil {
			return err
		}
	}
	return nil
}

func (ecr *ECRRegistry) Hostname() string {
	return fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", ecr.Account, ecr.Region)
}

func (ecr *ECRRegistry) Login() error {
	log.Infof("Authenticating to ECR registry %s", ecr.Hostname())
	username, password, err := aws.EcrLogin(ecr.Region, ecr.Profile)
	if err != nil {
		return err
	}
	if err := RegistryLogin(username, password, ecr.Hostname()); err != nil {
		return err
	}
	return nil
}

func RegistryLogin(username, password, ecrHostname string) error {
	loginCmd := exec.Command("helm", "registry", "login", "--username", username, "--password-stdin", ecrHostname)
	loginCmd.Stdin = strings.NewReader(password)
	loginOutput, err := loginCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to login to registry: %v\n%s", err, string(loginOutput))
	}
	return nil
}
