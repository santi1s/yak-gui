package rollouts

import (
	"context"
	"fmt"
	"strconv"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var providedUndoFlags RolloutsUndoFlags

type RolloutsUndoFlags struct {
	rollout    string
	toRevision int64
}

func undo(cmd *cobra.Command, args []string) error {
	rolloutName := providedUndoFlags.rollout
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

	if providedUndoFlags.toRevision > 0 {
		// Rollback to specific revision
		err = rollbackToRevision(dynamicClient, rollout, providedUndoFlags.toRevision)
		if err != nil {
			return fmt.Errorf("failed to rollback rollout %s to revision %d: %s", rolloutName, providedUndoFlags.toRevision, err)
		}
		cli.Printf("Rollout %s rolled back to revision %d successfully\n", rolloutName, providedUndoFlags.toRevision)
	} else {
		// Rollback to previous revision
		err = rollbackToPrevious(dynamicClient, rollout)
		if err != nil {
			return fmt.Errorf("failed to rollback rollout %s: %s", rolloutName, err)
		}
		cli.Printf("Rollout %s rolled back to previous revision successfully\n", rolloutName)
	}

	return nil
}

func rollbackToRevision(dynamicClient dynamic.Interface, rollout *unstructured.Unstructured, revision int64) error {
	// Get ReplicaSets to find the target revision
	rsGVR := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "replicasets",
	}

	// Get the selector from the rollout
	selector, found, err := unstructured.NestedStringMap(rollout.Object, "spec", "selector", "matchLabels")
	if err != nil || !found {
		return fmt.Errorf("unable to get selector from rollout")
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

	rsList, err := dynamicClient.Resource(rsGVR).Namespace(rollout.GetNamespace()).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("failed to list ReplicaSets: %s", err)
	}

	// Find the ReplicaSet with the target revision
	var targetRS *unstructured.Unstructured
	for _, rs := range rsList.Items {
		annotations := rs.GetAnnotations()
		if annotations == nil {
			continue
		}

		revisionStr, exists := annotations["deployment.kubernetes.io/revision"]
		if !exists {
			continue
		}

		rsRevision, err := strconv.ParseInt(revisionStr, 10, 64)
		if err != nil {
			continue
		}

		if rsRevision == revision {
			targetRS = &rs
			break
		}
	}

	if targetRS == nil {
		return fmt.Errorf("revision %d not found", revision)
	}

	// Get the pod template from the target ReplicaSet
	targetTemplate, found, err := unstructured.NestedMap(targetRS.Object, "spec", "template")
	if err != nil || !found {
		return fmt.Errorf("unable to get template from target ReplicaSet")
	}

	// Update the rollout template
	if err := unstructured.SetNestedMap(rollout.Object, targetTemplate, "spec", "template"); err != nil {
		return fmt.Errorf("failed to update rollout template: %s", err)
	}

	// Update the rollout
	rolloutGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "rollouts",
	}

	_, err = dynamicClient.Resource(rolloutGVR).Namespace(rollout.GetNamespace()).Update(context.Background(), rollout, metav1.UpdateOptions{})
	return err
}

func rollbackToPrevious(dynamicClient dynamic.Interface, rollout *unstructured.Unstructured) error {
	// For simplicity, we'll use the undo annotation which tells Argo Rollouts to rollback
	annotations := rollout.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations["rollout.argoproj.io/undo"] = "true"
	rollout.SetAnnotations(annotations)

	// Update the rollout
	rolloutGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "rollouts",
	}

	_, err := dynamicClient.Resource(rolloutGVR).Namespace(rollout.GetNamespace()).Update(context.Background(), rollout, metav1.UpdateOptions{})
	return err
}

var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Rollback rollout to previous revision",
	Example: `yak rollouts undo -r my-rollout
yak rollouts undo -r my-rollout --to-revision 3`,
	RunE: undo,
}

func init() {
	undoCmd.Flags().StringVarP(&providedUndoFlags.rollout, "rollout", "r", "", "Rollout name (required)")
	undoCmd.Flags().Int64Var(&providedUndoFlags.toRevision, "to-revision", 0, "Rollback to specific revision (optional)")
	_ = undoCmd.MarkFlagRequired("rollout")
}
