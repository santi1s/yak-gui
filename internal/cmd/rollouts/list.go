package rollouts

import (
	"context"
	"fmt"
	"sort"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var providedListFlags RolloutsListFlags

type RolloutsListFlags struct {
	all bool
}

type rolloutListItem struct {
	Name      string
	Namespace string
	Status    string
	Replicas  string
	Age       string
	Strategy  string
}

func list(cmd *cobra.Command, args []string) error {
	config, _, err := helper.InitKubeClusterConfig()
	if err != nil {
		return fmt.Errorf("unable to connect to Kubernetes cluster: %s", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to create dynamic client: %s", err)
	}

	rolloutGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "rollouts",
	}

	var rollouts *unstructured.UnstructuredList
	if providedListFlags.all {
		rollouts, err = dynamicClient.Resource(rolloutGVR).List(context.Background(), metav1.ListOptions{})
	} else {
		rollouts, err = dynamicClient.Resource(rolloutGVR).Namespace(resolveNamespace()).List(context.Background(), metav1.ListOptions{})
	}

	if err != nil {
		return fmt.Errorf("failed to list rollouts: %s", err)
	}

	if providedFlags.json || providedFlags.yaml {
		return formatOutput(rollouts)
	} else {
		formatRolloutList(rollouts)
	}
	return nil
}

func formatRolloutList(rollouts *unstructured.UnstructuredList) {
	const nameLabel = "NAME"
	const namespaceLabel = "NAMESPACE"
	const statusLabel = "STATUS"
	const replicasLabel = "REPLICAS"
	const ageLabel = "AGE"
	const strategyLabel = "STRATEGY"

	var nameWidth = len(nameLabel)
	var namespaceWidth = len(namespaceLabel)
	var statusWidth = len(statusLabel)
	var replicasWidth = len(replicasLabel)
	var ageWidth = len(ageLabel)
	var strategyWidth = len(strategyLabel)

	items := []rolloutListItem{}
	for _, rollout := range rollouts.Items {
		item := buildRolloutListItem(&rollout)
		items = append(items, item)

		nameWidth = max(nameWidth, len(item.Name))
		namespaceWidth = max(namespaceWidth, len(item.Namespace))
		statusWidth = max(statusWidth, len(item.Status))
		replicasWidth = max(replicasWidth, len(item.Replicas))
		ageWidth = max(ageWidth, len(item.Age))
		strategyWidth = max(strategyWidth, len(item.Strategy))
	}

	// Sort by namespace, then by name
	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace != items[j].Namespace {
			return items[i].Namespace < items[j].Namespace
		}
		return items[i].Name < items[j].Name
	})

	nameWidth++
	namespaceWidth++
	statusWidth++
	replicasWidth++
	ageWidth++
	strategyWidth++

	if providedListFlags.all {
		cli.Printf("%-*s %-*s %-*s %-*s %-*s %s\n",
			nameWidth, nameLabel,
			namespaceWidth, namespaceLabel,
			statusWidth, statusLabel,
			replicasWidth, replicasLabel,
			ageWidth, ageLabel,
			strategyLabel)

		for _, item := range items {
			cli.Printf("%-*s %-*s %-*s %-*s %-*s %s\n",
				nameWidth, item.Name,
				namespaceWidth, item.Namespace,
				statusWidth, item.Status,
				replicasWidth, item.Replicas,
				ageWidth, item.Age,
				item.Strategy)
		}
	} else {
		cli.Printf("%-*s %-*s %-*s %-*s %s\n",
			nameWidth, nameLabel,
			statusWidth, statusLabel,
			replicasWidth, replicasLabel,
			ageWidth, ageLabel,
			strategyLabel)

		for _, item := range items {
			cli.Printf("%-*s %-*s %-*s %-*s %s\n",
				nameWidth, item.Name,
				statusWidth, item.Status,
				replicasWidth, item.Replicas,
				ageWidth, item.Age,
				item.Strategy)
		}
	}
}

func buildRolloutListItem(rollout *unstructured.Unstructured) rolloutListItem {
	name := rollout.GetName()
	namespace := rollout.GetNamespace()

	// Get spec replicas
	specReplicas, _, _ := unstructured.NestedInt64(rollout.Object, "spec", "replicas")
	statusReplicas, _, _ := unstructured.NestedInt64(rollout.Object, "status", "replicas")

	// Get phase/status
	phase, _, _ := unstructured.NestedString(rollout.Object, "status", "phase")
	if phase == "" {
		phase = "Unknown"
	}

	// Get strategy
	strategy := "Unknown"
	if canarySpec, found, _ := unstructured.NestedMap(rollout.Object, "spec", "strategy", "canary"); found && canarySpec != nil {
		strategy = "Canary"
	} else if blueGreenSpec, found, _ := unstructured.NestedMap(rollout.Object, "spec", "strategy", "blueGreen"); found && blueGreenSpec != nil {
		strategy = "BlueGreen"
	}

	// Calculate age
	age := "Unknown"
	if creationTimestamp := rollout.GetCreationTimestamp(); !creationTimestamp.IsZero() {
		age = helper.GetAge(creationTimestamp.Time)
	}

	return rolloutListItem{
		Name:      name,
		Namespace: namespace,
		Status:    phase,
		Replicas:  fmt.Sprintf("%d/%d", statusReplicas, specReplicas),
		Age:       age,
		Strategy:  strategy,
	}
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List rollouts",
	Example: `yak rollouts list
yak rollouts list --all`,
	RunE: list,
}

func init() {
	listCmd.Flags().BoolVar(&providedListFlags.all, "all", false, "List rollouts from all namespaces")
}
