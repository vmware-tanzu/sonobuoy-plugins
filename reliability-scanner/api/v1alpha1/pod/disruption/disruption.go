package disruption

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/sonobuoy-plugins/reliability-scanner/internal"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var checkName string = "disruption"

// QuerierSpec defines the Specification for a Querier.
type QuerierSpec struct {
}

// Querier defines the query and set of checks.
type Querier struct {
	client *kubernetes.Clientset
	Spec   *QuerierSpec `yaml:"spec"`
}

// AddtoRunner configures a runner with the Querier for this check.
func (querier *Querier) AddtoRunner(runner *internal.Runner) {
	runner.Queriers = append(runner.Queriers, querier)
	runner.Logger.WithFields(log.Fields{
		"check_name": checkName,
		"phase":      "add",
	}).Info("complete")
}

// NewQuerier returns a new configured Querier.
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

// Start runs the Querier.
func (q Querier) Start(cfg *internal.QuerierConfig) {
	cfg.Logger.WithFields(log.Fields{
		"check_name": checkName,
		"phase":      "add",
	}).Info(internal.CheckStartMsg)

	checkItem := internal.ReportItem{
		Name:   checkName,
		Status: "passed",
	}

	namespaces, err := q.client.CoreV1().Namespaces().List(cfg.Context, metav1.ListOptions{})
	if err != nil {
		checkItem.Status = "failed"
	}

	covered := make(map[string]map[string]bool)
	pods, err := q.client.CoreV1().Pods("").List(cfg.Context, metav1.ListOptions{})
	if err != nil {
		checkItem.Status = "failed"
	}
	for _, pod := range pods.Items {
		pods := make(map[string]bool)
		pods[pod.ObjectMeta.Name] = false
		covered[pod.ObjectMeta.Namespace] = pods
	}

	for _, namespace := range namespaces.Items {
		pdbs, err := q.client.PolicyV1beta1().PodDisruptionBudgets(namespace.Name).List(cfg.Context, metav1.ListOptions{})
		if err != nil {
			checkItem.Status = "failed"
		}
		for _, pdb := range pdbs.Items {
			pods, err := q.client.CoreV1().Pods(namespace.Name).List(cfg.Context, metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(pdb.Spec.Selector)})
			if err != nil {
				checkItem.Status = "failed"
			}
			for _, pod := range pods.Items {
				if !internal.IsSonobouyPod(pod.ObjectMeta.Name) {
					details := make(map[string]interface{})
					details["managing_disruption_budget"] = pdb.ObjectMeta.Name
					item := internal.Item{
						Name:    fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
						Status:  "passed",
						Details: details,
					}
					checkItem.Items = append(checkItem.Items, item)
					covered[namespace.Name][pod.ObjectMeta.Name] = true
				}
			}
		}
		for ns, pods := range covered {
			for pod := range pods {
				if pods[pod] == false {
					item := internal.Item{
						Name:   fmt.Sprintf("%s/%s", ns, pod),
						Status: "failed",
					}
					checkItem.Items = append(checkItem.Items, item)
				}
			}
		}

	}

	cfg.Logger.WithFields(log.Fields{
		"component":  "check",
		"check_name": checkName,
		"phase":      "debug",
	}).Info(covered)

	cfg.Logger.WithFields(log.Fields{
		"component":  "check",
		"check_name": checkName,
		"phase":      "complete",
	}).Info(internal.CheckCompleteMsg)

	cfg.Results <- checkItem

	cfg.Logger.WithFields(log.Fields{
		"component":  "check",
		"check_name": checkName,
		"phase":      "write",
	}).Info(internal.CheckWriteMsg)
}
