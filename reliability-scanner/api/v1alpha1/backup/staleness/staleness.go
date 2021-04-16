package staleness

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/sonobuoy-plugins/reliability-scanner/internal"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var checkName string = "staleness"

// QuerierSpec defines the Specification for a Querier.
type QuerierSpec struct {
	BackupNamespace string        `yaml:"backup_namespace"`
	MaxAge          time.Duration `yaml:"max_age"`
}

// Querier defines the query and set of checks.
type Querier struct {
	client dynamic.Interface
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
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return out, err
	}
	out.client = client
	return out, nil
}

// Start runs the Querier.
func (q Querier) Start(cfg *internal.QuerierConfig) {
	cfg.Logger.WithFields(log.Fields{
		"check_name": checkName,
		"phase":      "run",
	}).Info(internal.CheckStartMsg)
	var items []internal.Item
	checkItem := internal.ReportItem{
		Name:   checkName,
		Status: "passed",
		Items:  items,
	}
	backupGVR := schema.GroupVersionResource{Group: "velero.io", Version: "v1", Resource: "backups"}

	backups, err := q.client.Resource(backupGVR).Namespace(q.Spec.BackupNamespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		cfg.Logger.WithFields(log.Fields{
			"check_name": checkName,
			"phase":      "run",
		}).Info(err)
		details := make(map[string]interface{})
		details["error"] = "no backups defined"
		item := internal.Item{
			Name:    q.Spec.BackupNamespace,
			Status:  "failed",
			Details: details,
		}
		checkItem.Items = append(checkItem.Items, item)
		cfg.Results <- checkItem
		return
	}
	cfg.Logger.WithFields(log.Fields{
		"check_name": checkName,
		"phase":      "run",
		"backups":    backups,
	}).Info(backups)

	if backups != nil {
		if len(backups.Items) > 0 {
			for _, backup := range backups.Items {
				item := internal.Item{
					Name:   backup.GetName(),
					Status: "failed",
				}

				details := make(map[string]interface{})

				phase, found, err := unstructured.NestedString(backup.Object, "status", "phase")
				if err != nil || !found {
					cfg.Logger.WithFields(log.Fields{
						"component":   "check",
						"check_name":  checkName,
						"backup_name": backup.GetName(),
						"error":       err,
					}).Error("unable to retrieve backup status")
					details[backup.GetName()] = err
					item.Status = "failed"
				}

				if phase == "Completed" {
					item.Status = "passed"
				}

				expiration, found, err := unstructured.NestedString(backup.Object, "status", "expiration")
				if err != nil || !found {
					cfg.Logger.WithFields(log.Fields{
						"component":   "check",
						"check_name":  checkName,
						"backup_name": backup.GetName(),
						"error":       err,
					}).Error("unable to retrieve backup expiration")
					details[backup.GetName()] = err
					item.Status = "failed"
					continue
				}

				expiry, _ := time.Parse(time.RFC3339, expiration)
				t := time.Now()
				if t.After(expiry) {
					details["expired"] = expiry
					item.Status = "failed"
				}

				details["phase"] = phase
				item.Details = details
				checkItem.Items = append(checkItem.Items, item)
			}
		}
	}

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
