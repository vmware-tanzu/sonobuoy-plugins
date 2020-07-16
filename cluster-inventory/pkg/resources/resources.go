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
	"context"

	"github.com/vmware-tanzu/sonobuoy-plugins/cluster-inventory/pkg/reports"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type NamespacedWorkloads map[string]*WorkloadsTree

func GetWorkloads(client *kubernetes.Clientset) (NamespacedWorkloads, error) {
	namespaceList, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return NamespacedWorkloads{}, err
	}

	namespaceWorkloads := NamespacedWorkloads{}
	for _, ns := range namespaceList.Items {
		workloads, err := getWorkloadsForNamespace(client, ns.Name)
		if err != nil {
			return namespaceWorkloads, err
		}
		namespaceWorkloads[ns.Name] = workloads

	}
	return namespaceWorkloads, nil
}

func (nw NamespacedWorkloads) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name:   "Namespaced Workloads",
		Status: "complete",
	}

	for _, wt := range nw {
		item.Items = append(item.Items, wt.GenerateSonobuoyItem())
	}

	return item
}
