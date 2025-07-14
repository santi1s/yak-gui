package rollouts

import (
	"context"
	"fmt"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var providedPauseFlags RolloutsPauseFlags

type RolloutsPauseFlags struct {
	rollout string
}

func pause(cmd *cobra.Command, args []string) error {
	rolloutName := providedPauseFlags.rollout
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

	// Check if rollout is already paused
	if paused, found, _ := unstructured.NestedBool(rollout.Object, "spec", "paused"); found && paused {
		cli.Printf("Rollout %s is already paused\n", rolloutName)
		return nil
	}

	// Set paused to true
	if err := unstructured.SetNestedField(rollout.Object, true, "spec", "paused"); err != nil {
		return fmt.Errorf("failed to set paused field: %s", err)
	}

	// Update the rollout
	_, err = dynamicClient.Resource(rolloutGVR).Namespace(resolveNamespace()).Update(context.Background(), rollout, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to pause rollout %s: %s", rolloutName, err)
	}

	cli.Printf("Rollout %s paused successfully\n", rolloutName)
	return nil
}

var pauseCmd = &cobra.Command{
	Use:     "pause",
	Short:   "Pause a rollout",
	Example: `yak rollouts pause -r my-rollout`,
	RunE:    pause,
}

func init() {
	pauseCmd.Flags().StringVarP(&providedPauseFlags.rollout, "rollout", "r", "", "Rollout name (required)")
	_ = pauseCmd.MarkFlagRequired("rollout")
}
