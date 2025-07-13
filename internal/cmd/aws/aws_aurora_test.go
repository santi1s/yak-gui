package aws

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/smithy-go"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/stretchr/testify/assert"
)

type mockRDSClient struct {
	RDSAPI
	dbClustersOutput                    *rds.DescribeDBClustersOutput
	describeDBInstancesOutput           *rds.DescribeDBInstancesOutput
	deleteDBInstanceOutput              *rds.DeleteDBInstanceOutput
	deleteDBClusterOutput               *rds.DeleteDBClusterOutput
	createDBInstanceOutput              *rds.CreateDBInstanceOutput
	restoreDBClusterToPointInTimeOutput *rds.RestoreDBClusterToPointInTimeOutput
	restoreDBClusterFromSnapshotOutput  *rds.RestoreDBClusterFromSnapshotOutput
	describeDBClusterSnapshotOutput     *rds.DescribeDBClusterSnapshotsOutput
	err                                 error
}

func (m mockRDSClient) DescribeDBClusters(context.Context, *rds.DescribeDBClustersInput, ...func(*rds.Options)) (*rds.DescribeDBClustersOutput, error) {
	return m.dbClustersOutput, m.err
}

func (m mockRDSClient) DescribeDBInstances(context.Context, *rds.DescribeDBInstancesInput, ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	return m.describeDBInstancesOutput, m.err
}

func (m mockRDSClient) DeleteDBInstance(context.Context, *rds.DeleteDBInstanceInput, ...func(*rds.Options)) (*rds.DeleteDBInstanceOutput, error) {
	return m.deleteDBInstanceOutput, m.err
}

func (m mockRDSClient) DeleteDBCluster(context.Context, *rds.DeleteDBClusterInput, ...func(*rds.Options)) (*rds.DeleteDBClusterOutput, error) {
	return m.deleteDBClusterOutput, m.err
}

func (m mockRDSClient) CreateDBInstance(context.Context, *rds.CreateDBInstanceInput, ...func(*rds.Options)) (*rds.CreateDBInstanceOutput, error) {
	return m.createDBInstanceOutput, m.err
}

func (m mockRDSClient) RestoreDBClusterToPointInTime(context.Context, *rds.RestoreDBClusterToPointInTimeInput, ...func(*rds.Options)) (*rds.RestoreDBClusterToPointInTimeOutput, error) {
	return m.restoreDBClusterToPointInTimeOutput, m.err
}

func (m mockRDSClient) RestoreDBClusterFromSnapshot(context.Context, *rds.RestoreDBClusterFromSnapshotInput, ...func(*rds.Options)) (*rds.RestoreDBClusterFromSnapshotOutput, error) {
	return m.restoreDBClusterFromSnapshotOutput, m.err
}

func (m mockRDSClient) DescribeDBClusterSnapshots(context.Context, *rds.DescribeDBClusterSnapshotsInput, ...func(*rds.Options)) (*rds.DescribeDBClusterSnapshotsOutput, error) {
	return m.describeDBClusterSnapshotOutput, m.err
}

/*
This test multiple functions that use the same client method describeDBClusters
*/
func TestDescribeCluster(t *testing.T) {
	scenarios := []struct {
		name           string
		expectedError  error
		outputClusters []types.DBCluster
		expectedResult bool
	}{
		{
			name: "cluster exists",
			outputClusters: []types.DBCluster{
				{
					DBClusterIdentifier: aws.String("test"),
				},
			},
			expectedResult: true,
		},
		{
			name:           "cluster does not exist",
			outputClusters: nil,
			expectedResult: false,
		},
		{
			name: "cluster available",
			outputClusters: []types.DBCluster{
				{
					DBClusterIdentifier: aws.String("test"),
					Status:              aws.String("available"),
				},
			},
			expectedResult: true,
		},
		{
			name: "cluster not available",
			outputClusters: []types.DBCluster{
				{
					DBClusterIdentifier: aws.String("test"),
					Status:              aws.String("creating"),
				},
			},
			expectedResult: false,
		},
	}

	for _, tt := range scenarios {
		c := mockRDSClient{
			dbClustersOutput: &rds.DescribeDBClustersOutput{DBClusters: tt.outputClusters},
		}
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.outputClusters, describeCluster(c, "test"))
			if strings.Contains(tt.name, "exist") {
				assert.Equal(t, tt.expectedResult, clusterExist(c, "test"))
			}
			if strings.Contains(tt.name, "available") {
				assert.Equal(t, tt.expectedResult, clusterAvailable(c, "test"))
			}
		})
	}
}

