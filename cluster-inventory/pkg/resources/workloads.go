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

type WorkloadsTree struct {
	Deployments            map[string]*Deployment
	ReplicaSets            map[string]*ReplicaSet
	ReplicationControllers map[string]*ReplicationController
	StatefulSets           map[string]*StatefulSet
	DaemonSets             map[string]*DaemonSet
	Jobs                   map[string]*Job
	CronJobs               map[string]*CronJob
	Pods                   map[string]*Pod

	client    *kubernetes.Clientset
	namespace string
}

func (w WorkloadsTree) GenerateSonobuoyItem() reports.SonobuoyResultsItem {
	item := reports.SonobuoyResultsItem{
		Name:   w.namespace,
		Status: "complete",
		Metadata: map[string]string{
			"kind": "Namespace",
		},
	}

	if len(w.Deployments) > 0 {
		deployments := reports.SonobuoyResultsItem{
			Name:   "Deployments",
			Status: "complete",
		}
		for _, deployment := range w.Deployments {
			deployments.Items = append(deployments.Items, deployment.GenerateSonobuoyItem())
		}

		item.Items = append(item.Items, deployments)
	}

	if len(w.ReplicaSets) > 0 {
		replicaSets := reports.SonobuoyResultsItem{
			Name:   "Replica Sets",
			Status: "complete",
		}
		for _, replicaset := range w.ReplicaSets {
			replicaSets.Items = append(replicaSets.Items, replicaset.GenerateSonobuoyItem())
		}

		item.Items = append(item.Items, replicaSets)
	}

	if len(w.ReplicationControllers) > 0 {
		replicationControllers := reports.SonobuoyResultsItem{
			Name:   "Replication Controllers",
			Status: "complete",
		}
		for _, rc := range w.ReplicationControllers {
			replicationControllers.Items = append(replicationControllers.Items, rc.GenerateSonobuoyItem())
		}

		item.Items = append(item.Items, replicationControllers)
	}

	if len(w.StatefulSets) > 0 {
		statefulSets := reports.SonobuoyResultsItem{
			Name:   "Stateful Sets",
			Status: "complete",
		}
		for _, ss := range w.StatefulSets {
			statefulSets.Items = append(statefulSets.Items, ss.GenerateSonobuoyItem())
		}

		item.Items = append(item.Items, statefulSets)
	}

	if len(w.DaemonSets) > 0 {
		daemonSets := reports.SonobuoyResultsItem{
			Name:   "Daemon Sets",
			Status: "complete",
		}
		for _, ds := range w.DaemonSets {
			daemonSets.Items = append(daemonSets.Items, ds.GenerateSonobuoyItem())
		}

		item.Items = append(item.Items, daemonSets)
	}

	if len(w.CronJobs) > 0 {
		cronJobs := reports.SonobuoyResultsItem{
			Name:   "Cron Jobs",
			Status: "complete",
		}
		for _, cj := range w.CronJobs {
			cronJobs.Items = append(cronJobs.Items, cj.GenerateSonobuoyItem())
		}

		item.Items = append(item.Items, cronJobs)
	}

	if len(w.Jobs) > 0 {
		jobs := reports.SonobuoyResultsItem{
			Name:   "Jobs",
			Status: "complete",
		}
		for _, j := range w.Jobs {
			jobs.Items = append(jobs.Items, j.GenerateSonobuoyItem())
		}

		item.Items = append(item.Items, jobs)
	}

	if len(w.Pods) > 0 {
		pods := reports.SonobuoyResultsItem{
			Name:   "Pods",
			Status: "complete",
		}
		for _, p := range w.Pods {
			pods.Items = append(pods.Items, p.GenerateSonobuoyItem())
		}

		item.Items = append(item.Items, pods)
	}

	return item
}

