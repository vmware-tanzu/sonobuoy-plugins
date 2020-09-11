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

	appsv1 "k8s.io/api/apps/v1"
)

type Deployment struct {
	appsv1.Deployment
	ReplicaSets map[string]*ReplicaSet
}

func (d Deployment) statusMessage() string {
	return fmt.Sprintf("Desired: %d, Up-to-date: %d, Total: %d, Available: %d",
		d.Spec.Replicas, d.Status.UpdatedReplicas, d.Status.Replicas, d.Status.AvailableReplicas)
}

func (d Deployment) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name: d.Name,
		Metadata: map[string]string{
			"kind": "Deployment",
			"uid":  string(d.UID),
		},
		Details: map[string]interface{}{
			"status":             d.Status,
			"deploymentStrategy": d.Spec.Strategy,
			"minReadySeconds":    d.Spec.MinReadySeconds,
			"paused":             d.Spec.Paused,
		},
	}

	if replicas := d.Spec.Replicas; replicas != nil {
		item.Details["replicas"] = replicas
	}

	if progressDeadline := d.Spec.ProgressDeadlineSeconds; progressDeadline != nil {
		item.Details["progressDeadlineSeconds"] = progressDeadline
	}

	if rhl := d.Spec.RevisionHistoryLimit; rhl != nil {
		item.Details["revisionHistoryLimit"] = rhl
	}

	if selector := d.Spec.Selector; selector != nil {
		item.Details["selector"] = selector
	}

	if nodeSelector := d.Spec.Template.Spec.NodeSelector; nodeSelector != nil {
		item.Details["nodeSelector"] = nodeSelector
	}

	if len(d.Labels) > 0 {
		item.Details["labels"] = d.Labels
	}

	for _, rs := range d.ReplicaSets {
		item.Items = append(item.Items, rs.GenerateSonobuoyItem())
	}

	return item
}
