package aws

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/santi1s/yak/cli"
	"github.com/spf13/cobra"
)

var (
	auroraCloneCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "aurora clone management",
		Long: `Create a clone of an Aurora cluster.
    --source/-s flag is mandatory.
    If --from-snapshot flag is set, the command will create a clone from a snapshots
    and --source flags are snapshot names which must exist within the AWS account or be shared with it.
    If --from-snapshot flag is not set, the command will create a clone from a cluster
    and --source flags are cluster names which must exist within the AWS account or be shared with it.
    `,
		RunE: createClone,
	}
)

func init() {}

func createClone(cmd *cobra.Command, args []string) error {
	client := initAwsRdsClient()

	if len(cloneProvidedFlags.sources) == 0 {
		_, _ = cli.PrintfErr("No source provided with --source/-s flag")
		return nil
	}

	if len(cloneProvidedFlags.targets) == 0 {
		_, _ = cli.PrintfErr("No target name provided with --target flag")
		return nil
	}

	if len(cloneProvidedFlags.sources) != len(cloneProvidedFlags.targets) {
		_, _ = cli.PrintfErr("Number of sources and targets must be the same")
		return nil
	}

	if cloneProvidedFlags.fromSnapshot {
		return createCloneFromSnapshot(client)
	} else {
		return createCloneFromCluster(client)
	}
}

func createCloneFromCluster(client RDSAPI) error {
	srcClusterList, err := discoverSrc(client, cloneProvidedFlags.sources)
	if err != nil {
		cli.Printf("Error while looking for source cluster %s: %s \n", cloneProvidedFlags.sources, err)
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(srcClusterList))
	startTime := time.Now()

	for index, src := range srcClusterList {
		clusterID := cloneProvidedFlags.targets[index]
		go func(src types.DBCluster, targetClusterID string) {
			defer wg.Done()

			awsTags := fmtAWSTags(cloneProvidedFlags.tags)
			for !clusterExist(client, targetClusterID) {
				outputCluster, err := restoreClusterToPointInTime(client, src, targetClusterID, awsTags)
				time.Sleep(time.Second)
				cli.Printf("[%ds] Waiting for cluster %s to be created\n", getSecondSinceTime(startTime), targetClusterID)
				if err != nil {
					_, _ = cli.PrintfErr("Error: %s \n", err)
					continue
				}
				if outputCluster == nil {
					cli.Printf("[%ds] outputCluster is nil\n", getSecondSinceTime(startTime))
					continue
				}
			}
			cli.Printf("[%ds] Cluster %s created \n", getSecondSinceTime(startTime), targetClusterID)

			var instanceID string
			for !instanceExist(client, targetClusterID) {
				outputInstance, err := createInstance(client, *src.Engine, *src.DBSubnetGroup, targetClusterID, cloneProvidedFlags.size, awsTags)
				time.Sleep(time.Second)
				if err != nil {
					_, _ = cli.PrintfErr("Error: %s \n", err)
					continue
				}
				if outputInstance == nil {
					cli.Printf("[%ds] outputInstance is nil\n", getSecondSinceTime(startTime))
					continue
				}

				instanceID = *outputInstance.DBInstance.DBInstanceIdentifier
				cli.Printf("[%ds] Instance %s created in cluster %s \n", getSecondSinceTime(startTime), instanceID, targetClusterID)
				for !instanceAvailable(client, instanceID) {
					cli.Printf("[%ds] Waiting for instance %s to be available\n", getSecondSinceTime(startTime), instanceID)
					time.Sleep(30 * time.Second)
				}
			}
			for !clusterAvailable(client, targetClusterID) {
				cli.Printf("[%ds] Waiting for cluster %s to be available\n", getSecondSinceTime(startTime), targetClusterID)
				time.Sleep(30 * time.Second)
			}
			cli.Printf("[%ds] Instance %s is available in cluster %s \n", getSecondSinceTime(startTime), instanceID, targetClusterID)
		}(src, clusterID)
	}
	wg.Wait()

	return nil
}