/*
This test multiple functions that use the same client method describeDBInstances
*/
func TestDescribeInstance(t *testing.T) {
	scenarios := []struct {
		name            string
		expectedError   error
		outputInstances []types.DBInstance
		expectedResult  bool
	}{
		{
			name: "instance exists",
			outputInstances: []types.DBInstance{
				{
					DBInstanceIdentifier: aws.String("test"),
				},
			},
			expectedResult: true,
		},
		{
			name:            "instance does not exist",
			outputInstances: nil,
			expectedResult:  false,
		},
		{
			name: "instance available",
			outputInstances: []types.DBInstance{
				{
					DBInstanceIdentifier: aws.String("test"),
					DBInstanceStatus:     aws.String("available"),
				},
			},
			expectedResult: true,
		},
		{
			name: "instance not available",
			outputInstances: []types.DBInstance{
				{
					DBInstanceIdentifier: aws.String("test"),
					DBInstanceStatus:     aws.String("creating"),
				},
			},
			expectedResult: false,
		},
	}

	for _, tt := range scenarios {
		c := mockRDSClient{
			describeDBInstancesOutput: &rds.DescribeDBInstancesOutput{DBInstances: tt.outputInstances},
		}
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.outputInstances, describeInstance(c, "test"))
			if strings.Contains(tt.name, "exist") {
				assert.Equal(t, tt.expectedResult, instanceExist(c, "test"))
			}
			if strings.Contains(tt.name, "available") {
				assert.Equal(t, tt.expectedResult, instanceAvailable(c, "test"))
			}
		})
	}
}

func TestDiscoverClone(t *testing.T) {
	scenarios := []struct {
		name           string
		expectedError  error
		outputClusters []types.DBCluster
		expectedResult []types.DBCluster
	}{
		{
			name: "clone found",
			outputClusters: []types.DBCluster{
				{
					DBClusterIdentifier: aws.String("test"),
				},
			},
			expectedResult: []types.DBCluster{
				{
					DBClusterIdentifier: aws.String("test"),
				},
			},
		},
		{
			name: "multiple clones found",
			outputClusters: []types.DBCluster{
				{
					DBClusterIdentifier: aws.String("test"),
				},
				{
					DBClusterIdentifier: aws.String("test_2"),
				},
			},
			expectedResult: []types.DBCluster{
				{
					DBClusterIdentifier: aws.String("test"),
				},
				{
					DBClusterIdentifier: aws.String("test_2"),
				},
			},
		},
		{
			name:           "no clone found",
			outputClusters: []types.DBCluster{},
			expectedResult: nil,
		},
	}

	for _, tt := range scenarios {
		c := mockRDSClient{
			dbClustersOutput: &rds.DescribeDBClustersOutput{DBClusters: tt.outputClusters},
		}
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedResult, discoverClone(c, []string{"test"}))
		})
	}
}

