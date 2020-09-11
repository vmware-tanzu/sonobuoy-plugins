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

package inventory

import (
	"github.com/vmware-tanzu/sonobuoy-plugins/cluster-inventory/pkg/cluster"
	"github.com/vmware-tanzu/sonobuoy-plugins/cluster-inventory/pkg/resources"
	"k8s.io/client-go/kubernetes"
)

type Collector struct {
	client *kubernetes.Clientset
}

func (c *Collector) Run() (Results, error) {
	// Get Cluster components
	clusterComponents, err := cluster.GetComponents(c.client)
	if err != nil {
		return Results{}, err
	}

	// New details to collect

	// Namespaces
	namespaces := resources.GetNamespaces(c.client)

	// WorkloadsTree
	workloads, _ := resources.GetWorkloads(c.client)

	// Networking
	// - Services
	// - Ingress
	// - Network Policies?
	// CSI

	return Results{
		ClusterComponents: clusterComponents,
		Namespaces:        namespaces,
		Workloads:         workloads,
	}, nil
}

func NewCollector(client *kubernetes.Clientset) *Collector {
	return &Collector{client: client}
}
