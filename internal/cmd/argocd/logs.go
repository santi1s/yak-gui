package argocd

import (
	"context"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	argocdhelper "github.com/santi1s/yak/internal/helper/argocd"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ArgoCDLogsFlags struct {
	application string
	podName     string
	container   string
	follow      bool
	previous    bool
	tail        int64
	since       string
}

var providedLogsFlags ArgoCDLogsFlags

func logs(cmd *cobra.Command, args []string) error {
	appName := providedLogsFlags.application

	apiclient, err := argocdhelper.ArgocdLogin(&argocdhelper.LoginParams{
		ArgocdServer: providedFlags.addr,
	})
	if err != nil {
		return fmt.Errorf("unable to establish connection to argocd: %s", err)
	}

	myApp, err := argocdhelper.GetApplication(apiclient.AppClient, appName, providedFlags.project)
	if err != nil {
		return err
	}

	// Get Kubernetes client
	config, _, err := helper.InitKubeClusterConfig()
	if err != nil {
		return fmt.Errorf("unable to initialize kubernetes config: %s", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to create kubernetes client: %s", err)
	}

	// Get pods managed by the application
	pods, err := getApplicationPods(apiclient.AppClient, myApp.Name, myApp.Namespace, clientset)
	if err != nil {
		return fmt.Errorf("failed to get pods for application %s: %s", appName, err)
	}

	if len(pods) == 0 {
		cli.Printf("No pods found for application %s\n", appName)
		return nil
	}

	// Handle pod selection - if --pod flag was provided without value, show selection menu
	if providedLogsFlags.podName == "SELECT" {
		// Interactive pod selection
		selectedPod, err := podSelectMenu(pods)
		if err != nil {
			return fmt.Errorf("failed to select pod: %s", err)
		}
		if selectedPod == "" {
			cli.Println("No pod selected")
			return nil
		}
		providedLogsFlags.podName = selectedPod
	}

	// If specific pod is requested, filter to that pod
	if providedLogsFlags.podName != "" {
		filteredPods := []corev1.Pod{}
		for _, pod := range pods {
			if pod.Name == providedLogsFlags.podName {
				filteredPods = append(filteredPods, pod)
				break
			}
		}
		if len(filteredPods) == 0 {
			return fmt.Errorf("pod %s not found in application %s", providedLogsFlags.podName, appName)
		}
		pods = filteredPods
	}

	// Get logs for each pod
	for _, pod := range pods {
		err := getPodLogs(clientset, pod)
		if err != nil {
			cli.Printf("Failed to get logs for pod %s: %s\n", pod.Name, err)
		}
	}

	return nil
}

func podSelectMenu(pods []corev1.Pod) (string, error) {
	if len(pods) == 0 {
		return "", fmt.Errorf("no pods available")
	}

	var podOptions []string
	for _, pod := range pods {
		status := string(pod.Status.Phase)
		if pod.Status.Phase == corev1.PodRunning {
			// Show more detailed status for running pods
			readyContainers := 0
			totalContainers := len(pod.Status.ContainerStatuses)
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.Ready {
					readyContainers++
				}
			}
			status = fmt.Sprintf("Running (%d/%d)", readyContainers, totalContainers)
		}

		namespace := pod.Namespace
		if namespace == "" {
			namespace = "default"
		}

		option := fmt.Sprintf("%s [%s] (%s)", pod.Name, status, namespace)
		podOptions = append(podOptions, option)
	}

	// Add "None" option to allow user to cancel
	podOptions = append(podOptions, "None - Cancel")

	var selectedOption string
	prompt := &survey.Select{
		Message: "Select a pod:",
		Options: podOptions,
	}

	err := survey.AskOne(prompt, &selectedOption)
	if err != nil {
		return "", err
	}

	// Handle "None" selection
	if selectedOption == "None - Cancel" {
		return "", nil
	}

	// Extract pod name from the selected option (everything before the first space)
	parts := strings.Split(selectedOption, " ")
	if len(parts) > 0 {
		return parts[0], nil
	}

	return "", fmt.Errorf("invalid selection")
}

func getApplicationPods(appClient application.ApplicationServiceClient, appName, appNamespace string, clientset *kubernetes.Clientset) ([]corev1.Pod, error) {
	var pods []corev1.Pod

	// Get the resource tree for the application
	resourceTree, err := appClient.ResourceTree(context.Background(), &application.ResourcesQuery{
		ApplicationName: &appName,
		AppNamespace:    &appNamespace,
	})
	if err != nil {
		return pods, err
	}

	// Find all pod resources in the tree
	var podNamespaces = make(map[string][]string) // namespace -> []podNames

	for _, node := range resourceTree.Nodes {
		if node.Kind == "Pod" {
			if podNamespaces[node.Namespace] == nil {
				podNamespaces[node.Namespace] = []string{}
			}
			podNamespaces[node.Namespace] = append(podNamespaces[node.Namespace], node.Name)
		}
	}

	// Get actual pod objects from Kubernetes
	for namespace, podNames := range podNamespaces {
		for _, podName := range podNames {
			pod, err := clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
			if err != nil {
				cli.Printf("Warning: Could not get pod %s/%s: %s\n", namespace, podName, err)
				continue
			}
			pods = append(pods, *pod)
		}
	}

	return pods, nil
}

func getPodLogs(clientset *kubernetes.Clientset, pod corev1.Pod) error {
	// Determine which container to get logs from
	containerName := providedLogsFlags.container
	if containerName == "" && len(pod.Spec.Containers) > 0 {
		containerName = pod.Spec.Containers[0].Name
	}

	// Build log options
	logOptions := &corev1.PodLogOptions{
		Container: containerName,
		Follow:    providedLogsFlags.follow,
		Previous:  providedLogsFlags.previous,
	}

	if providedLogsFlags.tail > 0 {
		logOptions.TailLines = &providedLogsFlags.tail
	}

	if providedLogsFlags.since != "" {
		// Parse since time (simplified - only supports duration strings like "5m", "1h")
		logOptions.SinceSeconds = parseSinceSeconds(providedLogsFlags.since)
	}

	// Get logs
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)
	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return err
	}
	defer podLogs.Close()

	// Print header
	cli.Printf("==> %s/%s (%s) <==\n", pod.Namespace, pod.Name, containerName)

	// Stream logs
	buf := make([]byte, 2048)
	for {
		numBytes, err := podLogs.Read(buf)
		if numBytes == 0 {
			break
		}
		if err != nil {
			break
		}
		message := string(buf[:numBytes])
		cli.Print(message)
	}

	if len(pod.Spec.Containers) > 1 {
		cli.Printf("\n")
	}

	return nil
}

