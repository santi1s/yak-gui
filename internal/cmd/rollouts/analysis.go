package rollouts

import (
	"context"
	"fmt"
	"sort"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	providedAnalysisFlags RolloutsAnalysisFlags

	analysisCmd = &cobra.Command{
		Use:   "analysis",
		Short: "Manage analysis templates and runs",
	}
)

type RolloutsAnalysisFlags struct {
	template string
	run      string
	all      bool
}

func init() {
	analysisCmd.AddCommand(analysisListCmd)
	analysisCmd.AddCommand(analysisGetCmd)
	analysisCmd.AddCommand(analysisStatusCmd)
	analysisCmd.AddCommand(analysisLogsCmd)
}

// Analysis List Command
var analysisListCmd = &cobra.Command{
	Use:   "list",
	Short: "List analysis templates or runs",
	Example: `yak rollouts analysis list
yak rollouts analysis list --all`,
	RunE: analysisList,
}

func analysisList(cmd *cobra.Command, args []string) error {
	config, _, err := helper.InitKubeClusterConfig()
	if err != nil {
		return fmt.Errorf("unable to connect to Kubernetes cluster: %s", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to create dynamic client: %s", err)
	}

	templateGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "analysistemplates",
	}

	runGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "analysisruns",
	}

	var templates *unstructured.UnstructuredList
	var runs *unstructured.UnstructuredList

	if providedAnalysisFlags.all {
		templates, err = dynamicClient.Resource(templateGVR).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list analysis templates: %s", err)
		}

		runs, err = dynamicClient.Resource(runGVR).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list analysis runs: %s", err)
		}
	} else {
		templates, err = dynamicClient.Resource(templateGVR).Namespace(resolveNamespace()).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list analysis templates: %s", err)
		}

		runs, err = dynamicClient.Resource(runGVR).Namespace(resolveNamespace()).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list analysis runs: %s", err)
		}
	}

	if providedFlags.json || providedFlags.yaml {
		result := map[string]interface{}{
			"templates": templates,
			"runs":      runs,
		}
		return formatOutput(result)
	} else {
		formatAnalysisList(templates, runs)
	}
	return nil
}

func formatAnalysisList(templates *unstructured.UnstructuredList, runs *unstructured.UnstructuredList) {
	cli.Printf("ANALYSIS TEMPLATES:\n")
	if len(templates.Items) == 0 {
		cli.Printf("  No analysis templates found\n")
	} else {
		for _, template := range templates.Items {
			metrics, _, _ := unstructured.NestedSlice(template.Object, "spec", "metrics")
			cli.Printf("  %s/%s (%d metrics)\n", template.GetNamespace(), template.GetName(), len(metrics))
		}
	}

	cli.Printf("\nANALYSIS RUNS:\n")
	if len(runs.Items) == 0 {
		cli.Printf("  No analysis runs found\n")
	} else {
		for _, run := range runs.Items {
			phase, _, _ := unstructured.NestedString(run.Object, "status", "phase")
			if phase == "" {
				phase = "Unknown"
			}
			cli.Printf("  %s/%s (%s)\n", run.GetNamespace(), run.GetName(), phase)
		}
	}
}

// Analysis Get Command
var analysisGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get analysis template or run details",
	Example: `yak rollouts analysis get -t my-template
yak rollouts analysis get -r my-run`,
	RunE: analysisGet,
}

func analysisGet(cmd *cobra.Command, args []string) error {
	config, _, err := helper.InitKubeClusterConfig()
	if err != nil {
		return fmt.Errorf("unable to connect to Kubernetes cluster: %s", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to create dynamic client: %s", err)
	}

	if providedAnalysisFlags.template != "" {
		return getAnalysisTemplate(dynamicClient, providedAnalysisFlags.template)
	} else if providedAnalysisFlags.run != "" {
		return getAnalysisRun(dynamicClient, providedAnalysisFlags.run)
	} else {
		return fmt.Errorf("either --template or --run must be specified")
	}
}

func getAnalysisTemplate(dynamicClient dynamic.Interface, templateName string) error {
	templateGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "analysistemplates",
	}

	template, err := dynamicClient.Resource(templateGVR).Namespace(resolveNamespace()).Get(context.Background(), templateName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("analysis template %s not found in namespace %s: %s", templateName, resolveNamespace(), err)
	}

	if providedFlags.json || providedFlags.yaml {
		return formatOutput(template)
	} else {
		formatAnalysisTemplateDetails(template)
	}
	return nil
}

func getAnalysisRun(dynamicClient dynamic.Interface, runName string) error {
	runGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "analysisruns",
	}

	run, err := dynamicClient.Resource(runGVR).Namespace(resolveNamespace()).Get(context.Background(), runName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("analysis run %s not found in namespace %s: %s", runName, resolveNamespace(), err)
	}

	if providedFlags.json || providedFlags.yaml {
		return formatOutput(run)
	} else {
		formatAnalysisRunDetails(run)
	}
	return nil
}

func formatAnalysisTemplateDetails(template *unstructured.Unstructured) {
	name := template.GetName()
	namespace := template.GetNamespace()

	cli.Printf("Analysis Template: %s/%s\n", namespace, name)
	cli.Printf("---\n")

	metrics, found, _ := unstructured.NestedSlice(template.Object, "spec", "metrics")
	if found && len(metrics) > 0 {
		cli.Printf("Metrics:\n")
		for i, metric := range metrics {
			metricMap, ok := metric.(map[string]interface{})
			if !ok {
				continue
			}

			metricName, _ := metricMap["name"].(string)
			provider, _ := metricMap["provider"].(map[string]interface{})

			cli.Printf("  %d. %s\n", i+1, metricName)

			for providerType, providerConfig := range provider {
				cli.Printf("     Provider: %s\n", providerType)
				if providerConfig != nil {
					cli.Printf("     Config: %+v\n", providerConfig)
				}
			}
		}
	}
}