func TestDiscoverSrcCluster(t *testing.T) {
	scenarios := []struct {
		name           string
		input          []string
		inputAPIErr    error
		outputClusters []types.DBCluster
		expectedResult []types.DBCluster
		expectedErr    error
	}{
		{
			name:  "src found",
			input: []string{"test"},
			outputClusters: []types.DBCluster{
				{
					DBClusterIdentifier: aws.String("test"),
				},
			},
			expectedResult: []types.DBCluster{
				{
					DBClusterIdentifier: aws.String("test"),
				},
			},
		},
		{
			name:  "multiple src found",
			input: []string{"test", "test_2"},
			outputClusters: []types.DBCluster{
				{
					DBClusterIdentifier: aws.String("test"),
				},
				{
					DBClusterIdentifier: aws.String("test_2"),
				},
			},
			expectedResult: []types.DBCluster{
				{
					DBClusterIdentifier: aws.String("test"),
				},
				{
					DBClusterIdentifier: aws.String("test_2"),
				},
			},
		},
		{
			name:           "no src found",
			input:          []string{"test"},
			outputClusters: []types.DBCluster{},
			expectedResult: nil,
			expectedErr:    errNoDBClusterSrcFound,
		},
	}

	for _, tt := range scenarios {
		c := mockRDSClient{
			dbClustersOutput: &rds.DescribeDBClustersOutput{DBClusters: tt.outputClusters},
		}
		t.Run(tt.name, func(t *testing.T) {
			src, err := discoverSrc(c, tt.input)
			if err != nil {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)
				}
			}
			assert.Equal(t, tt.expectedResult, src)
		})
	}
}

func TestDeleteInstance(t *testing.T) {
	scenarios := []struct {
		name, input   string
		output        *rds.DeleteDBInstanceOutput
		expectedError error
	}{
		{
			name:  "instance deleted",
			input: "test",
			output: &rds.DeleteDBInstanceOutput{
				DBInstance: &types.DBInstance{
					DBInstanceIdentifier: aws.String("test"),
				},
			},
			expectedError: nil,
		},
		{
			name:          "instance not found",
			input:         "test_2",
			output:        &rds.DeleteDBInstanceOutput{},
			expectedError: errDBInstanceNotFound,
		},
	}

	for _, tt := range scenarios {
		c := mockRDSClient{
			deleteDBInstanceOutput: tt.output,
		}
		t.Run(tt.name, func(t *testing.T) {
			output, err := deleteInstance(c, tt.input)
			if err != nil {
				assert.Error(t, err, errDBInstanceNotFound)
			}
			assert.Equal(t, tt.output, output)
		})
	}
}

func TestDeleteCluster(t *testing.T) {
	scenarios := []struct {
		name          string
		input         types.DBCluster
		output        *rds.DeleteDBClusterOutput
		expectedError error
		stdout        string
	}{
		{
			name:  "cluster deleted",
			input: types.DBCluster{DBClusterIdentifier: aws.String("test")},
			output: &rds.DeleteDBClusterOutput{
				DBCluster: &types.DBCluster{
					DBClusterIdentifier: aws.String("test"),
				},
			},
			expectedError: nil,
		},
		{
			name:          "cluster not found",
			input:         types.DBCluster{DBClusterIdentifier: aws.String("test_2")},
			output:        &rds.DeleteDBClusterOutput{},
			expectedError: errDBClusterNotFound,
		},
	}

	for _, tt := range scenarios {
		c := mockRDSClient{
			deleteDBClusterOutput: tt.output,
		}
		t.Run(tt.name, func(t *testing.T) {
			output, err := deleteCluster(c, tt.input)
			if err != nil {
				assert.Error(t, err, errDBClusterNotFound)
			}
			assert.Equal(t, tt.output, output)
		})
	}
}

