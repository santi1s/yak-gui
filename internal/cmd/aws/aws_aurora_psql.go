package aws

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/teleport"
	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v3"
)

type psqlFlags struct {
	clusterIdentifier  string
	instanceIdentifier string
	endpoint           string
	clusterPort        int
	pgOptions          string
	dbName             string
	dbUser             string
	dbPassword         string
	dbCommand          string
	psqlArgs           string
	clusterRegion      string
	writer             bool
	replication        bool
	reviewers          string
	bypassTsh          bool
}

type ClusterInfo struct {
	ClusterName    string
	Endpoint       string
	ReaderEndpoint string
	DBName         string
}

var (
	auroraPsqlCmd = &cobra.Command{
		Use:    "psql",
		Short:  "Connect in psql on Aurora cluster",
		RunE:   psql,
		PreRun: list,
	}
	describeDBClustersOutput *rds.DescribeDBClustersOutput
	providedPsqlFlags        psqlFlags
	rdsClient                *rds.Client
	cfg                      aws.Config
)

func list(cmd *cobra.Command, args []string) {
	var err error
	cfg, err = config.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	rdsClient = rds.NewFromConfig(cfg)

	if !cmd.Flags().Changed("region") {
		providedPsqlFlags.clusterRegion = cfg.Region
	}

	if !cmd.Flags().Changed("instance-identifier") {
		describeDBClustersOutput, err = rdsClient.DescribeDBClusters(context.TODO(), &rds.DescribeDBClustersInput{})
		if err != nil {
			log.Fatalf("Error listing RDS clusters: %s", err)
		}
		if len(describeDBClustersOutput.DBClusters) == 0 {
			log.Fatal(errors.New("no cluster found for this environment"))
		}

		if !cmd.Flags().Changed("cluster-identifier") {
			cList := clustersExtractInfo(describeDBClustersOutput.DBClusters)
			if len(cList) != 1 {
				yamlDBClusters, err := yaml.Marshal(cList)
				if err != nil {
					log.Fatalf("Error marshaling to YAML: %s", err)
				}
				cli.Println("=== Please select a cluster name with --cluster-identifier parameter ===")
				cli.Println(string(yamlDBClusters))
				os.Exit(1)
			}
			providedPsqlFlags.clusterIdentifier = cList[0].ClusterName
		}
	}
}

func psql(cmd *cobra.Command, args []string) error {
	if cmd.Flags().Changed("cluster-identifier") {
		cluster, err := getCluster(rdsClient, providedPsqlFlags.clusterIdentifier)
		if err != nil {
			log.Fatalf("Error getting cluster: %s", err)
		}

		if providedPsqlFlags.endpoint == "" {
			if !providedPsqlFlags.writer {
				if cli.AskConfirmation("Do you want to connect to the writer endpoint ?") {
					providedPsqlFlags.writer = true
				}
			}
			providedPsqlFlags.endpoint = targetEndpoint(cluster, providedPsqlFlags.writer)
		} else {
			b, err := checkIfEndpointBelongsToCluster(rdsClient, providedPsqlFlags.clusterIdentifier, providedPsqlFlags.endpoint)
			if err != nil {
				log.Fatalf("Error checking if endpoint belongs to cluster: %s", err)
			}
			if !b {
				log.Fatal(fmt.Errorf("endpoint %s does not belong to cluster %s", providedPsqlFlags.endpoint, providedPsqlFlags.clusterIdentifier))
			}
		}

		eStatus, err := getEndpointStatus(rdsClient, providedPsqlFlags.clusterIdentifier, providedPsqlFlags.endpoint)
		if err != nil {
			log.Fatalf("Error getting endpoint status: %s", err)
		}
		if eStatus != rdsEndpointAvailableStatus {
			log.Fatal(fmt.Errorf("endpoint %s is not available, maybe because this is a writer and the active one is not on this AWS region", providedPsqlFlags.endpoint))
		}
	} else if cmd.Flags().Changed("instance-identifier") {
		instance, err := getInstance(rdsClient, providedPsqlFlags.instanceIdentifier)
		if err != nil {
			log.Fatalf("Error getting instance: %s", err)
		}
		providedPsqlFlags.endpoint = *instance.Endpoint.Address
	} else {
		os.Exit(1)
	}

	return psqlExec()
}

