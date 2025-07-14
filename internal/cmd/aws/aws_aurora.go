package aws

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/smithy-go"
	"github.com/santi1s/yak/cli"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type RDSAPI interface {
	DescribeDBClusters(ctx context.Context, params *rds.DescribeDBClustersInput, optFns ...func(*rds.Options)) (*rds.DescribeDBClustersOutput, error)
	DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
	DescribeDBClusterSnapshots(ctx context.Context, params *rds.DescribeDBClusterSnapshotsInput, optFns ...func(*rds.Options)) (*rds.DescribeDBClusterSnapshotsOutput, error)
	DeleteDBInstance(ctx context.Context, params *rds.DeleteDBInstanceInput, optFns ...func(*rds.Options)) (*rds.DeleteDBInstanceOutput, error)
	DeleteDBCluster(ctx context.Context, params *rds.DeleteDBClusterInput, optFns ...func(*rds.Options)) (*rds.DeleteDBClusterOutput, error)
	CreateDBInstance(ctx context.Context, params *rds.CreateDBInstanceInput, optFns ...func(*rds.Options)) (*rds.CreateDBInstanceOutput, error)
	RestoreDBClusterToPointInTime(ctx context.Context, params *rds.RestoreDBClusterToPointInTimeInput, optFns ...func(*rds.Options)) (*rds.RestoreDBClusterToPointInTimeOutput, error)
	RestoreDBClusterFromSnapshot(ctx context.Context, params *rds.RestoreDBClusterFromSnapshotInput, optFns ...func(*rds.Options)) (*rds.RestoreDBClusterFromSnapshotOutput, error)
}

var (
	auroraCmd = &cobra.Command{
		Use:   "aurora",
		Short: "aurora management",
	}
	errDBInstanceAlreadyExists     = errors.New("DBInstance already exists")
	errDBInstanceNotFound          = errors.New("DBInstance not found")
	errDBClusterNotFound           = errors.New("DBCluster not found")
	errNoDBClusterSrcFound         = errors.New("DB cluster source(s) not found")
	errDeleteInstance              = errors.New("error while deleting instance")
	errDeleteCluster               = errors.New("error while deleting cluster")
	errDBSubnetGroupNotFound       = errors.New("subnet group not found")
	errDBParameterGroupNotFound    = errors.New("parameter group not found")
	errDBClusterSnapshotNotFound   = errors.New("snapshot not found")
	errInvalidVPCNetworkState      error
	errInvalidParameterCombination error
	errInvalidParameterValue       error
)

func init() {
	auroraCmd.AddCommand(auroraCloneCmd)
	auroraCmd.AddCommand(auroraClusterCmd)
	auroraCmd.AddCommand(auroraPsqlCmd)
}

func initAwsRdsClient() RDSAPI {
	// init aws config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	client := rds.NewFromConfig(cfg)
	return client
}

// Refact idea: describeCluster, discoverClone, discoverSrc use the same api call with different parameters
func describeCluster(client RDSAPI, name string) []types.DBCluster {
	output, err := client.DescribeDBClusters(
		context.TODO(),
		&rds.DescribeDBClustersInput{
			DBClusterIdentifier: aws.String(name),
		},
	)
	if err != nil {
		var dbcnff *types.DBClusterNotFoundFault
		if errors.As(err, &dbcnff) {
			cli.Printf("Cluster %s not found, continuing ...\n", name)
		}
		return nil
	}
	return output.DBClusters
}

func describeClusterSnapshot(client RDSAPI, name string) []types.DBClusterSnapshot {
	output, err := client.DescribeDBClusterSnapshots(
		context.TODO(),
		&rds.DescribeDBClusterSnapshotsInput{
			DBClusterSnapshotIdentifier: aws.String(name),
			IncludeShared:               aws.Bool(true),
		},
	)
	if err != nil {
		var dbcsnff *types.DBClusterSnapshotNotFoundFault
		if errors.As(err, &dbcsnff) {
			cli.Printf("Cluster snapshot %s not found, continuing ...\n", name)
		}
		return nil
	}
	return output.DBClusterSnapshots
}

func clusterExist(client RDSAPI, name string) bool {
	clusters := describeCluster(client, name)
	return len(clusters) > 0
}

func clusterAvailable(client RDSAPI, name string) bool {
	clusters := describeCluster(client, name)

	if clusters != nil && *clusters[0].Status == "available" {
		cli.Printf("Cluster %s is available\n", *clusters[0].DBClusterIdentifier)
		return true
	}
	return false
}

func describeInstance(client RDSAPI, name string) []types.DBInstance {
	output, err := client.DescribeDBInstances(
		context.TODO(),
		&rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(name),
		})
	if err != nil {
		var dbinff *types.DBInstanceNotFoundFault
		if errors.As(err, &dbinff) {
			cli.Printf("Instance %s not found, continuing ...\n", name)
		}
		return nil
	}
	return output.DBInstances
}

func instanceExist(client RDSAPI, name string) bool {
	instances := describeInstance(client, name)
	return len(instances) > 0
}

