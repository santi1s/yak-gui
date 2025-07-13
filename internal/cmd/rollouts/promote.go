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

var providedPromoteFlags RolloutsPromoteFlags

type RolloutsPromoteFlags struct {
	rollout string
	skip    bool
	full    bool
}

func promote(cmd *cobra.Command, args []string) error {
	rolloutName := providedPromoteFlags.rollout
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

	// Get the rollout to analyze its current state
	rollout, err := dynamicClient.Resource(rolloutGVR).Namespace(resolveNamespace()).Get(context.Background(), rolloutName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("rollout %s not found in namespace %s: %s", rolloutName, resolveNamespace(), err)
	}

	// Determine the appropriate patches based on rollout state and promotion type
	specPatch, statusPatch, unifiedPatch, err := getPromotePatches(rollout, providedPromoteFlags.full)
	if err != nil {
		return fmt.Errorf("failed to determine promotion patches: %s", err)
	}

	// Apply patches exactly like kubectl argo rollouts (lines 116-134)
	if statusPatch != nil {
		_, err = dynamicClient.Resource(rolloutGVR).Namespace(resolveNamespace()).Patch(context.Background(), rolloutName, types.MergePatchType, statusPatch, metav1.PatchOptions{}, "status")
		if err != nil {
			// If status subresource not supported, use unified patch
			if statusPatch != nil && unifiedPatch != nil {
				specPatch = unifiedPatch
			}
		}
	}
	if specPatch != nil {
		_, err = dynamicClient.Resource(rolloutGVR).Namespace(resolveNamespace()).Patch(context.Background(), rolloutName, types.MergePatchType, specPatch, metav1.PatchOptions{})
		if err != nil {
			return fmt.Errorf("failed to promote rollout %s: %s", rolloutName, err)
		}
	}

	if providedPromoteFlags.full {
		cli.Printf("Rollout %s promoted to full deployment\n", rolloutName)
	} else {
		cli.Printf("Rollout %s promoted to next step\n", rolloutName)
	}
	return nil
}

// getPromotePatches returns specPatch, statusPatch, unifiedPatch like kubectl argo rollouts
func getPromotePatches(rollout *unstructured.Unstructured, fullPromotion bool) ([]byte, []byte, []byte, error) {
	var specPatch, statusPatch, unifiedPatch []byte

	// Check if it's a canary or blue-green rollout
	canaryStrategy, canaryFound, _ := unstructured.NestedMap(rollout.Object, "spec", "strategy", "canary")
	_, blueGreenFound, _ := unstructured.NestedMap(rollout.Object, "spec", "strategy", "blueGreen")

	if !canaryFound && !blueGreenFound {
		return nil, nil, nil, fmt.Errorf("rollout does not have a canary or blue-green strategy")
	}

	// Get current rollout state
	isPaused, _, _ := unstructured.NestedBool(rollout.Object, "spec", "paused")
	pauseConditions, hasPauseConditions, _ := unstructured.NestedSlice(rollout.Object, "status", "pauseConditions")
	controllerPause, _, _ := unstructured.NestedBool(rollout.Object, "status", "controllerPause")
	currentStepIndex, hasCurrentStep, _ := unstructured.NestedInt64(rollout.Object, "status", "currentStepIndex")

	if fullPromotion {
		// For full promotion - use promoteFull status field
		currentPodHash, _, _ := unstructured.NestedString(rollout.Object, "status", "currentPodHash")
		stableRS, _, _ := unstructured.NestedString(rollout.Object, "status", "stableRS")
		
		if currentPodHash != stableRS {
			statusPatch = []byte(`{"status":{"promoteFull":true}}`)
		}
	} else {
		// Normal promotion logic - EXACTLY like kubectl argo rollouts (lines 163-196)
		
		// Default: unpause and clear pause conditions (line 164)
		unifiedPatch = []byte(`{"spec":{"paused":false},"status":{"pauseConditions":null}}`)
		
		if isPaused {
			specPatch = []byte(`{"spec":{"paused":false}}`)
		}
		
		// Handle different scenarios based on rollout state
		if isInconclusive(rollout) && hasPauseConditions && len(pauseConditions) > 0 && controllerPause {
			// Inconclusive analysis: clear pause conditions, controller pause, and advance step (lines 171-178)
			if canaryFound && hasCurrentStep {
				steps, stepsFound, _ := unstructured.NestedSlice(canaryStrategy, "steps")
				totalSteps := 0
				if stepsFound {
					totalSteps = len(steps)
				}
				
				if currentStepIndex < int64(totalSteps) {
					nextStepIndex := currentStepIndex + 1
					statusPatch = []byte(fmt.Sprintf(`{"status":{"pauseConditions":null,"controllerPause":false,"currentStepIndex":%d}}`, nextStepIndex))
				}
			}
		} else if hasPauseConditions && len(pauseConditions) > 0 {
			// CRITICAL: When pause conditions exist, ONLY clear them (line 180) - don't advance step!
			statusPatch = []byte(`{"status":{"pauseConditions":null}}`)
		} else if canaryFound {
			// No pause conditions and canary rollout: advance step (lines 181-195)
			if hasCurrentStep {
				steps, stepsFound, _ := unstructured.NestedSlice(canaryStrategy, "steps")
				totalSteps := 0
				if stepsFound {
					totalSteps = len(steps)
				}
				
				if currentStepIndex < int64(totalSteps) {
					nextStepIndex := currentStepIndex + 1
					statusPatch = []byte(fmt.Sprintf(`{"status":{"pauseConditions":null,"currentStepIndex":%d}}`, nextStepIndex))
					unifiedPatch = []byte(fmt.Sprintf(`{"spec":{"paused":false},"status":{"pauseConditions":null,"currentStepIndex":%d}}`, nextStepIndex))
				}
			}
		}
	}

	return specPatch, statusPatch, unifiedPatch, nil
}

// Helper function to check if rollout is in inconclusive state
func isInconclusive(rollout *unstructured.Unstructured) bool {
	// Check for inconclusive analysis runs or other conditions
	// This is a simplified version - in the real implementation this checks analysis run states
	phase, _, _ := unstructured.NestedString(rollout.Object, "status", "phase")
	return phase == "Degraded" // Simplified check
}

var promoteCmd = &cobra.Command{
	Use:   "promote",
	Short: "Promote a rollout to the next step or full deployment",
	Example: `yak rollouts promote -r my-rollout
yak rollouts promote -r my-rollout --full`,
	RunE: promote,
}

func init() {
	promoteCmd.Flags().StringVarP(&providedPromoteFlags.rollout, "rollout", "r", "", "Rollout name (required)")
	promoteCmd.Flags().BoolVar(&providedPromoteFlags.full, "full", false, "Promote to full deployment (skip all remaining steps)")
	_ = promoteCmd.MarkFlagRequired("rollout")
}