func TestCreateInstance(t *testing.T) {
	scenarios := []struct {
		name        string
		input       map[string]interface{}
		inputAPIErr error
		output      *rds.CreateDBInstanceOutput
		expectedErr error
	}{
		{
			name: "invalid engine",
			input: map[string]interface{}{
				"clusterID":   "test_instance_invalid_engine",
				"size":        "db.t2.micro",
				"tags":        []types.Tag{},
				"engine":      "test_invalid_engine",
				"subnetGroup": "",
			},
			inputAPIErr: &smithy.GenericAPIError{
				Code:    "InvalidParameterCombination",
				Message: "The engine name requested for your DB instance (aurora) doesn't match the engine name of your DB cluster (test_invalid_engine).",
				Fault:   smithy.ErrorFault(0),
			},
			output:      nil,
			expectedErr: fmt.Errorf("invalidParameterCombination: %s", "The engine name requested for your DB instance (aurora) doesn't match the engine name of your DB cluster (test_invalid_engine)."),
		},
		{
			name: "instance created",
			input: map[string]interface{}{
				"clusterID": "test_instance_created",
				"size":      "db.t2.micro",
				"tags": []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("test"),
					},
				},
				"engine":      "",
				"subnetGroup": "",
			},
			inputAPIErr: nil,
			output: &rds.CreateDBInstanceOutput{
				DBInstance: (*types.DBInstance)(nil),
			},
			expectedErr: nil,
		},
		{
			name: "instance already exists",
			input: map[string]interface{}{
				"clusterID":   "test_instance_exists",
				"size":        "db.t2.micro",
				"tags":        []types.Tag{},
				"engine":      "",
				"subnetGroup": "",
			},
			inputAPIErr: &types.DBInstanceAlreadyExistsFault{},
			output:      nil,
			expectedErr: errDBInstanceAlreadyExists,
		},
		{
			name: "cluster not found",
			input: map[string]interface{}{
				"clusterID": "test_instance_exists",
				"size":      "db.t2.micro",
				"tags": []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("test"),
					},
				},
				"engine":      "",
				"subnetGroup": "",
			},
			inputAPIErr: &types.DBClusterNotFoundFault{},
			output:      nil,
			expectedErr: errDBClusterNotFound,
		},
	}
	for _, tt := range scenarios {
		c := mockRDSClient{
			createDBInstanceOutput: tt.output,
			err:                    tt.inputAPIErr,
		}
		t.Run(tt.name, func(t *testing.T) {
			output, err := createInstance(c, tt.input["engine"].(string), tt.input["subnetGroup"].(string), tt.input["clusterID"].(string), tt.input["size"].(string), tt.input["tags"].([]types.Tag))
			if err != nil {
				assert.Nil(t, output)
				assert.Error(t, err)
			}
			assert.Equal(t, tt.output, output)
		})
	}
}

