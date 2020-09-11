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
	"encoding/json"
	"fmt"

	"github.com/vmware-tanzu/sonobuoy-plugins/cluster-inventory/pkg/reports"

	v1 "k8s.io/api/core/v1"
)

type Pod struct {
	v1.Pod
}

func generateContainerSonobuoyItem(container v1.Container, statuses []v1.ContainerStatus, isInit bool) reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name: container.Name,
		Metadata: map[string]string{
			"kind": "Container",
		},
		Details: map[string]interface{}{
			"image": container.Image,
		},
	}

	if isInit {
		item.Metadata["init"] = fmt.Sprint(isInit)
	}

	for _, status := range statuses {
		if status.Name == container.Name {
			switch {
			case status.State.Running != nil:
				item.Status = "Running"
				item.Details["state"] = map[string]interface{}{
					"running": status.State.Running,
				}
			case status.State.Waiting != nil:
				item.Status = "Waiting"
				item.Details["state"] = map[string]interface{}{
					"waiting": status.State.Waiting,
				}
			case status.State.Terminated != nil:
				item.Status = "Terminated"
				item.Details["state"] = map[string]interface{}{
					"terminated": status.State.Terminated,
				}
			}

			item.Details["imageID"] = status.ImageID
			item.Details["ready"] = status.Ready
			item.Details["restartCount"] = status.RestartCount
		}
	}

	item.Details["command"] = container.Command
	item.Details["args"] = container.Args
	item.Details["volumeMounts"] = container.VolumeMounts

	return item
}

func (p Pod) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name:   p.Name,
		Status: string(p.Status.Phase),
		Metadata: map[string]string{
			"kind": "Pod",
			"uid":  string(p.UID),
		},
		Details: map[string]interface{}{
			// TODO add container ports?
			// pvc, privileged containers, host namespaces etc
			"conditions":     p.Status.Conditions,
			"hostIP":         p.Status.HostIP,
			"node":           p.Spec.NodeName,
			"podIP":          p.Status.PodIP,
			"priority":       p.Spec.Priority,
			"qos":            p.Status.QOSClass,
			"serviceAccount": p.Spec.ServiceAccountName,
		},
	}

	volumes := []interface{}{}
	for _, v := range p.Spec.Volumes {
		var volume interface{}

		// Marshal and unmarshal the volume struct to force empty fields
		// to be discarded through the "omitempty" annotation.
		data, _ := json.Marshal(v)
		_ = json.Unmarshal(data, &volume)

		volumes = append(volumes, volume)
	}

	item.Details["volumes"] = volumes

	if len(p.Labels) > 0 {
		item.Details["labels"] = p.Labels
	}

	if len(p.Spec.Tolerations) > 0 {
		item.Details["tolerations"] = p.Spec.Tolerations
	}

	if len(p.Spec.NodeSelector) > 0 {
		item.Details["nodeSelector"] = p.Spec.NodeSelector
	}

	for _, container := range p.Spec.InitContainers {
		item.Items = append(item.Items, generateContainerSonobuoyItem(container, p.Status.InitContainerStatuses, true))
	}

	for _, container := range p.Spec.Containers {
		item.Items = append(item.Items, generateContainerSonobuoyItem(container, p.Status.ContainerStatuses, false))
	}

	return item
}
