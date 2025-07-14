package aws

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/santi1s/yak/cli"
	"github.com/spf13/cobra"
)

var (
	auroraCloneDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "aurora clone management",
		RunE:  deleteClone,
	}
)

func init() {}

func deleteClone(cmd *cobra.Command, args []string) error {
	client := initAwsRdsClient()

	cloneClusterList := discoverClone(client, cloneProvidedFlags.targets)

	var wg sync.WaitGroup
	wg.Add(len(cloneClusterList))

	for _, clone := range cloneClusterList {
		cli.Printf("%s cluster selected for deletion\n", *clone.DBClusterIdentifier)
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
		}(clone)
	}
	wg.Wait()

	return nil
}
