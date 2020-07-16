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
	"strings"

	"github.com/vmware-tanzu/sonobuoy-plugins/cluster-inventory/pkg/reports"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ControlPlane struct {
	Provider        string
	IsHA            bool
	NumNodes        int
	AuditLogEnabled bool

	err error
}

func (c ControlPlane) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name:    "Control Plane",
		Details: map[string]interface{}{},
	}
	if c.err != nil {
		item.Status = "incomplete"
		item.Details["error"] = c.err
		return item
	}

	item.Status = "complete"
	item.Details["auditLogEnabled"] = c.AuditLogEnabled
	item.Details["isHA"] = c.IsHA
	item.Details["numNodes"] = c.NumNodes

	if c.Provider != "" {
		item.Details["provider"] = c.Provider
	}

	return item
}

func getProvider(node v1.Node) string {
	provider := ""
	switch {
	case strings.Contains(node.Spec.ProviderID, "aws://"):
		provider = "AWS"
	case strings.Contains(node.Spec.ProviderID, "gce://"):
		provider = "GKE"
	case strings.Contains(node.Spec.ProviderID, "azure://"):
		provider = "Azure"
	}
	return provider
}

func controlPlaneNodeCount(nodeList *v1.NodeList) int {
	controlPlaneNodes := 0
	for _, node := range nodeList.Items {
		if _, ok := node.ObjectMeta.Labels["node-role.kubernetes.io/master"]; ok {
			controlPlaneNodes++
		}
	}
	return controlPlaneNodes
}

func auditLoggingEnabled(client *kubernetes.Clientset) bool {
	podList, err := client.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{LabelSelector: "component=kube-apiserver"})
	if err != nil {
		return false
	}

	auditLogging := false
	for _, pod := range podList.Items {
		for _, container := range pod.Spec.Containers {
			for _, param := range container.Command {
				// Log file backend
				if strings.Contains(param, "audit-log-path") {
					auditLogging = true
				}
				// Webhook backend
				if strings.Contains(param, "audit-webhook-config-file") {
					auditLogging = true
				}
			}
		}
	}

	return auditLogging
}

func GetControlPlane(client *kubernetes.Clientset) ControlPlane {
	nodeList, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return ControlPlane{
			err: err,
		}
	}

	numControlPlaneNodes := controlPlaneNodeCount(nodeList)
	auditLogging := auditLoggingEnabled(client)

	return ControlPlane{
		Provider:        getProvider(nodeList.Items[0]),
		IsHA:            numControlPlaneNodes > 1,
		NumNodes:        numControlPlaneNodes,
		AuditLogEnabled: auditLogging,
	}
}
