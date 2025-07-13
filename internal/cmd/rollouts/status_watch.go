package rollouts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type WatchStatusOptions struct {
	Watch    bool
	Timeout  time.Duration
	Interval time.Duration
}

// watchRolloutStatus watches a rollout for status changes until completion or timeout
func watchRolloutStatus(namespace, rolloutName string, options WatchStatusOptions) error {
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

	if !options.Watch {
		// Single status check
		return getSingleRolloutStatus(dynamicClient, rolloutGVR, namespace, rolloutName)
	}

	// Watch mode
	return watchRolloutStatusLoop(dynamicClient, rolloutGVR, namespace, rolloutName, options)
}

func getSingleRolloutStatus(client dynamic.Interface, gvr schema.GroupVersionResource, namespace, rolloutName string) error {
	rollout, err := client.Resource(gvr).Namespace(namespace).Get(context.Background(), rolloutName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("rollout %s not found: %s", rolloutName, err)
	}

	status := buildRolloutStatus(rollout)
	message := formatStatusMessage(status)
	cli.Printf("%s\n", message)
	return nil
}

func watchRolloutStatusLoop(client dynamic.Interface, gvr schema.GroupVersionResource, namespace, rolloutName string, options WatchStatusOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup timeout if specified
	if options.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	var lastMessage string
	ticker := time.NewTicker(options.Interval)
	defer ticker.Stop()

	// Print initial status
	rollout, err := client.Resource(gvr).Namespace(namespace).Get(ctx, rolloutName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("rollout %s not found: %s", rolloutName, err)
	}

	status := buildRolloutStatus(rollout)
	lastMessage = printStatusUpdate(status, lastMessage)

	// Check if already in terminal state
	if isTerminalState(status.Status) {
		return checkFinalStatus(status)
	}

	// Watch loop
	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("rollout status watch exceeded timeout")
			}
			return ctx.Err()

		case <-ticker.C:
			rollout, err := client.Resource(gvr).Namespace(namespace).Get(ctx, rolloutName, metav1.GetOptions{})
			if err != nil {
				cli.Printf("Error fetching rollout: %s\n", err)
				continue
			}

			status := buildRolloutStatus(rollout)
			lastMessage = printStatusUpdate(status, lastMessage)

			if isTerminalState(status.Status) {
				return checkFinalStatus(status)
			}
		}
	}
}

func printStatusUpdate(status statusMap, lastMessage string) string {
	message := formatStatusMessage(status)

	if message != lastMessage {
		cli.Printf("%s\n", message)
	}
	return message
}

func formatStatusMessage(status statusMap) string {
	message := status.Status

	if status.Message != "<none>" && status.Message != "" {
		message = fmt.Sprintf("%s - %s", status.Status, status.Message)
	}

	// Add revision info if available
	var suffixes []string
	if status.Revision != "" {
		suffixes = append(suffixes, status.Revision)
	}
	if status.Analysis != "" {
		suffixes = append(suffixes, status.Analysis)
	}

	if len(suffixes) > 0 {
		message = fmt.Sprintf("%s (%s)", message, strings.Join(suffixes, ", "))
	}

	return message
}

func isTerminalState(status string) bool {
	return status == "Healthy" || status == "Degraded"
}

func checkFinalStatus(status statusMap) error {
	if status.Status == "Degraded" {
		return fmt.Errorf("the rollout is in a degraded state with message: %s", status.Message)
	}
	return nil
}
