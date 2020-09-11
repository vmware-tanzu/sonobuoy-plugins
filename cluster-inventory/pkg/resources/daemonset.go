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

type DaemonSet struct {
	appsv1.DaemonSet
	Pods map[string]*Pod
}

func (d DaemonSet) statusMessage() string {
	s := d.Status
	return fmt.Sprintf("Current: %d, Desired: %d, Ready: %d, Up-to-date: %d, Available: %d",
		s.CurrentNumberScheduled, s.DesiredNumberScheduled, s.NumberReady, s.UpdatedNumberScheduled, s.NumberAvailable)
}

func (d DaemonSet) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name:   d.Name,
		Status: d.statusMessage(),
		Metadata: map[string]string{
			"kind": "DaemonSet",
			"uid":  string(d.UID),
		},
		Details: map[string]interface{}{
			"status":         d.Status,
			"updateStrategy": d.Spec.UpdateStrategy,
		},
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

	for _, pod := range d.Pods {
		item.Items = append(item.Items, pod.GenerateSonobuoyItem())
	}

	return item
}