func (w *WorkloadsTree) addPods() error {
	podList, err := w.client.CoreV1().Pods(w.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, pod := range podList.Items {
		if _, ok := w.Pods[pod.Name]; !ok {
			w.Pods[pod.Name] = &Pod{pod}
		}
	}
	return nil
}

func (w *WorkloadsTree) addDeployments() error {
	deploymentList, err := w.client.AppsV1().Deployments(w.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, deployment := range deploymentList.Items {
		if _, ok := w.Deployments[deployment.Name]; !ok {
			w.Deployments[deployment.Name] = &Deployment{
				Deployment:  deployment,
				ReplicaSets: map[string]*ReplicaSet{},
			}
		}
	}
	return nil
}

func (w *WorkloadsTree) addReplicaSets() error {
	replicaSetList, err := w.client.AppsV1().ReplicaSets(w.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, replicaSet := range replicaSetList.Items {
		if _, ok := w.ReplicaSets[replicaSet.Name]; !ok {
			w.ReplicaSets[replicaSet.Name] = &ReplicaSet{
				ReplicaSet: replicaSet,
				Pods:       map[string]*Pod{},
			}
		}
	}
	return nil
}

func (w *WorkloadsTree) addReplicationControllers() error {
	replicationControllerList, err := w.client.CoreV1().ReplicationControllers(w.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, replicationController := range replicationControllerList.Items {
		if _, ok := w.ReplicationControllers[replicationController.Name]; !ok {
			w.ReplicationControllers[replicationController.Name] = &ReplicationController{
				ReplicationController: replicationController,
				Pods:                  map[string]*Pod{},
			}
		}
	}
	return nil
}

func (w *WorkloadsTree) addStatefulSets() error {
	statefulSetsList, err := w.client.AppsV1().StatefulSets(w.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, statefulSet := range statefulSetsList.Items {
		if _, ok := w.StatefulSets[statefulSet.Name]; !ok {
			w.StatefulSets[statefulSet.Name] = &StatefulSet{
				StatefulSet: statefulSet,
				Pods:        map[string]*Pod{},
			}
		}
	}
	return nil
}

func (w *WorkloadsTree) addDaemonSets() error {
	daemonSetList, err := w.client.AppsV1().DaemonSets(w.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, daemonSet := range daemonSetList.Items {
		if _, ok := w.DaemonSets[daemonSet.Name]; !ok {
			w.DaemonSets[daemonSet.Name] = &DaemonSet{
				DaemonSet: daemonSet,
				Pods:      map[string]*Pod{},
			}
		}
	}
	return nil
}

func (w *WorkloadsTree) addJobs() error {
	jobList, err := w.client.BatchV1().Jobs(w.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, job := range jobList.Items {
		if _, ok := w.Jobs[job.Name]; !ok {
			w.Jobs[job.Name] = &Job{
				Job:  job,
				Pods: map[string]*Pod{},
			}
		}
	}
	return nil
}

func (w *WorkloadsTree) addCronJobs() error {
	cronJobList, err := w.client.BatchV1beta1().CronJobs(w.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, cronJob := range cronJobList.Items {
		if _, ok := w.CronJobs[cronJob.Name]; !ok {
			w.CronJobs[cronJob.Name] = &CronJob{
				CronJob: cronJob,
				Jobs:    map[string]*Job{},
			}
		}
	}
	return nil
}

func (w *WorkloadsTree) resolveControllerOwnerReferences() {
	for _, pod := range w.Pods {
		controller := metav1.GetControllerOf(&pod.Pod)
		if controller != nil {
			switch controller.Kind {
			case "ReplicaSet":
				rs, ok := w.ReplicaSets[controller.Name]
				if ok && rs.UID == controller.UID {
					rs.Pods[pod.Name] = pod
					delete(w.Pods, pod.Name)
					w.ReplicaSets[controller.Name] = rs
				}
			case "ReplicationController":
				rc, ok := w.ReplicationControllers[controller.Name]
				if ok && rc.UID == controller.UID {
					rc.Pods[pod.Name] = pod
					delete(w.Pods, pod.Name)
					w.ReplicationControllers[controller.Name] = rc
				}
			case "DaemonSet":
				ds, ok := w.DaemonSets[controller.Name]
				if ok && ds.UID == controller.UID {
					ds.Pods[pod.Name] = pod
					delete(w.Pods, pod.Name)
					w.DaemonSets[controller.Name] = ds
				}
			case "StatefulSet":
				ss, ok := w.StatefulSets[controller.Name]
				if ok && ss.UID == controller.UID {
					ss.Pods[pod.Name] = pod
					delete(w.Pods, pod.Name)
					w.StatefulSets[controller.Name] = ss
				}
			case "Job":
				job, ok := w.Jobs[controller.Name]
				if ok && job.UID == controller.UID {
					job.Pods[pod.Name] = pod
					delete(w.Pods, pod.Name)
					w.Jobs[controller.Name] = job
				}
			}
		}
	}

	for _, rs := range w.ReplicaSets {
		controller := metav1.GetControllerOf(&rs.ReplicaSet)
		if controller != nil {
			switch controller.Kind {
			case "Deployment":
				d, ok := w.Deployments[controller.Name]
				if ok && d.UID == controller.UID {
					d.ReplicaSets[rs.Name] = rs
					delete(w.ReplicaSets, rs.Name)
					w.Deployments[controller.Name] = d
				}
			}
		}
	}

	for _, job := range w.Jobs {
		controller := metav1.GetControllerOf(&job.Job)
		if controller != nil {
			switch controller.Kind {
			case "CronJob":
				cj, ok := w.CronJobs[controller.Name]
				if ok && cj.UID == controller.UID {
					cj.Jobs[job.Name] = job
					delete(w.Jobs, job.Name)
					w.CronJobs[controller.Name] = cj
				}
			}
		}
	}
}

func (w *WorkloadsTree) Populate() error {
	if err := w.addPods(); err != nil {
		return err
	}

	if err := w.addDeployments(); err != nil {
		return err
	}

	if err := w.addReplicaSets(); err != nil {
		return err
	}

	if err := w.addReplicationControllers(); err != nil {
		return err
	}

	if err := w.addStatefulSets(); err != nil {
		return err
	}

	if err := w.addDaemonSets(); err != nil {
		return err
	}

	if err := w.addJobs(); err != nil {
		return err
	}

	if err := w.addCronJobs(); err != nil {
		return err
	}

	w.resolveControllerOwnerReferences()
	return nil
}

func NewWorkloadsTree(client *kubernetes.Clientset, namespace string) *WorkloadsTree {
	return &WorkloadsTree{
		Deployments:            map[string]*Deployment{},
		ReplicaSets:            map[string]*ReplicaSet{},
		ReplicationControllers: map[string]*ReplicationController{},
		StatefulSets:           map[string]*StatefulSet{},
		DaemonSets:             map[string]*DaemonSet{},
		Jobs:                   map[string]*Job{},
		CronJobs:               map[string]*CronJob{},
		Pods:                   map[string]*Pod{},

		client:    client,
		namespace: namespace,
	}
}

func getWorkloadsForNamespace(client *kubernetes.Clientset, namespace string) (*WorkloadsTree, error) {
	workloads := NewWorkloadsTree(client, namespace)

	if err := workloads.Populate(); err != nil {
		return nil, err
	}

	return workloads, nil
}
