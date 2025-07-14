package helper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/santi1s/yak/cli"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

func pathToKubeConfig() string {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		// Fall back to default kubeconfig location
		if home := os.Getenv("HOME"); home != "" {
			return filepath.Join(home, ".kube", "config")
		}
	}
	return strings.Split(kubeconfig, ":")[0]
}

func GetKubernetesCurrentContext() (string, error) {
	//config, client, err := InitKubeClusterConfig()
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: pathToKubeConfig()},
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		}).RawConfig()
	if err != nil {
		return "", err
	}
	currentContext := config.CurrentContext
	return currentContext, nil
}

// GetKubernetesCurrentNamespace returns the namespace from the current kubeconfig context
func GetKubernetesCurrentNamespace() (string, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: pathToKubeConfig()},
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		}).RawConfig()
	if err != nil {
		return "", err
	}

	currentContext := config.CurrentContext
	if currentContext == "" {
		return "default", nil
	}

	context, exists := config.Contexts[currentContext]
	if !exists {
		return "default", nil
	}

	if context.Namespace == "" {
		return "default", nil
	}

	return context.Namespace, nil
}

func InitKubeClusterConfig() (*rest.Config, *kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// Try to load config from kubeconfig file first
	config, err = clientcmd.BuildConfigFromFlags("", pathToKubeConfig())
	if err != nil {
		// If that fails, try in-cluster config (when running inside a pod)
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, nil, fmt.Errorf("unable to load kubeconfig or in-cluster config: %v", err)
		}
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return config, clientset, err
}

func GetPodsMatchingLabel(clientset *kubernetes.Clientset, kubeNamespace string, labelKey string, labelValue string) ([]v1.Pod, error) {
	labelToMatch := labelKey + "=" + labelValue
	listOptionsFetch := metav1.ListOptions{
		LabelSelector: labelToMatch,
	}
	pods, err := clientset.CoreV1().Pods(kubeNamespace).List(context.TODO(), listOptionsFetch)
	if err != nil {
		_, _ = cli.PrintfErr("[ERR] Unable to list pods in namespace %s, matching label %s, %v\n", kubeNamespace, labelToMatch, err)
		return []v1.Pod{}, err
	}
	return pods.Items, err
}

func GetDeployment(clientset *kubernetes.Clientset, kubeNamespace string, deploymentName string) (*appsv1.Deployment, error) {
	deploymentsClient := clientset.AppsV1().Deployments(kubeNamespace)
	deployment, err := deploymentsClient.Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		log.Errorf("Unable to get deployment %s in namespace %s: %v\n", deploymentName, kubeNamespace, err)
		return nil, err
	}
	return deployment, err
}

func DeleteDeployment(clientset *kubernetes.Clientset, kubeNamespace string, deploymentName string, askConfirmation bool) error {
	deploymentsClient := clientset.AppsV1().Deployments(kubeNamespace)
	deletePolicy := metav1.DeletePropagationForeground
	if askConfirmation {
		confirmation := cli.AskConfirmation(fmt.Sprintf("\nDo you want to delete the deployment %s in namespace %s", deploymentName, kubeNamespace))
		if !confirmation {
			return fmt.Errorf("\nDeletion interrupted")
		}
	}
	if err := deploymentsClient.Delete(context.TODO(), deploymentName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		log.Errorf("Unable to delete deployment %s in namespace %s: %v\n", deploymentName, kubeNamespace, err)
		return err
	}
	return nil
}

func ShutdownDeploymentInKubernetesWithDryRun(clusterName string, deploymentName string, kubeNamespace string, dryRun bool) error {
	_, clientset, err := InitKubeClusterConfig()
	if err != nil {
		return err
	}
	_, err = GetDeployment(clientset, kubeNamespace, deploymentName)
	if err != nil {
		return err
	}
	if !dryRun {
		err = DeleteDeployment(clientset, kubeNamespace, deploymentName, true)
		if err != nil {
			return err
		}
	} else {
		cli.Printf("\n[%s][DRY RUN] the deployment %s/%s would be deleted\n", clusterName, kubeNamespace, deploymentName)
	}
	return nil
}

func ExecPod(config *rest.Config, clientset *kubernetes.Clientset, pod v1.Pod, containerName string, command []string) error {
	ctx := context.Background()
	execRequest := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		Param("container", containerName)
	execRequest.VersionedParams(&v1.PodExecOptions{
		Command: command,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", execRequest.URL())
	if err != nil {
		_, _ = cli.PrintfErr("[ERR] Unable to establish connection to pod %s/%s, %v \n", pod.Namespace, pod.Name, err)
		return err
	}

	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr, Tty: true,
	})
	if err != nil {
		_, _ = cli.PrintfErr("[ERR] Command %s failed execution on %s/%s, %v \n", strings.Join(command, " "), pod.Namespace, pod.Name, err)
		return err
	}
	return err
}

func FetchSecretValues(clientset *kubernetes.Clientset, kubeNamespace string, secretName string) (map[string]string, error) {
	secretData := make(map[string]string)
	// Read the secret from the specified namespace
	secret, err := clientset.CoreV1().Secrets(kubeNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		_, _ = cli.PrintfErr("[ERR] The secret %s in namespace %s is not found because: %v\n", secretName, kubeNamespace, err)
		return secretData, err
	}

	for key, value := range secret.Data {
		secretData[key] = string(value)
	}
	return secretData, err
}
