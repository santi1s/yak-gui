package couchbase

import (
	"github.com/doctolib/yak/internal/constant"
	"github.com/doctolib/yak/internal/helper"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	kubernetes "k8s.io/client-go/kubernetes"
)

type couchbaseFlags struct {
	kubeNamespace string
}

var providedCouchbaseFlags couchbaseFlags

var (
	providedFlags couchbaseFlags

	couchbaseCmd = &cobra.Command{
		Use:   "couchbase",
		Short: "tools to manage couchbase clusters",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Root().Name() == constant.CliName && cmd.Root().PersistentPreRun != nil {
				cmd.Root().PersistentPreRun(cmd, args)
			}

			if cmd.HasParent() && cmd.Parent().Name() == "completion" {
				switch cmd.Name() {
				case "bash", "zsh", "fish", "powershell":
					cmd.ResetFlags()
				}
			}

			cmd.SilenceUsage = true
		},
	}
)

func GetRootCmd() *cobra.Command {
	return couchbaseCmd
}

// findCouchbaseClusterPods fetches all couchbase pods inside the current kubernetes cluster
// It returns the list of matching pods
func findCouchbaseClusterPods(clientset *kubernetes.Clientset, kubeNamespace string) ([]v1.Pod, error) {
	pods, err := helper.GetPodsMatchingLabel(clientset, kubeNamespace, "couchbase_server", "true")
	return pods, err
}

// findSgwPods fetches all sync gateway pods inside the current kubernetes cluster
// It returns the list of matching pods
func findSgwPods(clientset *kubernetes.Clientset, kubeNamespace string) ([]v1.Pod, error) {
	pods, err := helper.GetPodsMatchingLabel(clientset, kubeNamespace, "app", "sync-gateway")
	return pods, err
}

func init() {
	couchbaseCmd.PersistentFlags().StringVarP(&providedCouchbaseFlags.kubeNamespace, "kubenamespace", "n", "couchbase", "kubernetes namespace where couchbase pods are running")
	couchbaseCmd.AddCommand(collectLogsCmd)
}
