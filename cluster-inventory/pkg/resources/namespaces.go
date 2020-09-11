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
	"fmt"
	"github.com/vmware-tanzu/sonobuoy-plugins/cluster-inventory/pkg/reports"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Namespaces []Namespace

type Namespace struct {
	v1.Namespace
	quotas []v1.ResourceQuota
	limits []v1.LimitRange
}

func GetNamespaces(client *kubernetes.Clientset) []Namespace {
	nsList, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("could not fetch namespaces")
	}

	namespaces := []Namespace{}
	for _, ns := range nsList.Items {
		namespace := Namespace{Namespace: ns}

		limitRangeList, err := client.CoreV1().LimitRanges(ns.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Println("could not fetch limit ranges for ", ns.Name)
		} else {
			namespace.limits = limitRangeList.Items
		}

		resourceQuotaList, err := client.CoreV1().ResourceQuotas(ns.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Println("could not fetch resource quotas for ", ns.Name)
		} else {
			namespace.quotas = resourceQuotaList.Items
		}

		namespaces = append(namespaces, namespace)
	}
	return namespaces
}

type quota map[string]map[string]string

func parseQuota(rq v1.ResourceQuota) quota {
	q := quota{}

	for name, quantity := range rq.Status.Hard {
		if _, ok := q[name.String()]; !ok {
			q[name.String()] = map[string]string{}
		}
		q[name.String()]["limit"] = quantity.String()
	}

	for name, quantity := range rq.Status.Used {
		if _, ok := q[name.String()]; !ok {
			q[name.String()] = map[string]string{}
		}
		q[name.String()]["used"] = quantity.String()
	}

	return q
}

type limitRangeItem map[string]interface{}

func parseLimitRange(lri v1.LimitRangeItem) limitRangeItem {
	item := limitRangeItem{}
	item["type"] = lri.Type

	processResourceList := func(resourceList v1.ResourceList) map[string]string {
		resourceLimits := map[string]string{}
		for r, q := range resourceList {
			resourceLimits[r.String()] = q.String()
		}
		return resourceLimits
	}

	if len(lri.Default) > 0 {
		item["default"] = processResourceList(lri.Default)
	}

	if len(lri.DefaultRequest) > 0 {
		item["defaultRequest"] = processResourceList(lri.DefaultRequest)
	}

	if len(lri.Min) > 0 {
		item["min"] = processResourceList(lri.Min)
	}

	if len(lri.Max) > 0 {
		item["max"] = processResourceList(lri.Max)
	}

	if len(lri.MaxLimitRequestRatio) > 0 {
		item["maxLimitRequestRatio"] = processResourceList(lri.MaxLimitRequestRatio)
	}

	return item
}

func (n Namespace) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name:    n.Name,
		Status:  string(n.Status.Phase),
		Details: map[string]interface{}{},
	}

	if len(n.quotas) > 0 {
		quotas := map[string]quota{}
		for _, q := range n.quotas {
			quotas[q.Name] = parseQuota(q)
		}
		item.Details["resourceQuotas"] = quotas
	}

	if len(n.limits) > 0 {
		limits := map[string][]limitRangeItem{}
		for _, l := range n.limits {
			limits[l.Name] = []limitRangeItem{}
			for _, limit := range l.Spec.Limits {
				limits[l.Name] = append(limits[l.Name], parseLimitRange(limit))
			}
		}
		item.Details["limitRanges"] = limits
	}

	return item
}

func (n Namespaces) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name:   "Namespaces",
		Status: "complete",
	}

	for _, ns := range n {
		item.Items = append(item.Items, ns.GenerateSonobuoyItem())
	}

	return item
}
