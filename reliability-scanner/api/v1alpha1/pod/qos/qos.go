package qos

import (
	"fmt"

	"github.com/vmware-tanzu/sonobuoy-plugins/reliability-scanner/internal"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	checkName string = "qos"
)

// QuerierSpec defines the Specification for a Querier
type QuerierSpec struct {
	MinimumDesiredQOSClass string `yaml:"minimum_desired_qos_class"`
	IncludeDetail          bool   `yaml:"include_detail"`
	// ExcludeLabelled []string `yaml:"exclude_labelled"`
}

// Querier defines the query and set of checks
type Querier struct {
	client *kubernetes.Clientset
	Spec   *QuerierSpec `yaml:"spec"`
}

// NewQuerier returns a new configured Querier
func NewQuerier(spec *QuerierSpec) (Querier, error) {
	out := Querier{
		Spec: spec,
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return out, err
	}
	out.client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return out, err
	}
	return out, nil
}

// AddtoRunner configures a runner with the Querier for this check.
func (querier *Querier) AddtoRunner(runner *internal.Runner) {
	runner.Queriers = append(runner.Queriers, querier)
	runner.Logger.WithFields(log.Fields{
		"check_name": checkName,
		"phase":      "add",
	}).Info("complete")
}

// Start runs the Querier
func (q Querier) Start(cfg *internal.QuerierConfig) {
	cfg.Logger.WithFields(log.Fields{
		"check": checkName,
		"phase": "starting",
	}).Info(internal.CheckStartMsg)

	checkItem := internal.ReportItem{
		Name:   checkName,
		Status: "passed",
	}

	pods, err := q.client.CoreV1().Pods("").List(cfg.Context, metav1.ListOptions{})
	if err != nil {
		checkItem.Status = "failed"
	}

	for _, pod := range pods.Items {
		details := make(map[string]interface{})

		item := internal.Item{
			Name:    fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
			Status:  "failed",
			Details: details,
		}
		if meetsMinimum(string(pod.Status.QOSClass), q.Spec.MinimumDesiredQOSClass) {
			item.Status = "passed"
		}
		if q.Spec.IncludeDetail {
			details["qos_class"] = pod.Status.QOSClass
		}

		checkItem.Items = append(checkItem.Items, item)
	}

	cfg.Logger.WithFields(log.Fields{
		"check_name": checkName,
		"phase":      "complete",
	}).Info(internal.CheckCompleteMsg)

	cfg.Results <- checkItem

	cfg.Logger.WithFields(log.Fields{
		"check_name": checkName,
		"phase":      "write",
	}).Info(internal.CheckWriteMsg)
}