func instanceAvailable(client RDSAPI, name string) bool {
	instances := describeInstance(client, name)

	if instances != nil && *instances[0].DBInstanceStatus == "available" {
		cli.Printf("Instance %s is available\n", *instances[0].DBInstanceIdentifier)
		return true
	}
	return false
}

// Refact idea: describeCluster, discoverClone, discoverSrc use the same api call with different parameters
func discoverClone(client RDSAPI, targets []string) []types.DBCluster {
	var result []types.DBCluster

	output, err := client.DescribeDBClusters(context.TODO(), &rds.DescribeDBClustersInput{})
	if err != nil {
		cli.Printf("Error: %s \n", err)
	}
	for _, cluster := range output.DBClusters {
		for _, target := range targets {
			if cluster.DBClusterIdentifier != nil && strings.Contains(*cluster.DBClusterIdentifier, target) {
				result = append(result, cluster)
			}
		}
	}
	if len(result) == 0 {
		cli.Printf("No cluster found for: %v\n", targets)
		return nil
	}
	return result
}

// Refact idea: describeCluster, discoverClone, discoverSrc use the same api call with différent parameters
func discoverSrc(client RDSAPI, names []string) ([]types.DBCluster, error) {
	output, err := client.DescribeDBClusters(
		context.TODO(),
		&rds.DescribeDBClustersInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("db-cluster-id"),
					Values: names,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if len(output.DBClusters) == 0 {
		return nil, errNoDBClusterSrcFound
	}
	return output.DBClusters, nil
}

func deleteInstance(client RDSAPI, name string) (*rds.DeleteDBInstanceOutput, error) {
	outputInstance, err := client.DeleteDBInstance(
		context.TODO(),
		&rds.DeleteDBInstanceInput{
			DBInstanceIdentifier: aws.String(name),
			SkipFinalSnapshot:    aws.Bool(true),
		},
	)
	if err != nil {
		var dbinff *types.DBInstanceNotFoundFault
		if errors.As(err, &dbinff) {
			cli.Printf("Instance %s not found, continuing ...\n", name)
			return nil, errDBInstanceNotFound
		}
	}
	cli.Printf("Deleting instance %s\n", name)
	return outputInstance, nil
}

func deleteCluster(client RDSAPI, cluster types.DBCluster) (*rds.DeleteDBClusterOutput, error) {
	outputCluster, err := client.DeleteDBCluster(
		context.TODO(),
		&rds.DeleteDBClusterInput{
			DBClusterIdentifier: cluster.DBClusterIdentifier,
			SkipFinalSnapshot:   aws.Bool(true),
		},
	)
	if err != nil {
		var dbcnff *types.DBClusterNotFoundFault
		if errors.As(err, &dbcnff) {
			cli.Printf("Cluster %s not found, continuing ...\n", *cluster.DBClusterIdentifier)
			return nil, errDBClusterNotFound
		}
	}
	cli.Printf("Deleting cluster %s\n", *cluster.DBClusterIdentifier)
	return outputCluster, nil
}

func createInstance(client RDSAPI, srcEngine, srcDBSubnetGroup, name string, size string, tags []types.Tag) (*rds.CreateDBInstanceOutput, error) {
	outputInstance, err := client.CreateDBInstance(
		context.TODO(),
		&rds.CreateDBInstanceInput{
			DBClusterIdentifier:  aws.String(name),
			DBInstanceIdentifier: aws.String(name),
			DBInstanceClass:      aws.String(size),
			Engine:               aws.String(srcEngine),
			DBSubnetGroupName:    aws.String(srcDBSubnetGroup),
			Tags:                 tags,
		},
	)
	if err != nil {
		cli.Printf("Error: %s \n", err)
		var dbcnff *types.DBClusterNotFoundFault
		var dbiaef *types.DBInstanceAlreadyExistsFault
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "InvalidParameterCombination" {
				cli.Printf("InvalidParameterCombination: %s\n", apiErr.ErrorMessage())
				errInvalidParameterCombination = fmt.Errorf("InvalidParameterCombination: %s", apiErr.ErrorMessage())

				return nil, errInvalidParameterCombination
			} else {
				cli.Printf("code: %s, message: %s, fault: %s", apiErr.ErrorCode(), apiErr.ErrorMessage(), apiErr.ErrorFault().String())
			}
		}
		if errors.As(err, &dbiaef) {
			cli.Printf("Instance %s already exists\n", name)
			return nil, errDBInstanceAlreadyExists
		}
		if errors.As(err, &dbcnff) {
			cli.Printf("Cluster %s not found, abort operation\n", name)
			return nil, errDBClusterNotFound
		}
	}
	return outputInstance, nil
}

