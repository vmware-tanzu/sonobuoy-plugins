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

package whocan

import (
	whocancmd "github.com/aquasecurity/kubectl-who-can/pkg/cmd"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

// Checker is the interface for running a whocan Action
type Checker interface {
	// Check takes a whocan Action and returns the role bindings that allow that action to be performed.
	Check(whocancmd.Action) (roleBindings []rbac.RoleBinding, clusterRoleBindings []rbac.ClusterRoleBinding, err error)
}

// NewChecker creates a client which can run who-can queries.
func NewChecker(client *kubernetes.Clientset, restConfig *rest.Config) (Checker, error) {
	discoveryClient := memory.NewMemCacheClient(client)
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient)

	return whocancmd.NewWhoCan(restConfig, expander)
}
