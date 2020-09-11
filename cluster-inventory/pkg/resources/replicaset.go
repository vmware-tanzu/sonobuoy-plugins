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

type ReplicaSet struct {
	appsv1.ReplicaSet
	Pods map[string]*Pod
}

func (r ReplicaSet) statusMessage() string {
	return fmt.Sprintf("Desired: %d, Current: %d, Ready: %d, Available: %d",
		r.Spec.Replicas, r.Status.Replicas, r.Status.ReadyReplicas, r.Status.AvailableReplicas)
}

func (r ReplicaSet) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name:   r.Name,
		Status: r.statusMessage(),
		Metadata: map[string]string{
			"kind": "ReplicaSet",
			"uid":  string(r.UID),
		},
		Details: map[string]interface{}{
			"status":          r.Status,
			"replicas":        r.Spec.Replicas,
			"minReadySeconds": r.Spec.MinReadySeconds,
		},
	}

	if r.Spec.Selector != nil {
		item.Details["selector"] = r.Spec.Selector
	}

	if r.Spec.Template.Spec.NodeSelector != nil {
		item.Details["nodeSelector"] = r.Spec.Template.Spec.NodeSelector
	}

	if len(r.Labels) > 0 {
		item.Details["labels"] = r.Labels
	}

	for _, pod := range r.Pods {
		item.Items = append(item.Items, pod.GenerateSonobuoyItem())
	}

	return item
}