func TestRestoreClusterToPointInTime(t *testing.T) {
	scenarios := []struct {
		name, input     string
		inputAPIErr     error
		inputSrcCluster types.DBCluster
		output          *rds.RestoreDBClusterToPointInTimeOutput
		expectedErr     error
	}{
		{
			name:  "cluster restored",
			input: "test_restored",
			inputSrcCluster: types.DBCluster{
				DBClusterIdentifier:     aws.String("test_src"),
				DBClusterParameterGroup: aws.String("test_src"),
				DBSubnetGroup:           aws.String("test_src"),
				VpcSecurityGroups:       []types.VpcSecurityGroupMembership{{VpcSecurityGroupId: aws.String("test_sg")}},
			},
			inputAPIErr: nil,
			output: &rds.RestoreDBClusterToPointInTimeOutput{
				DBCluster: &types.DBCluster{
					DBClusterIdentifier: aws.String("test_restored"),
					VpcSecurityGroups:   []types.VpcSecurityGroupMembership{{VpcSecurityGroupId: aws.String("test_sg")}},
				},
			},
			expectedErr: nil,
		},
		{
			name:  "parameter group not found",
			input: "test_restored",
			inputSrcCluster: types.DBCluster{
				DBClusterIdentifier:     aws.String("test_src"),
				DBClusterParameterGroup: aws.String("test_na_pg"),
				DBSubnetGroup:           aws.String("test_src"),
			},
			inputAPIErr: &smithy.GenericAPIError{
				Code:    "DBClusterParameterGroupNotFoundFault",
				Message: "DBClusterParameterGroup not found: test_na_pg",
				Fault:   smithy.ErrorFault(0),
			},
			output:      nil,
			expectedErr: errDBParameterGroupNotFound,
		},
		{
			name:  "subnet group not found",
			input: "test_restored",
			inputSrcCluster: types.DBCluster{
				DBClusterIdentifier:     aws.String("test_src"),
				DBClusterParameterGroup: aws.String("test_src"),
				DBSubnetGroup:           aws.String("test_na_pg"),
			},
			inputAPIErr: &smithy.GenericAPIError{
				Code:    "DBSubnetGroupNotFoundFault",
				Message: "DB subnet group 'test_na_pg' does not exist.",
				Fault:   smithy.ErrorFault(0),
			},
			output:      nil,
			expectedErr: errDBSubnetGroupNotFound,
		},
		{
			name:  "source cluster not found",
			input: "test_restored",
			inputSrcCluster: types.DBCluster{
				DBClusterIdentifier:     aws.String("test_na_pg"),
				DBClusterParameterGroup: aws.String("test_src"),
				DBSubnetGroup:           aws.String("test_src"),
			},
			inputAPIErr: &smithy.GenericAPIError{
				Code:    "DBClusterNotFoundFault",
				Message: "DBClusterNotFoundFault: The source cluster could not be found or cannot be accessed: test_na_pg",
				Fault:   smithy.ErrorFault(0),
			},
			output:      nil,
			expectedErr: errDBClusterNotFound,
		},
		{
			name:  "invalid network state",
			input: "test_restored",
			inputSrcCluster: types.DBCluster{
				DBClusterIdentifier:     aws.String("test_na_pg"),
				DBClusterParameterGroup: aws.String("test_src"),
				DBSubnetGroup:           aws.String("error_sg"),
			},
			inputAPIErr: &smithy.GenericAPIError{
				Code:    "InvalidVPCNetworkStateFault",
				Message: "InvalidVPCNetworkStateFault: The DB subnet group doesn't meet Availability Zone (AZ) coverage requirement.",
				Fault:   smithy.ErrorFault(0),
			},
			output:      nil,
			expectedErr: fmt.Errorf("InvalidNetworkState: %s", "The DB subnet group doesn't meet Availability Zone (AZ) coverage requirement."),
		},
		{
			name:  "invalid parameter combination",
			input: "test_restored",
			inputSrcCluster: types.DBCluster{
				DBClusterIdentifier:     aws.String("test_src"),
				DBClusterParameterGroup: aws.String("error_pg"),
				DBSubnetGroup:           aws.String("test_src"),
			},
			inputAPIErr: &smithy.GenericAPIError{
				Code:    "InvalidParameterCombination",
				Message: "InvalidParameterCombination: The Parameter Group error_pg with DBParameterGroupFamily aurora-postgresql12 cannot be used for this instance. Please use a Parameter Group with DBParameterGroupFamily aurora-postgresql15",
				Fault:   smithy.ErrorFault(0),
			},
			output:      nil,
			expectedErr: fmt.Errorf("InvalidParameterCombination: %s", "The Parameter Group error_pg with DBParameterGroupFamily aurora-postgresql12 cannot be used for this instance. Please use a Parameter Group with DBParameterGroupFamily aurora-postgresql15"),
		},
		{
			name:  "invalid parameter value",
			input: "test_restored",
			inputSrcCluster: types.DBCluster{
				DBClusterIdentifier:     aws.String("test_na_pg"),
				DBClusterParameterGroup: aws.String("test_src"),
				DBSubnetGroup:           aws.String("test_src"),
			},
			inputAPIErr: &smithy.GenericAPIError{
				Code:    "InvalidParameterValue",
				Message: "InvalidParameterValue: The parameter DBClusterIdentifier is not a valid identifier. Identifiers must begin with a letter; must contain only ASCII letters, digits, and hyphens; and must not end with a hyphen or contain two consecutive hyphens.",
				Fault:   smithy.ErrorFault(0),
			},
			output:      nil,
			expectedErr: fmt.Errorf("InvalidParameterValue: %s", " The parameter DBClusterIdentifier is not a valid identifier. Identifiers must begin with a letter; must contain only ASCII letters, digits, and hyphens; and must not end with a hyphen or contain two consecutive hyphens."),
		},
	}
	for _, tt := range scenarios {
		c := mockRDSClient{
			restoreDBClusterToPointInTimeOutput: tt.output,
			err:                                 tt.inputAPIErr,
		}
		t.Run(tt.name, func(t *testing.T) {
			output, err := restoreClusterToPointInTime(c, tt.inputSrcCluster, tt.input, []types.Tag{})
			if err != nil {
				assert.Nil(t, output)
				assert.Error(t, err, tt.expectedErr)
			}
			assert.Equal(t, tt.output, output)
		})
	}
}