func parseSinceSeconds(since string) *int64 {
	// Simple duration parsing - supports formats like "5m", "1h", "30s"
	since = strings.TrimSpace(since)
	if since == "" {
		return nil
	}

	var multiplier int64 = 1
	if strings.HasSuffix(since, "s") {
		multiplier = 1
		since = strings.TrimSuffix(since, "s")
	} else if strings.HasSuffix(since, "m") {
		multiplier = 60
		since = strings.TrimSuffix(since, "m")
	} else if strings.HasSuffix(since, "h") {
		multiplier = 3600
		since = strings.TrimSuffix(since, "h")
	}

	// Try to parse the number
	var seconds int64
	_, _ = fmt.Sscanf(since, "%d", &seconds)
	result := seconds * multiplier
	return &result
}

var logsCmd = &cobra.Command{
	Use:     "logs",
	Short:   "Get logs from pods managed by an ArgoCD application",
	Example: "yak argocd logs --application my-app\nyak argocd logs --application my-app --pod\nyak argocd logs --application my-app --pod my-pod-123 --container my-container",
	RunE:    logs,
}

func init() {
	logsCmd.Flags().StringVarP(&providedLogsFlags.application, "application", "a", "", "ArgoCD application name")
	logsCmd.Flags().StringVarP(&providedLogsFlags.podName, "pod", "p", "", "Specific pod name to get logs from (omit value to select interactively)")
	logsCmd.Flags().Lookup("pod").NoOptDefVal = "SELECT"
	logsCmd.Flags().StringVarP(&providedLogsFlags.container, "container", "c", "", "Container name (defaults to first container)")
	logsCmd.Flags().BoolVarP(&providedLogsFlags.follow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().BoolVar(&providedLogsFlags.previous, "previous", false, "Get logs from previous container instance")
	logsCmd.Flags().Int64Var(&providedLogsFlags.tail, "tail", 0, "Number of lines to show from the end of the logs")
	logsCmd.Flags().StringVar(&providedLogsFlags.since, "since", "", "Show logs since duration (e.g., 5m, 1h)")
	_ = logsCmd.MarkFlagRequired("application")
}
