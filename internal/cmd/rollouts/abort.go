package rollouts

import (
	"context"
	"fmt"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

var providedAbortFlags RolloutsAbortFlags

type RolloutsAbortFlags struct {
	rollout string
}

func abort(cmd *cobra.Command, args []string) error {
	rolloutName := providedAbortFlags.rollout
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

	// Check current rollout status
	phase, _, _ := unstructured.NestedString(rollout.Object, "status", "phase")
	cli.Printf("Current rollout phase: %s\n", phase)

	// Check if abort is already set in status
	abortStatus, _, _ := unstructured.NestedBool(rollout.Object, "status", "abort")
	if abortStatus {
		cli.Printf("Abort already set in status: %t\n", abortStatus)
	}

	// The abort operation works by setting {"status":{"abort":true}} in the rollout's status
	// This matches the behavior of the official kubectl argo rollouts plugin
	// When the controller sees this status field, it will:
	// 1. Stop the current rollout progression
	// 2. Scale down the new revision
	// 3. Scale up the stable revision (rollback)

	// Create the patch to set abort status
	patchData := []byte(`{"status":{"abort":true}}`)

	// Apply the patch using JSON merge patch
	_, err = dynamicClient.Resource(rolloutGVR).Namespace(resolveNamespace()).Patch(
		context.Background(),
		rolloutName,
		types.MergePatchType,
		patchData,
		metav1.PatchOptions{},
		"status", // Use status subresource
	)
	if err != nil {
		// If status subresource fails, try without it (fallback)
		_, err = dynamicClient.Resource(rolloutGVR).Namespace(resolveNamespace()).Patch(
			context.Background(),
			rolloutName,
			types.MergePatchType,
			patchData,
			metav1.PatchOptions{},
		)
		if err != nil {
			return fmt.Errorf("failed to abort rollout %s: %s", rolloutName, err)
		}
	}

	cli.Printf("Rollout %s abort initiated. The controller will now rollback to the stable version.\n", rolloutName)
	cli.Printf("Use 'yak rollouts status -r %s' to monitor the rollback progress.\n", rolloutName)
	return nil
}

var abortCmd = &cobra.Command{
	Use:     "abort",
	Short:   "Abort a rollout and rollback to stable version",
	Example: `yak rollouts abort -r my-rollout`,
	RunE:    abort,
}

func init() {
	abortCmd.Flags().StringVarP(&providedAbortFlags.rollout, "rollout", "r", "", "Rollout name (required)")
	_ = abortCmd.MarkFlagRequired("rollout")
}
