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
	"github.com/vmware-tanzu/sonobuoy-plugins/cluster-inventory/pkg/reports"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Node struct {
	v1.Node
}

func (n Node) statusMessage() string {
	var status []string
	for _, cond := range n.Status.Conditions {
		if cond.Status == v1.ConditionTrue {
			status = append(status, string(cond.Type))
		}
	}

	if len(status) == 0 {
		status = append(status, "Unknown")
	}
	if n.Spec.Unschedulable {
		status = append(status, "SchedulingDisabled")
	}

	return strings.Join(status, ",")
}

func (n Node) parseResources() map[string]map[string]string {
	processResourceList := func(resourceList v1.ResourceList) map[string]string {
		resourceLimits := map[string]string{}
		for r, q := range resourceList {
			resourceLimits[r.String()] = q.String()
		}
		return resourceLimits
	}

	return map[string]map[string]string{
		"allocatable": processResourceList(n.Status.Allocatable),
		"capacity":    processResourceList(n.Status.Capacity),
	}
}

func (n *Node) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name:   n.Name,
		Status: n.statusMessage(),
		Details: map[string]interface{}{
			"conditions":      n.Status.Conditions,
			"images":          n.Status.Images,
			"resources":       n.parseResources(),
			"addresses":       n.Status.Addresses,
			"volumesInUse":    n.Status.VolumesInUse,
			"volumesAttached": n.Status.VolumesAttached,
			"nodeInfo":        n.Status.NodeInfo,
			"podCIDR":         n.Spec.PodCIDR,
			"unschedulable":   fmt.Sprint(n.Spec.Unschedulable),
		},
	}

	if len(n.Spec.PodCIDRs) > 0 {
		item.Details["podCIDRs"] = n.Spec.PodCIDRs
	}

	if len(n.Spec.ProviderID) > 0 {
		item.Details["providerID"] = n.Spec.ProviderID
	}

	if len(n.Spec.Taints) > 0 {
		item.Details["taints"] = n.Spec.Taints
	}

	if len(n.Labels) > 0 {
		item.Details["labels"] = n.Labels
	}

	return item
}

type Nodes struct {
	Nodes []Node

	err error
}

func (n Nodes) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name:   "Nodes",
		Status: "complete",
	}

	for _, node := range n.Nodes {
		item.Items = append(item.Items, node.GenerateSonobuoyItem())
	}

	return item
}

func GetNodes(client *kubernetes.Clientset) Nodes {
	nodeList, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return Nodes{
			err: err,
		}
	}

	nodes := []Node{}
	for _, node := range nodeList.Items {
		nodes = append(nodes, Node{node})
	}

	return Nodes{
		Nodes: nodes,
	}
}
