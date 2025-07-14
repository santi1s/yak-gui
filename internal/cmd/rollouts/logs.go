package rollouts

import (
	"context"
	"fmt"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var providedLogsFlags RolloutsLogsFlags

type RolloutsLogsFlags struct {
	rollout   string
	follow    bool
	previous  bool
	container string
	tail      int64
	replicas  int
}

func logs(cmd *cobra.Command, args []string) error {
	rolloutName := providedLogsFlags.rollout
	if rolloutName == "" {
		return fmt.Errorf("rollout name is required")
	}

	config, _, err := helper.InitKubeClusterConfig()
	if err != nil {
		return fmt.Errorf("unable to connect to Kubernetes cluster: %s", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to create dynamic client: %s", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to create Kubernetes clientset: %s", err)
	}

	rolloutGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "rollouts",
	}

	// Get the rollout
	rollout, err := dynamicClient.Resource(rolloutGVR).Namespace(resolveNamespace()).Get(context.Background(), rolloutName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("rollout %s not found in namespace %s: %s", rolloutName, resolveNamespace(), err)
	}

	// Get pods associated with the rollout
	pods, err := getPodsForRollout(clientset, rollout)
	if err != nil {
		return fmt.Errorf("failed to get pods for rollout %s: %s", rolloutName, err)
	}

	if len(pods) == 0 {
		cli.Printf("No pods found for rollout %s\n", rolloutName)
		return nil
	}

	// Limit the number of pods to show logs from
	maxReplicas := providedLogsFlags.replicas
	if maxReplicas > len(pods) {
		maxReplicas = len(pods)
	}

	// When using --follow, limit to 1 replica to avoid blocking
	if providedLogsFlags.follow && maxReplicas > 1 {
		maxReplicas = 1
		cli.Printf("Note: --follow flag limits output to 1 replica to avoid blocking\n")
	}

	cli.Printf("Showing logs from %d out of %d pods\n\n", maxReplicas, len(pods))

	// Get logs from the limited number of pods
	for i := 0; i < maxReplicas; i++ {
		pod := pods[i]

		// Determine which container to use
		containerName := providedLogsFlags.container
		var containerInfo string

		if containerName == "" {
			// No container specified, check if pod has multiple containers
			if len(pod.Spec.Containers) > 1 {
				// Use first container as default for multi-container pods
				containerName = pod.Spec.Containers[0].Name
				containerInfo = fmt.Sprintf(" [defaulting to container: %s]", containerName)
				if i == 0 { // Show info only for first pod to avoid repetition
					cli.Printf("Note: Multiple containers detected, using first container '%s'\n", containerName)
				}
			}
			// For single container pods, leave containerName empty (kubernetes will handle it)
		} else {
			containerInfo = fmt.Sprintf(" [container: %s]", containerName)
		}

		cli.Printf("=== Logs from pod %s%s (%d/%d) ===\n", pod.Name, containerInfo, i+1, maxReplicas)

		logOptions := &corev1.PodLogOptions{
			Follow:    providedLogsFlags.follow,
			Previous:  providedLogsFlags.previous,
			Container: containerName,
		}

		if providedLogsFlags.tail > 0 {
			logOptions.TailLines = &providedLogsFlags.tail
		}

		req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)
		logs, err := req.DoRaw(context.Background())
		if err != nil {
			cli.Printf("Error getting logs for pod %s: %s\n", pod.Name, err)
			continue
		}

		cli.Printf("%s\n", string(logs))
	}

	return nil
}

func getPodsForRollout(clientset *kubernetes.Clientset, rollout *unstructured.Unstructured) ([]corev1.Pod, error) {
	// Get the selector from the rollout
	selector, found, err := unstructured.NestedStringMap(rollout.Object, "spec", "selector", "matchLabels")
	if err != nil || !found {
		return nil, fmt.Errorf("unable to get selector from rollout")
	}

	// Convert map to label selector string
	var selectorParts []string
	for key, value := range selector {
		selectorParts = append(selectorParts, fmt.Sprintf("%s=%s", key, value))
	}
	labelSelector := ""
	if len(selectorParts) > 0 {
		labelSelector = selectorParts[0]
		for _, part := range selectorParts[1:] {
			labelSelector += "," + part
		}
	}

	// List pods with the selector
	podList, err := clientset.CoreV1().Pods(rollout.GetNamespace()).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	return podList.Items, nil
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get logs from rollout pods",
	Example: `yak rollouts logs -r my-rollout
yak rollouts logs -r my-rollout --follow
yak rollouts logs -r my-rollout --tail 100
yak rollouts logs -r my-rollout --replicas 5`,
	RunE: logs,
}

func init() {
	logsCmd.Flags().StringVarP(&providedLogsFlags.rollout, "rollout", "r", "", "Rollout name (required)")
	logsCmd.Flags().BoolVarP(&providedLogsFlags.follow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().BoolVar(&providedLogsFlags.previous, "previous", false, "Show logs from previous container instance")
	logsCmd.Flags().StringVarP(&providedLogsFlags.container, "container", "c", "", "Container name (for multi-container pods)")
	logsCmd.Flags().Int64Var(&providedLogsFlags.tail, "tail", 0, "Number of lines to show from the end of the logs")
	logsCmd.Flags().IntVar(&providedLogsFlags.replicas, "replicas", 3, "Number of replicas to show logs from")
	_ = logsCmd.MarkFlagRequired("rollout")
}
