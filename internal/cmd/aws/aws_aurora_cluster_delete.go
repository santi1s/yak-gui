package aws

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/doctolib/yak/cli"
	"github.com/spf13/cobra"
)

var (
	auroraClusterDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "aurora cluster management",
		RunE:  deleteDBCluster,
	}
)

func init() {}

func deleteDBCluster(cmd *cobra.Command, args []string) error {
	client := initAwsRdsClient()

	dbClusterList := []types.DBCluster{}
	for _, t := range clusterProvidedFlags.targets {
		dbClusterList = append(dbClusterList, describeCluster(client, t)...)
	}

	var wg sync.WaitGroup
	wg.Add(len(dbClusterList))

	for _, c := range dbClusterList {
		cli.Printf("%s cluster selected for deletion\n", *c.DBClusterIdentifier)
		go func(cluster types.DBCluster) {
			defer wg.Done()

			_, err := deleteInstance(client, *cluster.DBClusterIdentifier)
			if err != nil {
				cli.Printf("Error while deleting instance %s: %s \n", *cluster.DBClusterIdentifier, err)
				panic(err)
			}

			_, err = deleteCluster(client, cluster)
			if err != nil {
				cli.Printf("Error while deleting cluster %s: %s \n", *cluster.DBClusterIdentifier, err)
				panic(err)
			}

			for clusterExist(client, *cluster.DBClusterIdentifier) {
				cli.Printf("Waiting for cluster %s to be deleted\n", *cluster.DBClusterIdentifier)
				time.Sleep(5 * time.Second)
			}
			cli.Printf("Cluster %s deleted with instance %s \n", *cluster.DBClusterIdentifier, *cluster.DBClusterIdentifier)
		}(c)
	}
	wg.Wait()

	return nil
}