func createCloneFromSnapshot(client RDSAPI) error {
	allSourcesExist := true
	for _, src := range cloneProvidedFlags.sources {
		exists, err := snapshotExists(client, src)
		if err != nil {
			_, _ = cli.PrintfErr("There was an error checking for snapshot %s existence :\n%s\n", src, err)
			return err
		}
		if !exists {
			_, _ = cli.PrintfErr("Snapshot %s not found\n", src)
			allSourcesExist = false
		}
	}
	if !allSourcesExist {
		_, _ = cli.PrintfErr("Some snapshots were not found, aborting\n")
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(len(cloneProvidedFlags.sources))
	startTime := time.Now()

	for index, src := range cloneProvidedFlags.sources {
		clusterID := cloneProvidedFlags.targets[index]
		dbClusterParameterGroup := cloneProvidedFlags.dbClusterParameterGroup[index]
		go func(src string, targetClusterID string, dbClusterParameterGroup string) {
			defer wg.Done()

			snapshots := describeClusterSnapshot(client, src)
			if snapshots == nil {
				_, _ = cli.PrintfErr("Error: snapshot %s not found\n", src)
				return
			}
			if len(snapshots) > 1 {
				_, _ = cli.PrintfErr("Error: multiple snapshots found for %s\n", src)
				return
			}

			snapshot := snapshots[0]

			awsTags := fmtAWSTags(cloneProvidedFlags.tags)

			for !clusterExist(client, targetClusterID) {
				outputCluster, err := restoreClusterFromSnapshot(client, snapshot, targetClusterID, dbClusterParameterGroup, cloneProvidedFlags.subnetGroupName, awsTags)
				time.Sleep(time.Second)
				if err != nil {
					_, _ = cli.PrintfErr("Error: %s \n", err)
					continue
				}
				if outputCluster == nil {
					_, _ = cli.Printf("[%ds] outputCluster is nil\n", getSecondSinceTime(startTime))
					continue
				}
				targetClusterID = *outputCluster.DBCluster.DBClusterIdentifier
			}

			createdCluster := describeCluster(client, targetClusterID)
			if createdCluster == nil {
				_, _ = cli.PrintErr("Error: created cluster not found\n")
				return
			}
			_, _ = cli.Printf("[%ds] Cluster %s created \n", getSecondSinceTime(startTime), *createdCluster[0].DBClusterIdentifier)

			for !instanceExist(client, targetClusterID) {
				outputInstance, err := createInstance(client, *snapshot.Engine, *createdCluster[0].DBSubnetGroup, targetClusterID, cloneProvidedFlags.size, awsTags)
				time.Sleep(time.Second)
				if err != nil {
					_, _ = cli.PrintfErr("Error: %s \n", err)
					continue
				}
				if outputInstance == nil {
					_, _ = cli.Printf("[%ds] outputInstance is nil\n", getSecondSinceTime(startTime))
					continue
				}
				instanceID := *outputInstance.DBInstance.DBInstanceIdentifier
				_, _ = cli.Printf("[%ds] Instance %s created in cluster %s \n", getSecondSinceTime(startTime), instanceID, targetClusterID)
				for !instanceAvailable(client, instanceID) {
					_, _ = cli.Printf("[%ds] Waiting for instance %s to be available\n", getSecondSinceTime(startTime), instanceID)
					time.Sleep(30 * time.Second)
				}
			}
			for !clusterAvailable(client, targetClusterID) {
				_, _ = cli.Printf("[%ds] Waiting for cluster %s to be available\n", getSecondSinceTime(startTime), targetClusterID)
				time.Sleep(30 * time.Second)
			}
			cli.Printf("[%ds] Instance %s is available in cluster %s \n", getSecondSinceTime(startTime), targetClusterID, targetClusterID)
		}(src, clusterID, dbClusterParameterGroup)
	}
	wg.Wait()

	return nil
}
