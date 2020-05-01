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
	"strings"

	whocancmd "github.com/aquasecurity/kubectl-who-can/pkg/cmd"
	"github.com/pkg/errors"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// createAction inspects the given resource and verb and creates the Action necessary for a
// who-can query.
func createAction(resource, verb, namespace string) whocancmd.Action {
	// Determine if the resource type is a subresource based on the name form resource/subresource.
	// If the resource begins with "/", leave as is as a non-resource URL, otherwise attempt to split.
	// TODO we're not handling non-resource URLs here. Look up how they are handled in kubectl-who-can.
	var subResource string
	if !strings.HasPrefix(resource, "/") {
		resourceTokens := strings.SplitN(resource, "/", 2)
		resource = resourceTokens[0]
		if len(resourceTokens) > 1 {
			subResource = resourceTokens[1]
		}
	}

	allNamespaces := false
	if namespace == "*" {
		allNamespaces = true
		namespace = ""
	}

	return whocancmd.Action{
		Verb:          verb,
		Resource:      resource,
		SubResource:   subResource,
		Namespace:     namespace,
		AllNamespaces: allNamespaces,
	}
}

func createActions(namespaces []string, resources []metav1.APIResource) []whocancmd.Action {
	var actions []whocancmd.Action
	for _, namespace := range namespaces {
		for _, resource := range resources {
			for _, verb := range resource.Verbs {
				actions = append(actions, createAction(resource.Name, verb, namespace))
			}
		}
	}
	return actions
}

type Runner struct {
	checker Checker
}

func NewRunner(checker Checker) Runner {
	return Runner{
		checker: checker,
	}
}

func createResult(action whocancmd.Action, rbs []rbac.RoleBinding, crbs []rbac.ClusterRoleBinding) Result {
	resource := action.Resource
	if action.SubResource != "" {
		resource += "/" + action.SubResource
	}

	result := Result{
		Resource: resource,
		Verb:     action.Verb,
	}

	if action.AllNamespaces {
		result.Namespace = "*"
	} else {
		result.Namespace = action.Namespace
	}

	subjects := SubjectBindings{}
	for _, rb := range rbs {
		for _, subject := range rb.Subjects {
			ref := subjects[subject]
			ref.RoleBindings = append(ref.RoleBindings, rb.Name)
			subjects[subject] = ref
		}
	}
	for _, crb := range crbs {
		for _, subject := range crb.Subjects {
			ref := subjects[subject]
			ref.ClusterRoleBindings = append(ref.ClusterRoleBindings, crb.Name)
			subjects[subject] = ref
		}
	}
	result.Subjects = subjects

	return result
}

func (r *Runner) Run(namespaces []string, resources []metav1.APIResource) (Results, error) {
	actions := createActions(namespaces, resources)
	results := Results{}
	for _, action := range actions {
		rbs, crbs, err := r.checker.Check(action)
		if err != nil {
			return Results{}, errors.Wrap(err, "running checker")
		}

		results = append(results, createResult(action, rbs, crbs))
	}

	return results, nil
}