func restoreClusterToPointInTime(client RDSAPI, src types.DBCluster, name string, tags []types.Tag) (*rds.RestoreDBClusterToPointInTimeOutput, error) {
	var securityGroupIds []string
	for _, v := range src.VpcSecurityGroups {
		securityGroupIds = append(securityGroupIds, *v.VpcSecurityGroupId)
	}
	outputCluster, err := client.RestoreDBClusterToPointInTime(
		context.TODO(),
		&rds.RestoreDBClusterToPointInTimeInput{
			DBClusterIdentifier:             aws.String(name),
			SourceDBClusterIdentifier:       src.DBClusterIdentifier,
			DBClusterParameterGroupName:     src.DBClusterParameterGroup,
			DBSubnetGroupName:               src.DBSubnetGroup,
			RestoreType:                     aws.String("copy-on-write"),
			UseLatestRestorableTime:         aws.Bool(true),
			EnableIAMDatabaseAuthentication: aws.Bool(true),
			VpcSecurityGroupIds:             securityGroupIds,
			Tags:                            tags,
		},
	)
	if err != nil {
		var dbcnff *types.DBClusterNotFoundFault
		var dbsnff *types.DBSubnetGroupNotFoundFault
		var dbcpgnf *types.DBClusterParameterGroupNotFoundFault
		var ivpcnsf *types.InvalidVPCNetworkStateFault
		var apiErr smithy.APIError

		if errors.As(err, &dbcnff) {
			return nil, errDBClusterNotFound
		}
		if errors.As(err, &dbsnff) {
			return nil, errDBSubnetGroupNotFound
		}
		if errors.As(err, &dbcpgnf) {
			return nil, errDBParameterGroupNotFound
		}
		if errors.As(err, &ivpcnsf) {
			errInvalidVPCNetworkState = fmt.Errorf("invalid VPC network state: %s", *ivpcnsf.Message)
			return nil, errInvalidVPCNetworkState
		}
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "InvalidParameterValue" {
				errInvalidParameterValue = fmt.Errorf("InvalidParameterCombination: %s", apiErr.ErrorMessage())
				return nil, errInvalidParameterValue
			} else if apiErr.ErrorCode() == "InvalidParameterCombination" {
				errInvalidParameterCombination = fmt.Errorf("InvalidParameterCombination: %s", apiErr.ErrorMessage())
				return nil, errInvalidParameterCombination
			} else {
				cli.Printf("code: %s, message: %s, fault: %s", apiErr.ErrorCode(), apiErr.ErrorMessage(), apiErr.ErrorFault().String())
			}
		}
		return nil, err
	}
	return outputCluster, nil
}

func restoreClusterFromSnapshot(client RDSAPI, src types.DBClusterSnapshot, name string, dbClusterParameterGroup string, subnetGroupName string, tags []types.Tag) (*rds.RestoreDBClusterFromSnapshotOutput, error) {
	outputCluster, err := client.RestoreDBClusterFromSnapshot(
		context.TODO(),
		&rds.RestoreDBClusterFromSnapshotInput{
			DBClusterIdentifier:         aws.String(name),
			SnapshotIdentifier:          src.DBClusterSnapshotIdentifier,
			Engine:                      src.Engine,
			DBClusterParameterGroupName: aws.String(dbClusterParameterGroup),
			DBSubnetGroupName:           aws.String(subnetGroupName),
			Tags:                        tags,
			PubliclyAccessible:          aws.Bool(false),
		},
	)
	if err != nil {
		var dbscsnff *types.DBClusterSnapshotNotFoundFault
		var dbsnff *types.DBSubnetGroupNotFoundFault
		var dbcpgnf *types.DBClusterParameterGroupNotFoundFault
		var ivpcnsf *types.InvalidVPCNetworkStateFault
		var apiErr smithy.APIError

		if errors.As(err, &dbscsnff) {
			return nil, errDBClusterSnapshotNotFound
		}
		if errors.As(err, &dbsnff) {
			return nil, errDBSubnetGroupNotFound
		}
		if errors.As(err, &dbcpgnf) {
			return nil, errDBParameterGroupNotFound
		}
		if errors.As(err, &ivpcnsf) {
			errInvalidVPCNetworkState = fmt.Errorf("invalid VPC network state: %s", *ivpcnsf.Message)
			return nil, errInvalidVPCNetworkState
		}
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "InvalidParameterValue" {
				errInvalidParameterValue = fmt.Errorf("InvalidParameterCombination: %s", apiErr.ErrorMessage())
				return nil, errInvalidParameterValue
			} else if apiErr.ErrorCode() == "InvalidParameterCombination" {
				errInvalidParameterCombination = fmt.Errorf("InvalidParameterCombination: %s", apiErr.ErrorMessage())
				return nil, errInvalidParameterCombination
			} else {
				cli.Printf("code: %s, message: %s, fault: %s", apiErr.ErrorCode(), apiErr.ErrorMessage(), apiErr.ErrorFault().String())
			}
		}
		return nil, err
	}
	return outputCluster, nil
}

func snapshotExists(client RDSAPI, name string) (bool, error) {
	var dbcsnff *types.DBClusterSnapshotNotFoundFault
	_, err := client.DescribeDBClusterSnapshots(
		context.TODO(),
		&rds.DescribeDBClusterSnapshotsInput{
			DBClusterSnapshotIdentifier: aws.String(name),
			IncludeShared:               aws.Bool(true),
		},
	)
	if err != nil {
		if errors.As(err, &dbcsnff) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
