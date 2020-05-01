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
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
	rbac "k8s.io/api/rbac/v1"
)

// Result represents the result of a who-can query.
type Result struct {
	Resource  string          `json:"resource"`
	Verb      string          `json:"verb"`
	Namespace string          `json:"namespace"`
	Subjects  SubjectBindings `json:"subjects"`
}

type Results []Result

// Bindings represents RoleBindings or ClusterRoleBindings which may be applied to a subject.
type Bindings struct {
	RoleBindings        []string `json:"role-bindings,omitempty"`
	ClusterRoleBindings []string `json:"cluster-role-bindings,omitempty"`
}

// SubjectBindings represents the names of all Role and ClusterRole bindings bound to a subject.
type SubjectBindings map[rbac.Subject]Bindings

func (sb SubjectBindings) MarshalJSON() ([]byte, error) {
	var s []struct {
		rbac.Subject
		Bindings
	}
	for subject, bindings := range sb {
		s = append(s, struct {
			rbac.Subject
			Bindings
		}{
			Subject:  subject,
			Bindings: bindings,
		})
	}

	return json.Marshal(s)
}

// SubjectActionPermissions represents the role bindings that allow a particular subject to perform an action.
type SubjectActionPermissions struct {
	Resource string `json:"resource"`
	Verb     string `json:"verb"`
	Bindings
}

// SubjectNamespacePermissions represents all permissions granted within namespaces
type SubjectNamespacePermissions map[string][]SubjectActionPermissions

// MarshalJSON marshals SubjectNamespacePermissions into JSON
func (np SubjectNamespacePermissions) MarshalJSON() ([]byte, error) {
	var s []struct {
		Namespace                string                     `json:"namespace"`
		SubjectActionPermissions []SubjectActionPermissions `json:"actions"`
	}
	for namespace, results := range np {
		s = append(s, struct {
			Namespace                string                     `json:"namespace"`
			SubjectActionPermissions []SubjectActionPermissions `json:"actions"`
		}{
			Namespace:                namespace,
			SubjectActionPermissions: results,
		})
	}

	return json.Marshal(s)
}

// SubjectResults represents all the permissions for a subject
type SubjectResults map[rbac.Subject]SubjectNamespacePermissions

// MarshalJSON marshals SubjectResults into JSON
func (spm SubjectResults) MarshalJSON() ([]byte, error) {
	var s []struct {
		rbac.Subject
		Permissions SubjectNamespacePermissions `json:"permissions"`
	}
	for subject, val := range spm {
		s = append(s, struct {
			rbac.Subject
			Permissions SubjectNamespacePermissions `json:"permissions"`
		}{
			Subject:     subject,
			Permissions: val,
		})
	}

	return json.Marshal(s)
}

func resultsBySubject(results Results) SubjectResults {
	subjectResults := SubjectResults{}

	for _, result := range results {
		for subject, bindings := range result.Subjects {
			if _, ok := subjectResults[subject]; !ok {
				subjectResults[subject] = SubjectNamespacePermissions{}
			}
			actionPermissions := subjectResults[subject][result.Namespace]
			actionPermissions = append(actionPermissions, SubjectActionPermissions{
				Resource: result.Resource,
				Verb:     result.Verb,
				Bindings: bindings,
			})
			subjectResults[subject][result.Namespace] = actionPermissions
		}
	}
	return subjectResults
}

func (r *Results) WriteSubjectsReport(w io.Writer) error {
	subjects := resultsBySubject(*r)
	j, err := json.Marshal(subjects)
	if err != nil {
		return err
	}
	_, err = w.Write(j)
	return err
}

func (r *Results) WriteResourcesReport(w io.Writer) error {
	j, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = w.Write(j)
	return err
}

type SonobuoyResultsItem struct {
	Name     string                `json:"name" yaml:"name"`
	Status   string                `json:"status" yaml:"status,omitempty"`
	Metadata map[string]string     `json:"meta,omitempty" yaml:"meta,omitempty"`
	Details  map[string]string     `json:"details,omitempty" yaml:"details,omitempty"`
	Items    []SonobuoyResultsItem `json:"items,omitempty" yaml:"items,omitempty"`
}

func createSonobuoyResultsForSubjectBindings(subject rbac.Subject, bindings Bindings) SonobuoyResultsItem {
	subjectItem := SonobuoyResultsItem{
		Name: subject.Name,
		Details: map[string]string{
			"kind": subject.Kind,
		},
	}
	if subject.Namespace != "" {
		subjectItem.Details["namespace"] = subject.Namespace
	}

	if len(bindings.RoleBindings) != 0 {
		rbItem := SonobuoyResultsItem{
			Name: "rolebindings",
		}
		for _, rb := range bindings.RoleBindings {
			rbItem.Items = append(rbItem.Items, SonobuoyResultsItem{
				Name: rb,
			})
		}
		subjectItem.Items = append(subjectItem.Items, rbItem)
	}

	if len(bindings.ClusterRoleBindings) != 0 {
		crbItem := SonobuoyResultsItem{
			Name: "clusterrolebindings",
		}
		for _, crb := range bindings.ClusterRoleBindings {
			crbItem.Items = append(crbItem.Items, SonobuoyResultsItem{
				Name: crb,
			})
		}

		subjectItem.Items = append(subjectItem.Items, crbItem)
	}
	return subjectItem
}

func createSonobuoyResultsForResult(result Result) SonobuoyResultsItem {
	resultItem := SonobuoyResultsItem{
		Name: fmt.Sprintf("%v %v -n %v", result.Verb, result.Resource, result.Namespace),
		Details: map[string]string{
			"verb":      result.Verb,
			"resource":  result.Resource,
			"namespace": result.Namespace,
		},
	}
	for subject, bindings := range result.Subjects {
		resultItem.Items = append(resultItem.Items, createSonobuoyResultsForSubjectBindings(subject, bindings))
	}
	return resultItem
}

func (r *Results) WriteSonobuoyReport(w io.Writer) error {
	sonobuoyResults := SonobuoyResultsItem{
		Name:   "who-can",
		Status: "complete",
	}
	for _, result := range *r {
		sonobuoyResults.Items = append(sonobuoyResults.Items, createSonobuoyResultsForResult(result))
	}
	j, err := yaml.Marshal(sonobuoyResults)
	if err != nil {
		return err
	}
	_, err = w.Write(j)
	return err
}