func formatAnalysisRunDetails(run *unstructured.Unstructured) {
	name := run.GetName()
	namespace := run.GetNamespace()
	phase, _, _ := unstructured.NestedString(run.Object, "status", "phase")
	message, _, _ := unstructured.NestedString(run.Object, "status", "message")

	cli.Printf("Analysis Run: %s/%s\n", namespace, name)
	cli.Printf("Phase: %s\n", phase)
	if message != "" {
		cli.Printf("Message: %s\n", message)
	}
	cli.Printf("---\n")

	metricResults, found, _ := unstructured.NestedSlice(run.Object, "status", "metricResults")
	if found && len(metricResults) > 0 {
		cli.Printf("Metric Results:\n")
		for _, result := range metricResults {
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				continue
			}

			metricName, _ := resultMap["name"].(string)
			phase, _ := resultMap["phase"].(string)
			value, _ := resultMap["value"].(string)

			cli.Printf("  %s: %s", metricName, phase)
			if value != "" {
				cli.Printf(" (value: %s)", value)
			}
			cli.Printf("\n")
		}
	}
}

// Analysis Status Command
var analysisStatusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show analysis run status",
	Example: `yak rollouts analysis status -r my-run`,
	RunE:    analysisStatus,
}

func analysisStatus(cmd *cobra.Command, args []string) error {
	if providedAnalysisFlags.run == "" {
		return fmt.Errorf("analysis run name is required")
	}

	config, _, err := helper.InitKubeClusterConfig()
	if err != nil {
		return fmt.Errorf("unable to connect to Kubernetes cluster: %s", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to create dynamic client: %s", err)
	}

	runGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "analysisruns",
	}

	run, err := dynamicClient.Resource(runGVR).Namespace(resolveNamespace()).Get(context.Background(), providedAnalysisFlags.run, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("analysis run %s not found in namespace %s: %s", providedAnalysisFlags.run, resolveNamespace(), err)
	}

	if providedFlags.json || providedFlags.yaml {
		return formatOutput(run)
	} else {
		formatAnalysisRunStatus(run)
	}
	return nil
}

func formatAnalysisRunStatus(run *unstructured.Unstructured) {
	name := run.GetName()
	phase, _, _ := unstructured.NestedString(run.Object, "status", "phase")
	message, _, _ := unstructured.NestedString(run.Object, "status", "message")

	cli.Printf("Analysis Run: %s\n", name)
	cli.Printf("Phase: %s\n", phase)
	if message != "" {
		cli.Printf("Message: %s\n", message)
	}

	metricResults, found, _ := unstructured.NestedSlice(run.Object, "status", "metricResults")
	if found && len(metricResults) > 0 {
		cli.Printf("\nMetric Results:\n")

		// Sort metrics by name for consistent output
		var sortedResults []map[string]interface{}
		for _, result := range metricResults {
			if resultMap, ok := result.(map[string]interface{}); ok {
				sortedResults = append(sortedResults, resultMap)
			}
		}

		sort.Slice(sortedResults, func(i, j int) bool {
			nameI, _ := sortedResults[i]["name"].(string)
			nameJ, _ := sortedResults[j]["name"].(string)
			return nameI < nameJ
		})

		for _, result := range sortedResults {
			metricName, _ := result["name"].(string)
			phase, _ := result["phase"].(string)
			value, _ := result["value"].(string)
			message, _ := result["message"].(string)

			cli.Printf("  %s: %s", metricName, phase)
			if value != "" {
				cli.Printf(" (value: %s)", value)
			}
			if message != "" {
				cli.Printf(" - %s", message)
			}
			cli.Printf("\n")
		}
	}
}

// Analysis Logs Command
var analysisLogsCmd = &cobra.Command{
	Use:     "logs",
	Short:   "Get analysis run logs",
	Example: `yak rollouts analysis logs -r my-run`,
	RunE:    analysisLogs,
}

func analysisLogs(cmd *cobra.Command, args []string) error {
	if providedAnalysisFlags.run == "" {
		return fmt.Errorf("analysis run name is required")
	}

	// For now, we'll show the analysis run status and metrics
	// In a full implementation, you might want to get logs from the underlying pods
	return analysisStatus(cmd, args)
}

func init() {
	analysisListCmd.Flags().BoolVar(&providedAnalysisFlags.all, "all", false, "List from all namespaces")

	analysisGetCmd.Flags().StringVarP(&providedAnalysisFlags.template, "template", "t", "", "Analysis template name")
	analysisGetCmd.Flags().StringVarP(&providedAnalysisFlags.run, "run", "r", "", "Analysis run name")
	analysisGetCmd.MarkFlagsMutuallyExclusive("template", "run")

	analysisStatusCmd.Flags().StringVarP(&providedAnalysisFlags.run, "run", "r", "", "Analysis run name (required)")
	_ = analysisStatusCmd.MarkFlagRequired("run")

	analysisLogsCmd.Flags().StringVarP(&providedAnalysisFlags.run, "run", "r", "", "Analysis run name (required)")
	_ = analysisLogsCmd.MarkFlagRequired("run")
}
