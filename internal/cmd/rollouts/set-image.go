package rollouts

import (
	"context"
	"fmt"
	"strings"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var providedSetImageFlags RolloutsSetImageFlags

type RolloutsSetImageFlags struct {
	rollout   string
	container string
	image     string
}

func setImage(cmd *cobra.Command, args []string) error {
	rolloutName := providedSetImageFlags.rollout
	if rolloutName == "" {
		return fmt.Errorf("rollout name is required")
	}

	if providedSetImageFlags.image == "" {
		return fmt.Errorf("image is required")
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

	// Get containers from the rollout template
	containers, found, err := unstructured.NestedSlice(rollout.Object, "spec", "template", "spec", "containers")
	if err != nil || !found {
		return fmt.Errorf("unable to get containers from rollout template: %s", err)
	}

	// Update the image for the specified container (or first container if not specified)
	var updated bool
	for i, container := range containers {
		containerMap, ok := container.(map[string]interface{})
		if !ok {
			continue
		}

		containerName, _ := containerMap["name"].(string)

		// If container name is specified, match it; otherwise update the first container
		if providedSetImageFlags.container == "" || containerName == providedSetImageFlags.container {
			containerMap["image"] = providedSetImageFlags.image
			containers[i] = containerMap
			updated = true

			cli.Printf("Updated container %s image to %s\n", containerName, providedSetImageFlags.image)

			// If we specified a container name, we're done
			if providedSetImageFlags.container != "" {
				break
			}
		}
	}

	if !updated {
		if providedSetImageFlags.container != "" {
			return fmt.Errorf("container %s not found in rollout", providedSetImageFlags.container)
		} else {
			return fmt.Errorf("no containers found in rollout")
		}
	}

	// Set the updated containers back
	if err := unstructured.SetNestedSlice(rollout.Object, containers, "spec", "template", "spec", "containers"); err != nil {
		return fmt.Errorf("failed to update containers: %s", err)
	}

	// Update the rollout
	_, err = dynamicClient.Resource(rolloutGVR).Namespace(resolveNamespace()).Update(context.Background(), rollout, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update rollout %s: %s", rolloutName, err)
	}

	cli.Printf("Rollout %s image updated successfully\n", rolloutName)
	return nil
}

// parseImageFlag parses container=image format or just image
func parseImageFlag(imageFlag string) (string, string, error) {
	parts := strings.SplitN(imageFlag, "=", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	} else if len(parts) == 1 {
		return "", parts[0], nil
	}
	return "", "", fmt.Errorf("invalid image format")
}

var setImageCmd = &cobra.Command{
	Use:   "set-image",
	Short: "Update rollout image",
	Example: `yak rollouts set-image -r my-rollout --image nginx:1.20
yak rollouts set-image -r my-rollout --container web --image nginx:1.20`,
	RunE: setImage,
}

func init() {
	setImageCmd.Flags().StringVarP(&providedSetImageFlags.rollout, "rollout", "r", "", "Rollout name (required)")
	setImageCmd.Flags().StringVarP(&providedSetImageFlags.container, "container", "c", "", "Container name (optional, defaults to first container)")
	setImageCmd.Flags().StringVar(&providedSetImageFlags.image, "image", "", "New image (required)")
	_ = setImageCmd.MarkFlagRequired("rollout")
	_ = setImageCmd.MarkFlagRequired("image")
}
