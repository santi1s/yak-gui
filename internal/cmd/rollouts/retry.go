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

var providedRetryFlags RolloutsRetryFlags

type RolloutsRetryFlags struct {
	rollout string
}

func retry(cmd *cobra.Command, args []string) error {
	rolloutName := providedRetryFlags.rollout
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
	abortStatus, _, _ := unstructured.NestedBool(rollout.Object, "status", "abort")

	cli.Printf("Current rollout phase: %s\n", phase)
	cli.Printf("Current abort status: %t\n", abortStatus)

	// Retry is specifically designed for aborted rollouts
	if !abortStatus {
		cli.Printf("Warning: Rollout %s is not aborted (abort=false), retry may not be necessary\n", rolloutName)
	}

	// The retry operation works by setting {"status":{"abort":false}} in the rollout's status
	// This matches the behavior of the official kubectl argo rollouts plugin
	// This clears the abort status and allows the rollout to continue

	// Create the patch to clear abort status
	patchData := []byte(`{"status":{"abort":false}}`)

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
			return fmt.Errorf("failed to retry rollout %s: %s", rolloutName, err)
		}
	}

	cli.Printf("Rollout %s retry initiated successfully. The rollout will continue from where it was aborted.\n", rolloutName)
	return nil
}

var retryCmd = &cobra.Command{
	Use:     "retry",
	Short:   "Retry a failed rollout step",
	Example: `yak rollouts retry -r my-rollout`,
	RunE:    retry,
}

func init() {
	retryCmd.Flags().StringVarP(&providedRetryFlags.rollout, "rollout", "r", "", "Rollout name (required)")
	_ = retryCmd.MarkFlagRequired("rollout")
}
