/*
Collect logs from couchbase cluster, sync gateways and the operator and send them to the support

Usage:

	yak couchbase logs [flags]

Flags:

	    --cao                   Collect CAO logs as well (default true)
	-b, --caoBinary string      The path to cao binary in your local machine (default "/usr/local/bin/cao")
	-c, --customerName string   Doctolib customer name on Couchbase support portal (default "doctolib+sas")
	-h, --help                  help for logs
	-t, --ticketNumber int      Ticket number for couchbase support
	-u, --uploadHost string     Upload logs host (default "https://uploads.couchbase.com")

Global Flags:

	-n, --kubenamespace string   kubernetes namespace where couchbase pods are running (default "couchbase")
*/
package couchbase

import (
	"errors"
	"fmt"
	"sync"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/spf13/cobra"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	errCaoLogsCollection       = errors.New("cannot collect CAO logs, ensure that you're providing the path to your local CAO binary")
	errCannotSendLogsToSupport = errors.New("cannot send logs to Couchbase support")
	providedCouchbaseLogFlags  couchbaseLogFlags
)

type couchbaseLogFlags struct {
	caoBinary string
	cao       bool
	couchbaseFlags
	customerName   string
	sgw            bool
	ticketNumber   int
	uploadLogsHost string
}

func logsCollect(cmd *cobra.Command, args []string) error {
	var errCaoLogs, err error
	var cbFiles, sgwFiles []string
	var logsCaoFileName, logCaoFile string

	errCaoLogs = nil
	logCaoFile = "None"
	cli.Println("Log collection started ...")
	if !providedCouchbaseLogFlags.cao {
		cli.Println("CAO log collection")
		logsCaoFileName, err = logsCollectCao(providedCouchbaseLogFlags.caoBinary, providedCouchbaseFlags.kubeNamespace)
		if err != nil {
			return err
		} else {
			logCaoFile = strings.Clone(logsCaoFileName)
			errCaoLogs = sendCaoLogsToSupport(logsCaoFileName)
		}
	}
	config, clientset, err := helper.InitKubeClusterConfig()
	if err != nil {
		return err
	}
	cbPods, err := findCouchbaseClusterPods(clientset, providedCouchbaseFlags.kubeNamespace)
	if err != nil {
		return err
	} else {
		cbFiles, err = logsCollectFromPods(config, clientset, cbPods, "/opt/couchbase/bin/cbcollect_info", "couchbase-server")
		if err != nil {
			return errCannotSendLogsToSupport
		}
	}
	if !providedCouchbaseLogFlags.sgw {
		sgwPods, err := findSgwPods(clientset, providedCouchbaseFlags.kubeNamespace)
		if err != nil {
			return err
		} else {
			sgwFiles, err = logsCollectFromPods(config, clientset, sgwPods, "/opt/couchbase-sync-gateway/tools/sgcollect_info", "sync-gateway")
			if err != nil {
				return errCannotSendLogsToSupport
			}
		}
	}
	cli.Printf("CAO logs: %s \nCouchbase server logs: %v \nSync gateway logs: %v have been successfully sent to support", logCaoFile, cbFiles, sgwFiles)
	return errCaoLogs
}

func logsCollectCao(pathCaoBinary string, kubeNamespace string) (string, error) {
	cmd := exec.Command(pathCaoBinary, "collect-logs", "--all", "--log-level", "1", "-n", kubeNamespace) //#nosec G204
	out, err := cmd.Output()
	if err != nil {
		return "", errCaoLogsCollection
	}
	// otherwise, grab the logs file name from output
	outputStrings := strings.Split(string(out), " ")
	logsFileName := strings.TrimSpace(outputStrings[len(outputStrings)-1])
	return logsFileName, err
}

func sendCaoLogsToSupport(caoLogFile string) error {
	// example: formattedURL = https://uploads.couchbase.com/doctolib+sas/55807/
	formattedURL := fmt.Sprintf("%s/%s/%d/", providedCouchbaseLogFlags.uploadLogsHost, providedCouchbaseLogFlags.customerName, providedCouchbaseLogFlags.ticketNumber)
	cmd := exec.Command("curl", "--upload-file", caoLogFile, formattedURL)
	_, err := cmd.Output()
	if err != nil {
		return errCannotSendLogsToSupport
	}
	// Remove CAO logs file
	defer os.Remove(caoLogFile)
	return err
}

func logsCollectFromPods(config *rest.Config, clientset *kubernetes.Clientset, pods []v1.Pod, binary string, containerName string) ([]string, error) {
	cli.Println("Collecting servers logs")

	if len(pods) == 0 {
		return nil, fmt.Errorf("no pods provided")
	}

	currentTime := time.Now()
	timestamp := currentTime.Format("2006-01-02T15-04-05Z07-00") // RFC3339 without colons

	// Limit concurrent operations
	maxConcurrent := 3
	if len(pods) < maxConcurrent {
		maxConcurrent = len(pods)
	}

	type result struct {
		fileName string
		err      error
		podName  string
	}

	results := make(chan result, len(pods))
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for _, pod := range pods {
		wg.Add(1)
		go func(p v1.Pod) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			fileName := fmt.Sprintf("%s-%s.zip", pod.Name, timestamp)
			command := fmt.Sprintf("%s --ticket %d --customer=%s --upload-host=%s /tmp/%s",
				binary,
				providedCouchbaseLogFlags.ticketNumber,
				providedCouchbaseLogFlags.customerName,
				providedCouchbaseLogFlags.uploadLogsHost,
				fileName)
			err := helper.ExecPod(config, clientset, pod, containerName, strings.Split(command, " "))

			results <- result{
				fileName: fileName,
				err:      err,
				podName:  p.Name,
			}
		}(pod)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var logCollectedFiles []string
	var errors []error

	for res := range results {
		if res.err != nil {
			cli.Printf("Failed to collect logs from pod %s: %v\n", res.podName, res.err)
			errors = append(errors, fmt.Errorf("pod %s: %w", res.podName, res.err))
		} else {
			cli.Printf("Successfully collected logs from pod %s\n", res.podName)
			logCollectedFiles = append(logCollectedFiles, res.fileName)
		}
	}

	if len(errors) > 0 {
		return logCollectedFiles, fmt.Errorf("failed to collect logs from %d out of %d pods: %v", len(errors), len(pods), errors)
	}

	return logCollectedFiles, nil
}

var collectLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Collect logs from couchbase cluster, sync gateways and the operator and send them to the support",
	RunE:  logsCollect,
}

func init() {
	collectLogsCmd.Flags().StringVarP(&providedCouchbaseLogFlags.customerName, "customerName", "c", "doctolib+sas", "Doctolib customer name on Couchbase support portal")
	collectLogsCmd.Flags().IntVarP(&providedCouchbaseLogFlags.ticketNumber, "ticketNumber", "t", 0, "Ticket number for couchbase support")
	collectLogsCmd.Flags().StringVarP(&providedCouchbaseLogFlags.uploadLogsHost, "uploadHost", "u", "https://uploads.couchbase.com", "Upload logs host")
	collectLogsCmd.Flags().StringVarP(&providedCouchbaseLogFlags.caoBinary, "caoBinary", "b", "/usr/local/bin/cao", "The path to cao binary in your local machine; checkout https://docs.couchbase.com/operator/current/tools/cao.html to install it")
	collectLogsCmd.Flags().BoolVar(&providedCouchbaseLogFlags.cao, "no-cao", false, "Don't collect CAO logs")
	collectLogsCmd.Flags().BoolVar(&providedCouchbaseLogFlags.sgw, "no-sgw", false, "Don't collect sync-gateway logs")
}
