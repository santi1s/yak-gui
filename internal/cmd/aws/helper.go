package aws

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	log "github.com/sirupsen/logrus"
)

func matchStringWithWildcard(input string, patterns []string) bool {
	for _, pattern := range patterns {
		match, err := filepath.Match(pattern, input)
		if err != nil {
			log.Errorf("Error matching pattern '%s': %v\n", pattern, err)
			continue
		}
		if match {
			return match
		}
	}
	return false
}

func getAWSAccountNo(client *sts.Client) (string, error) {
	result, err := client.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}

	return *result.Account, nil
}

const rdsEndpointAvailableStatus = "available"

func targetEndpoint(cluster types.DBCluster, writer bool) string {
	if writer {
		return *cluster.Endpoint
	}
	return *cluster.ReaderEndpoint
}

func endpointRegion(endpoint string) string {
	return strings.Split(endpoint[:len(endpoint)-len(".rds.amazonaws.com")], ".")[2]
}

func generateIamDBPassword(cfg *aws.Config, dbEndpoint string, dbPort int, dbRegion string, dbUser string) (string, error) {
	dbEndpoint = fmt.Sprintf("%s:%d", dbEndpoint, dbPort)

	authToken, err := auth.BuildAuthToken(
		context.TODO(),
		dbEndpoint,
		dbRegion,
		dbUser,
		cfg.Credentials)

	if err != nil {
		return "", err
	}
	return authToken, nil
}

func getCluster(client *rds.Client, clusterIdentifier string) (types.DBCluster, error) {
	input := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(clusterIdentifier),
	}

	output, err := client.DescribeDBClusters(context.TODO(), input)
	if err != nil {
		return types.DBCluster{}, fmt.Errorf("error describing RDS cluster: %w", err)
	}

	if len(output.DBClusters) == 0 {
		return types.DBCluster{}, fmt.Errorf("no clusters found with identifier: %s", clusterIdentifier)
	}

	return output.DBClusters[0], nil
}

func getInstance(client *rds.Client, instanceIdentifier string) (types.DBInstance, error) {
	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instanceIdentifier),
	}

	output, err := client.DescribeDBInstances(context.TODO(), input)
	if err != nil {
		return types.DBInstance{}, fmt.Errorf("error describing RDS cluster instance: %w", err)
	}

	if len(output.DBInstances) == 0 {
		return types.DBInstance{}, fmt.Errorf("no cluster instances found with identifier: %s", instanceIdentifier)
	}

	return output.DBInstances[0], nil
}

func clustersExtractInfo(dbClusters []types.DBCluster) []ClusterInfo {
	var res []ClusterInfo
	for _, cluster := range dbClusters {
		info := ClusterInfo{}
		if cluster.DBClusterIdentifier != nil {
			info.ClusterName = *cluster.DBClusterIdentifier
		}
		if cluster.Endpoint != nil {
			info.Endpoint = *cluster.Endpoint
		}
		if cluster.ReaderEndpoint != nil {
			info.ReaderEndpoint = *cluster.ReaderEndpoint
		}
		if cluster.DatabaseName != nil {
			info.DBName = *cluster.DatabaseName
		}
		res = append(res, info)
	}
	return res
}

func getEndpointStatus(client *rds.Client, clusterIdentifier string, endpoint string) (string, error) {
	input := &rds.DescribeDBClusterEndpointsInput{
		DBClusterIdentifier: aws.String(clusterIdentifier),
	}

	output, err := client.DescribeDBClusterEndpoints(context.TODO(), input)
	if err != nil {
		return "", err
	}

	if len(output.DBClusterEndpoints) == 0 {
		return "", fmt.Errorf("no endpoints found for cluster: %s", clusterIdentifier)
	}

	for _, ep := range output.DBClusterEndpoints {
		if *ep.Endpoint == endpoint {
			return *ep.Status, nil
		}
	}
	return "", fmt.Errorf("no endpoint found with name: %s", endpoint)
}

func checkIfEndpointBelongsToCluster(client *rds.Client, clusterIdentifier string, endpoint string) (bool, error) {
	input := &rds.DescribeDBClusterEndpointsInput{
		DBClusterIdentifier: aws.String(clusterIdentifier),
	}

	output, err := client.DescribeDBClusterEndpoints(context.TODO(), input)
	if err != nil {
		return false, err
	}

	if len(output.DBClusterEndpoints) == 0 {
		return false, fmt.Errorf("no endpoints found for cluster: %s", clusterIdentifier)
	}

	for _, ep := range output.DBClusterEndpoints {
		if *ep.Endpoint == endpoint {
			return true, nil
		}
	}
	return false, nil
}

func fmtAWSTags(tags map[string]string) []types.Tag {
	var awsTags []types.Tag
	for k, v := range tags {
		awsTags = append(awsTags, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	return awsTags
}

func getSecondSinceTime(startTime time.Time) int64 {
	t := time.Now()
	return int64(t.Sub(startTime).Seconds())
}

func sortCloneFlags(cmd *cobra.Command, args []string) {
	if len(cloneProvidedFlags.targets) > 0 {
		sort.Strings(cloneProvidedFlags.targets)
	}
	if len(cloneProvidedFlags.sources) > 0 {
		sort.Strings(cloneProvidedFlags.sources)
	}
}
