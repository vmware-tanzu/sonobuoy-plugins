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
	"net"

	"github.com/vmware-tanzu/sonobuoy-plugins/cluster-inventory/pkg/reports"
)

type NetworkStatus struct {
	ExternalDNS bool
}

func GetNetworkStatus() NetworkStatus {
	externalDNS := true
	if _, err := net.LookupIP("google.com"); err != nil {
		externalDNS = false
	}

	return NetworkStatus{
		ExternalDNS: externalDNS,
	}
}

func (n NetworkStatus) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	return reports.SonobuoyResultsItem{
		Name:   "Network Status",
		Status: "complete",
		Details: map[string]interface{}{
			"externalDNS": n.ExternalDNS,
		},
	}
}