func TestSnapshotExists(t *testing.T) {
	scenarios := []struct {
		name              string
		inputSnapshotName string
		inputAPIError     error
		output            bool
		outputSnapshots   []types.DBClusterSnapshot
		expectedError     error
		stdout            string
	}{
		{
			name:              "snapshot exists",
			inputSnapshotName: "test",
			inputAPIError:     nil,
			output:            true,
			outputSnapshots: []types.DBClusterSnapshot{
				{
					DBClusterSnapshotIdentifier: aws.String("test"),
				},
				{
					DBClusterSnapshotIdentifier: aws.String("something_else"),
				},
			},
			expectedError: nil,
		},
		{
			name:              "snapshot does not exist",
			inputSnapshotName: "test",
			inputAPIError:     errDBClusterSnapshotNotFound,
			output:            false,
			outputSnapshots: []types.DBClusterSnapshot{
				{
					DBClusterSnapshotIdentifier: aws.String("whatever"),
				},
				{
					DBClusterSnapshotIdentifier: aws.String("another_one"),
				},
			},
			expectedError: nil,
		},
		{
			name:              "some error but not errDBClusterSnapshotNotFound",
			inputSnapshotName: "test",
			inputAPIError:     errors.New("some error"),
			output:            false,
			outputSnapshots: []types.DBClusterSnapshot{
				{
					DBClusterSnapshotIdentifier: aws.String("test"),
				},
				{
					DBClusterSnapshotIdentifier: aws.String("something_else"),
				},
			},
			expectedError: errors.New("some error"),
		},
	}

	for _, tt := range scenarios {
		c := mockRDSClient{
			describeDBClusterSnapshotOutput: &rds.DescribeDBClusterSnapshotsOutput{
				DBClusterSnapshots: tt.outputSnapshots,
			},
			err: tt.inputAPIError,
		}
		t.Run(tt.name, func(t *testing.T) {
			output, err := snapshotExists(c, tt.inputSnapshotName)
			if tt.expectedError != nil || err != nil {
				assert.Error(t, err, tt.expectedError)
			}
			assert.Equal(t, tt.output, output)
		})
	}
}

