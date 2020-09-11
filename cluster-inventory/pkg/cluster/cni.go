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
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/vmware-tanzu/sonobuoy-plugins/cluster-inventory/pkg/reports"

	"github.com/containernetworking/cni/libcni"
)

type CNIStatus struct {
	*libcni.NetworkConfigList
	confFile string

	err error
}

const cniConfDir = "/etc/cni/net.d/"

// GetCNI looks in the CNI conf directory and uses the first conflist/conf file to determine
// the CNI information.
func GetCNI() CNIStatus {
	cniStatus := CNIStatus{}

	confFiles, err := libcni.ConfFiles(cniConfDir, []string{".conf", ".conflist", ".json"})
	if err != nil {
		cniStatus.err = err
		return cniStatus
	} else if len(confFiles) == 0 {
		cniStatus.err = fmt.Errorf("no CNI configuration files found in %s", cniConfDir)
		return cniStatus
	}

	sort.Strings(confFiles)

	// Iterate over all found confFiles and select the first one that passes all
	// validation steps

	var confList *libcni.NetworkConfigList
	for _, confFile := range confFiles {
		if filepath.Ext(confFile) == ".conflist" {
			confList, err = libcni.ConfListFromFile(confFile)
			if err != nil {
				fmt.Printf("error loading CNI conflist file %q: %v", confFile, err)
				continue
			}
		} else {
			conf, err := libcni.ConfFromFile(confFile)
			if err != nil {
				fmt.Printf("error loading CNI conf file %q: %v", confFile, err)
				continue
			}

			confList, err = libcni.ConfListFromConf(conf)
			if err != nil {
				fmt.Printf("error converting CNI conf to conflist: %v", err)
				continue
			}
		}

		if len(confList.Plugins) == 0 {
			fmt.Printf("CNI conflist %q has no networks, skipping", confFile)
			continue
		}

		cniConfig := libcni.CNIConfig{Path: []string{"/opt/cni/bin"}}
		_, err := cniConfig.ValidateNetworkList(context.TODO(), confList)
		if err != nil {
			fmt.Printf("error validating CNI conflist %q: %v", confFile, err)
			continue
		}

		// Once we have found a valid CNI conflist, store the details and exit the loop
		fmt.Println("Using CNI config file", confFile)
		cniStatus.confFile = confFile
		cniStatus.NetworkConfigList = confList
		break
	}

	return cniStatus
}

func (c CNIStatus) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name:   "CNI",
		Status: "complete",
	}

	if c.err != nil {
		item.Status = "incomplete"
		item.Details = map[string]interface{}{
			"error": c.err,
		}
		return item
	}

	cniItem := reports.SonobuoyResultsItem{
		Name:   c.NetworkConfigList.Name,
		Status: "complete",
		Metadata: map[string]string{
			"confFile": c.confFile,
		},
		Details: map[string]interface{}{
			"cniVersion":   c.CNIVersion,
			"disableCheck": c.DisableCheck,
		},
	}

	for _, plugin := range c.Plugins {
		cniItem.Items = append(cniItem.Items, reports.SonobuoyResultsItem{
			Name: plugin.Network.Name,
			Details: map[string]interface{}{
				"version":      plugin.Network.CNIVersion,
				"type":         plugin.Network.Type,
				"capabilities": plugin.Network.Capabilities,
				"ipam":         plugin.Network.IPAM,
				"dns":          plugin.Network.DNS,
			},
		})
	}

	item.Items = append(item.Items, cniItem)

	return item
}
