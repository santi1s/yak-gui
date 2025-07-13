package rollouts

import (
	"context"
	"fmt"
	"time"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

var providedRestartFlags RolloutsRestartFlags

type RolloutsRestartFlags struct {
	rollout string
}

func restart(cmd *cobra.Command, args []string) error {
	rolloutName := providedRestartFlags.rollout
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

	// Verify the rollout exists
	_, err = dynamicClient.Resource(rolloutGVR).Namespace(resolveNamespace()).Get(context.Background(), rolloutName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("rollout %s not found in namespace %s: %s", rolloutName, resolveNamespace(), err)
	}

	// The restart operation works by setting {"spec":{"restartAt":"timestamp"}} in the rollout's spec
	// This matches the behavior of the official kubectl argo rollouts plugin
	// This triggers a restart of the rollout pods

	// Create timestamp for restart
	restartAt := time.Now().UTC().Format(time.RFC3339)

	// Create the patch to set restartAt
	patchData := []byte(fmt.Sprintf(`{"spec":{"restartAt":"%s"}}`, restartAt))

	// Apply the patch using JSON merge patch
	_, err = dynamicClient.Resource(rolloutGVR).Namespace(resolveNamespace()).Patch(
		context.Background(),
		rolloutName,
		types.MergePatchType,
		patchData,
		metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to restart rollout %s: %s", rolloutName, err)
	}

	cli.Printf("Rollout %s restarted successfully\n", rolloutName)
	return nil
}

var restartCmd = &cobra.Command{
	Use:     "restart",
	Short:   "Restart the pods of a rollout",
	Example: `yak rollouts restart -r my-rollout`,
	RunE:    restart,
}

func init() {
	restartCmd.Flags().StringVarP(&providedRestartFlags.rollout, "rollout", "r", "", "Rollout name (required)")
	_ = restartCmd.MarkFlagRequired("rollout")
}
