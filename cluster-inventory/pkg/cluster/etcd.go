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
	"log"
	"net/url"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// TODO Is it even possible to query etcd/apiserver in the case where we are running on a managed cluster?
// Similar to how we deal with things in kube-bench -
// How we get the etcd information will depend on the node we are running on
/*
If we are on a control plane node, then we have access to the certs etc that are required to connect to the etcd server
If we are on a worker node - we don't have that. We can attempt to fall back to the basic etcd information, but it may not be that useful
Some of it was available through componentstatus but that is being deprecated

Potential next steps:
* Figure out how install etcdctl in the image
* Use a node selector and make two "versions" of the plugin which can do basic vs detailed etcd
* Ideas for detailed etcd check:
	- list members
	- version
	- performance check (can't print this in JSON but could include a string blob - better than nothing)
* If we don't have the detailed check available, fall back to attempting to find the etcd nodes from parsing the API server pods command lines
	- Check that this approach works with the kind HA cluster (where both API servers use localhost as the etcd address because it is on the same node)
*/

type EtcdStatus struct {
	AsPods         bool
	OnControlPlane bool
	NumOfNodes     int
}

func parseAddresses(list string) []string {
	etcdAddresses := []string{}
	step1 := strings.Split(list, "=")
	urls := []string{}
	if len(step1) > 1 {
		urls = strings.Split(step1[1], ",")
	}

	for _, address := range urls {
		etcdURL, err := url.Parse(address)
		if err != nil {
			fmt.Println("Error parsing ETCD address", err)
		}
		etcdAddresses = append(etcdAddresses, etcdURL.String())

	}
	return etcdAddresses
}

func getEtcdAddresses(client *kubernetes.Clientset, etcdInPods bool) ([]string, error) {
	addresses := []string{}
	if etcdInPods {
		pods, err := client.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{LabelSelector: "component=etcd"})
		if err != nil {
			// TODO handle error
			return addresses, err
		}

		// TODO: need to test for what this looks like when etcd isn't a pod still
		for _, pod := range pods.Items {
			addresses = append(addresses, pod.Spec.Hostname)
		}
	} else {
		// TODO: assumes kube-apiserver is a pod
		// probably assumption that can't remain.
		pods, err := client.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{LabelSelector: "component=kube-apiserver"})
		if err != nil {
			// handle error
			log.Println("Error looking for apiserver pod:", err)
		}

		if len(pods.Items) == 0 {
			// TODO we need to handle this properly
			log.Println("Could not find kube-apiserver pod")
			return addresses, nil
		}

		for _, args := range pods.Items[0].Spec.Containers[0].Command {
			if strings.Contains(args, "etcd-servers") {
				// parse out the addresses from the arg string
				addresses = parseAddresses(args)
				break
			}
		}
	}
	return addresses, nil
}

func contains(haystack []string, needle string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

// this will change later
func deployedOnControlPlaneNodes(client *kubernetes.Clientset, etcdNodes []string) bool {
	matches := 0
	nodeList, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		// TODO handle this error
		return false
	}

	for _, node := range nodeList.Items {
		if _, ok := node.ObjectMeta.Labels["node-role.kubernetes.io/master"]; ok {
			for _, address := range node.Status.Addresses {
				if contains(etcdNodes, address.Address) {
					matches++
				}
			}
		}
	}

	// if our number of matches is equal to number of nodes
	// we are all on masters
	return matches == len(etcdNodes)
}

func GetEtcdStatus(client *kubernetes.Clientset) (EtcdStatus, error) {
	es := EtcdStatus{}

	pods, err := client.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{LabelSelector: "component=etcd"})
	if err != nil {
		// TODO handle error
		log.Println("Error looking for etcd pod:", err)
	}

	if len(pods.Items) > 0 {
		es.AsPods = true
	}

	etcdAddresses, _ := getEtcdAddresses(client, es.AsPods)
	es.NumOfNodes = len(etcdAddresses)
	es.OnControlPlane = deployedOnControlPlaneNodes(client, etcdAddresses)
	return es, nil
}