func TestRestoreClusterFromSnapshot(t *testing.T) {
	scenarios := []struct {
		name, input      string
		inputAPIErr      error
		inputSrcSnapshot types.DBClusterSnapshot
		output           *rds.RestoreDBClusterFromSnapshotOutput
		expectedErr      error
	}{
		{
			name:  "cluster restored",
			input: "test_restored",
			inputSrcSnapshot: types.DBClusterSnapshot{
				DBClusterSnapshotIdentifier: aws.String("test_snapshot"),
				Engine:                      aws.String("test_engine"),
			},
			inputAPIErr: nil,
			output: &rds.RestoreDBClusterFromSnapshotOutput{
				DBCluster: &types.DBCluster{
					DBClusterIdentifier: aws.String("test_restored"),
				},
			},
			expectedErr: nil,
		},
		{
			name:  "snapshot not found",
			input: "test_restored",
			inputSrcSnapshot: types.DBClusterSnapshot{
				DBClusterSnapshotIdentifier: aws.String("test_na_snapshot"),
				Engine:                      aws.String("test_engine"),
			},
			inputAPIErr: &smithy.GenericAPIError{
				Code:    "DBClusterSnapshotNotFoundFault",
				Message: "DBClusterSnapshot not found: test_na_snapshot",
				Fault:   smithy.ErrorFault(0),
			},
			output:      nil,
			expectedErr: errDBClusterSnapshotNotFound,
		},
		{
			name:  "parameter group not found",
			input: "test_restored",
			inputSrcSnapshot: types.DBClusterSnapshot{
				DBClusterSnapshotIdentifier: aws.String("test_snapshot"),
				Engine:                      aws.String("test_engine"),
			},
			inputAPIErr: &smithy.GenericAPIError{
				Code:    "DBClusterParameterGroupNotFoundFault",
				Message: "DBClusterParameterGroup not found: test_na_pg",
				Fault:   smithy.ErrorFault(0),
			},
			output:      nil,
			expectedErr: errDBParameterGroupNotFound,
		},
		{
			name:  "subnet group not found",
			input: "test_restored",
			inputSrcSnapshot: types.DBClusterSnapshot{
				DBClusterSnapshotIdentifier: aws.String("test_snapshot"),
				Engine:                      aws.String("test_engine"),
			},
			inputAPIErr: &smithy.GenericAPIError{
				Code:    "DBSubnetGroupNotFoundFault",
				Message: "DB subnet group 'test_na_sg' does not exist.",
				Fault:   smithy.ErrorFault(0),
			},
			output:      nil,
			expectedErr: errDBSubnetGroupNotFound,
		},
		{
			name:  "invalid network state",
			input: "test_restored",
			inputSrcSnapshot: types.DBClusterSnapshot{
				DBClusterSnapshotIdentifier: aws.String("test_snapshot"),
				Engine:                      aws.String("test_engine"),
			},
			inputAPIErr: &smithy.GenericAPIError{
				Code:    "InvalidVPCNetworkStateFault",
				Message: "InvalidVPCNetworkStateFault: The DB subnet group doesn't meet Availability Zone (AZ) coverage requirement.",
				Fault:   smithy.ErrorFault(0),
			},
			output:      nil,
			expectedErr: fmt.Errorf("InvalidNetworkState: %s", "The DB subnet group doesn't meet Availability Zone (AZ) coverage requirement."),
		},
		{
			name:  "invalid parameter combination",
			input: "test_restored",
			inputSrcSnapshot: types.DBClusterSnapshot{
				DBClusterSnapshotIdentifier: aws.String("test_snapshot"),
				Engine:                      aws.String("test_engine"),
			},
			inputAPIErr: &smithy.GenericAPIError{
				Code:    "InvalidParameterCombination",
				Message: "InvalidParameterCombination: The Parameter Group error_pg with DBParameterGroupFamily aurora-postgresql12 cannot be used for this instance. Please use a Parameter Group with DBParameterGroupFamily aurora-postgresql15",
				Fault:   smithy.ErrorFault(0),
			},
			output:      nil,
			expectedErr: fmt.Errorf("InvalidParameterCombination: %s", "The Parameter Group error_pg with DBParameterGroupFamily aurora-postgresql12 cannot be used for this instance. Please use a Parameter Group with DBParameterGroupFamily aurora-postgresql15"),
		},
		{
			name:  "invalid parameter value",
			input: "test_restored",
			inputSrcSnapshot: types.DBClusterSnapshot{
				DBClusterSnapshotIdentifier: aws.String("test_snapshot"),
				Engine:                      aws.String("test_engine"),
			},
			inputAPIErr: &smithy.GenericAPIError{
				Code:    "InvalidParameterValue",
				Message: "InvalidParameterValue: The parameter DBClusterIdentifier is not a valid identifier. Identifiers must begin with a letter; must contain only ASCII letters, digits, and hyphens; and must not end with a hyphen or contain two consecutive hyphens.",
				Fault:   smithy.ErrorFault(0),
			},
			output:      nil,
			expectedErr: fmt.Errorf("InvalidParameterValue: %s", " The parameter DBClusterIdentifier is not a valid identifier. Identifiers must begin with a letter; must contain only ASCII letters, digits, and hyphens; and must not end with a hyphen or contain two consecutive hyphens."),
		},
	}
	for _, tt := range scenarios {
		c := mockRDSClient{
			restoreDBClusterFromSnapshotOutput: tt.output,
			err:                                tt.inputAPIErr,
		}
		t.Run(tt.name, func(t *testing.T) {
			output, err := restoreClusterFromSnapshot(c, tt.inputSrcSnapshot, tt.input, "", "", []types.Tag{})
			if err != nil {
				assert.Nil(t, output)
				assert.Error(t, err, tt.expectedErr)
			}
			assert.Equal(t, tt.output, output)
		})
	}
}
