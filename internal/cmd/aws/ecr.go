package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

func EcrLogin(awsRegion string, awsProfile string) (string, string, error) {
	var cfg aws.Config
	var err error

	if os.Getenv("YAK_ECR_AWS_NO_PROFILE") == "true" {
		cfg, err = config.LoadDefaultConfig(
			context.TODO(),
			config.WithRegion(awsRegion),
		)
		if err != nil {
			return "", "", fmt.Errorf("failed to load AWS config: %v", err)
		}
	} else {
		cfg, err = config.LoadDefaultConfig(
			context.TODO(),
			config.WithRegion(awsRegion),
			config.WithSharedConfigProfile(awsProfile))
		if err != nil {
			return "", "", fmt.Errorf("failed to load AWS config: %v", err)
		}
	}

	client := ecr.NewFromConfig(cfg)

	authTokenOutput, err := client.GetAuthorizationToken(context.TODO(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", "", fmt.Errorf("failed to get ECR authorization token: %v", err)
	}

	if len(authTokenOutput.AuthorizationData) == 0 {
		return "", "", fmt.Errorf("no authorization data found")
	}

	authData := authTokenOutput.AuthorizationData[0]
	token, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode authorization token: %v", err)
	}

	parts := strings.SplitN(string(token), ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid authorization token format")
	}
	username, password := parts[0], parts[1]
	return username, password, nil
}
