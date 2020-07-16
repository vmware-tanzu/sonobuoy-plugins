/*
Copyright the Sonobuoy contributors 2020

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
     http://www.apache.org/licenses/LICENSE-2.0

	 Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resources

import (
	"fmt"

	"github.com/vmware-tanzu/sonobuoy-plugins/cluster-inventory/pkg/reports"

	batchv1beta "k8s.io/api/batch/v1beta1"
)

type CronJob struct {
	batchv1beta.CronJob
	Jobs map[string]*Job
}

func (c CronJob) statusMessage() string {
	return fmt.Sprintf("Active: %d, Last Schedule: %v", len(c.Status.Active), c.Status.LastScheduleTime)
}

func (c CronJob) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name:   c.Name,
		Status: c.statusMessage(),
		Metadata: map[string]string{
			"kind": "CronJob",
			"uid":  string(c.UID),
		},
		Details: map[string]interface{}{
			"active":                    len(c.Status.Active),
			"schedule":                  c.Spec.Schedule,
			"suspend":                   c.Spec.Suspend,
			"concurrencyPolicy":         c.Spec.ConcurrencyPolicy,
			"lastScheduleTime":          c.Status.LastScheduleTime,
			"successfulJobHistoryLimit": c.Spec.SuccessfulJobsHistoryLimit,
			"failedJobHistoryLimit":     c.Spec.FailedJobsHistoryLimit,
		},
	}

	if len(c.Labels) > 0 {
		item.Details["labels"] = c.Labels
	}

	for _, job := range c.Jobs {
		item.Items = append(item.Items, job.GenerateSonobuoyItem())
	}

	return item
}
