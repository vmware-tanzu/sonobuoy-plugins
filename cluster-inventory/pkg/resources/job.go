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

	batchv1 "k8s.io/api/batch/v1"
)

type Job struct {
	batchv1.Job
	Pods map[string]*Pod
}

func (j Job) statusMessage() string {
	return fmt.Sprintf("Running: %d, Succeeded: %d, Failed: %d",
		j.Status.Active, j.Status.Succeeded, j.Status.Failed)
}

func (j Job) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name:   j.Name,
		Status: j.statusMessage(),
		Details: map[string]interface{}{
			"status":   j.Status,
			"selector": j.Spec.Selector,
		},
	}

	if j.Spec.Parallelism != nil {
		item.Details["parallelism"] = j.Spec.Parallelism
	}

	if j.Spec.Completions != nil {
		item.Details["completions"] = j.Spec.Completions
	}

	if j.Spec.ActiveDeadlineSeconds != nil {
		item.Details["activeDeadlineSeconds"] = j.Spec.ActiveDeadlineSeconds
	}

	if j.Spec.BackoffLimit != nil {
		item.Details["backoffLimit"] = j.Spec.BackoffLimit
	}

	if j.Spec.TTLSecondsAfterFinished != nil {
		item.Details["ttlSecondsAfterFinished"] = j.Spec.TTLSecondsAfterFinished
	}

	for _, pod := range j.Pods {
		item.Items = append(item.Items, pod.GenerateSonobuoyItem())
	}
	return item
}
