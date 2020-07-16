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

package cluster

import (
	"github.com/vmware-tanzu/sonobuoy-plugins/cluster-inventory/pkg/reports"
	"k8s.io/client-go/kubernetes"
)

type Components struct {
	Nodes         Nodes
	ControlPlane  ControlPlane
	CNI           CNIStatus
	NetworkStatus NetworkStatus

	// TODO retrieve CSI information
}

func (c Components) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name:   "Cluster Components",
		Status: "complete",
	}

	item.Items = append(item.Items,
		c.Nodes.GenerateSonobuoyItem(),
		c.CNI.GenerateSonobuoyItem(),
		c.ControlPlane.GenerateSonobuoyItem(),
		c.NetworkStatus.GenerateSonobuoyItem(),
	)

	return item
}

func GetComponents(client *kubernetes.Clientset) (*Components, error) {
	return &Components{
		Nodes:         GetNodes(client),
		ControlPlane:  GetControlPlane(client),
		CNI:           GetCNI(),
		NetworkStatus: GetNetworkStatus(),
	}, nil
}