func psqlExec() error {
	var err error
	stsClient := sts.NewFromConfig(cfg)

	if providedPsqlFlags.dbCommand != "" {
		providedPsqlFlags.dbCommand = fmt.Sprintf("-c \"%s\"", providedPsqlFlags.dbCommand)
	}

	var replicationOptions string
	if providedPsqlFlags.replication {
		replicationOptions = "replication=database"
	}

	pgPassword, ok := os.LookupEnv("PGPASSWORD")
	if ok && providedPsqlFlags.dbPassword == "" {
		providedPsqlFlags.dbPassword = pgPassword
	}

	var psqlCommand string
	var psqlArgsPrefix string
	if !providedPsqlFlags.bypassTsh && providedPsqlFlags.dbPassword == "" {
		tConfig, err := teleport.ReadTeleportConfig()
		if err != nil {
			log.Fatalf("Error reading Teleport config: %s", err)
		}

		if err := teleport.TshLogin(tConfig); err != nil {
			log.Fatalf("Error on tsh login: %s", err)
			return err
		}

		tshStatus, err := teleport.GetTshStatus()
		if err != nil {
			log.Fatalf("Error getting Teleport active roles: %s", err)
		}

		tResource, err := teleport.GetTeleportDBResourceFromURI(providedPsqlFlags.endpoint + ":" + fmt.Sprint(providedPsqlFlags.clusterPort))
		if err != nil {
			log.Fatalf("Error getting Teleport DB resource from URI: %s", err)
		}

		if !matchStringWithWildcard(providedPsqlFlags.dbUser, tResource.Users.Allowed) {
			awsAccountNo, err := getAWSAccountNo(stsClient)
			if err != nil {
				log.Fatalf("Error getting AWS account number: %s", err)
			}
			for _, c := range tConfig.Accounts[awsAccountNo].Roles {
				if c.Type == teleport.DBConfigType {
					roles := []teleport.Role{c}

					err = teleport.TshRequestNeededRoles(roles, tshStatus.Name, tConfig.Accounts[awsAccountNo].Name, tshStatus.ValidUntil, teleport.ReviewersToSlice(providedPsqlFlags.reviewers), teleport.DBConfigType)
					if err != nil {
						log.Fatalf("Error requesting Teleport privileged access: %s", err)
					}
					break
				}
			}
		}

		if err := teleport.TshDBLogin(tResource.Metadata.Name, providedPsqlFlags.dbUser, providedPsqlFlags.dbName); err != nil {
			log.Fatalf("Could not generate credentials for DB via Teleport: %s", err)
		}

		psqlArgsPrefix, err = teleport.TshGeneratePsqlArgsPrefix(tResource.Metadata.Name)
		if err != nil {
			log.Fatalf("Error generating psql query prefix: %s", err)
		}
	} else {
		if !cli.AskConfirmation("Bypassing tsh can trigger a crisis on production. Continue ?") {
			return nil
		}
		if providedPsqlFlags.dbPassword == "" {
			providedPsqlFlags.dbPassword, err = generateIamDBPassword(&cfg, providedPsqlFlags.endpoint, providedPsqlFlags.clusterPort, endpointRegion(providedPsqlFlags.endpoint), providedPsqlFlags.dbUser)
			if err != nil {
				log.Fatalf("Error generating IAM DB password: %s", err)
			}
		}
		psqlArgsPrefix = fmt.Sprintf("user=%s host=%s port=%d dbname=%s", providedPsqlFlags.dbUser, providedPsqlFlags.endpoint, providedPsqlFlags.clusterPort, providedPsqlFlags.dbName)
	}

	psqlCommand = fmt.Sprintf("psql \"%s %s\" %s %s", psqlArgsPrefix, replicationOptions, providedPsqlFlags.psqlArgs, providedPsqlFlags.dbCommand)
	log.Info("Connecting to Endpoint ", providedPsqlFlags.endpoint)
	cli.Println("# " + psqlCommand)

	psqlCommand = fmt.Sprintf("PGOPTIONS=\"%s\" %s", providedPsqlFlags.pgOptions, psqlCommand)
	if providedPsqlFlags.bypassTsh || providedPsqlFlags.dbPassword != "" {
		psqlCommand = fmt.Sprintf("PGPASSWORD=\"%s\" %s", providedPsqlFlags.dbPassword, psqlCommand)
	}

	pCmd := exec.Command("sh", "-c", psqlCommand)

	pCmd.Stdin = os.Stdin
	pCmd.Stdout = os.Stdout
	pCmd.Stderr = os.Stderr

	return pCmd.Run()
}

func init() {
	auroraPsqlCmd.Flags().IntVarP(&providedPsqlFlags.clusterPort, "port", "p", 5432, "Aurora cluster port")
	auroraPsqlCmd.Flags().StringVarP(&providedPsqlFlags.dbName, "dbname", "d", "doctolib", "Database name")
	auroraPsqlCmd.Flags().StringVarP(&providedPsqlFlags.dbUser, "username", "U", "dba", "Database user")
	auroraPsqlCmd.Flags().StringVarP(&providedPsqlFlags.dbPassword, "password", "W", "", "Database password")
	auroraPsqlCmd.Flags().StringVarP(&providedPsqlFlags.dbCommand, "command", "c", "", "Command to execute")
	auroraPsqlCmd.Flags().StringVarP(&providedPsqlFlags.psqlArgs, "psql-args", "a", "", "Additional psql arguments")
	auroraPsqlCmd.Flags().StringVar(&providedPsqlFlags.clusterRegion, "region", "", "Aurora cluster region")
	auroraPsqlCmd.Flags().StringVarP(&providedPsqlFlags.clusterIdentifier, "cluster-identifier", "i", "", "Aurora cluster identifier (mutually exclusive with instance-identifier)")
	auroraPsqlCmd.Flags().StringVarP(&providedPsqlFlags.instanceIdentifier, "instance-identifier", "I", "", "Aurora instance identifier (mutually exclusive with cluster-identifier)")
	auroraPsqlCmd.Flags().StringVar(&providedPsqlFlags.endpoint, "endpoint", "", "Aurora cluster endpoint")
	auroraPsqlCmd.Flags().BoolVar(&providedPsqlFlags.writer, "writer", false, "Connect to writer endpoint instead of reader endpoint")
	auroraPsqlCmd.Flags().StringVar(&providedPsqlFlags.pgOptions, "pg-options", "-c statement_timeout=5s", "Additional psql options")
	auroraPsqlCmd.Flags().BoolVarP(&providedPsqlFlags.replication, "replication", "R", false, "replication connection (add replication=database to psql command)")
	auroraPsqlCmd.Flags().StringVarP(&providedPsqlFlags.reviewers, "reviewers", "r", "", "override default reviewers of the request (comma separated)")
	auroraPsqlCmd.Flags().BoolVar(&providedPsqlFlags.bypassTsh, "bypass-tsh", false, "disable connection with tsh (it can trigger crisis on production)")
	auroraPsqlCmd.MarkFlagsMutuallyExclusive("cluster-identifier", "instance-identifier")
	auroraPsqlCmd.MarkFlagsMutuallyExclusive("endpoint", "instance-identifier")
}
