// Copyright 2019 GM Cruise LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modifications copyright (C) 2022 the Sonobuoy project contributors

package kube

import (
	"fmt"
	"path"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/restmapper"
)

type apiResource struct {
	GVK           schema.GroupVersionKind
	Name          string
	Namespace     string
	ClusterScoped bool
	Resource      string
	Subresource   string
}

// newResource discovers Resource mapping (only confirms apiVersion/Resource
// pair exists if apiGroup is provided) and returns new *apiResource.
func newResource(
	dClient discovery.DiscoveryInterface,
	name, namespace, apiGroup, resource, subresource string,
) (*apiResource, error) {
	gr, err := restmapper.GetAPIGroupResources(dClient)
	if err != nil {
		return nil, err
	}

	// Guess version from apiGroup. Version is optionally passed in as postfix of apiGroup after a `/`.
	version := ""
	apiGroupSplitted := strings.Split(apiGroup, "/")
	if len(apiGroupSplitted) > 1 {
		version = apiGroupSplitted[len(apiGroupSplitted)-1]
		apiGroup = strings.Join(apiGroupSplitted[:len(apiGroupSplitted)-1], "/")
	}

	partial := schema.GroupVersionResource{Group: apiGroup, Resource: resource, Version: version}
	rMapper := restmapper.NewDiscoveryRESTMapper(gr)

	gvk, err := rMapper.KindFor(partial)
	if err != nil {
		return nil, err
	}

	gvr, err := rMapper.ResourceFor(partial)
	if err != nil {
		return nil, err
	}

	r := &apiResource{
		GVK:         gvk,
		Name:        name,
		Namespace:   namespace,
		Resource:    gvr.Resource,
		Subresource: subresource,
	}
	return r.validate()
}

func newResourceForKind(
	dClient discovery.DiscoveryInterface,
	name, namespace, subresource string,
	gvk schema.GroupVersionKind,
) (*apiResource, error) {
	gr, err := restmapper.GetAPIGroupResources(dClient)
	if err != nil {
		return nil, err
	}

	mapping, err := restmapper.NewDiscoveryRESTMapper(gr).RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	r := &apiResource{
		GVK:           gvk,
		Name:          name,
		ClusterScoped: mapping.Scope.Name() == "root",
		Resource:      mapping.Resource.Resource,
		Subresource:   subresource,
	}
	if !r.ClusterScoped {
		r.Namespace = namespace
	}
	return r.validate()
}

func (r *apiResource) validate() (*apiResource, error) {
	if r.GVK.Kind == "Namespace" && r.Namespace != "" && r.Name != r.Namespace {
		return nil, fmt.Errorf("specified namespace `%s' doesn't match Namespace name: %v", r.Namespace, r)
	}

	return r, nil
}

func (r *apiResource) resourceSegments() []string {
	segments := []string{"/apis"}

	if r.GVK.Group == "" {
		segments = []string{"/api", r.GVK.Version}
	} else {
		segments = append(segments, r.GVK.Group, r.GVK.Version)
	}

	if r.Namespace != "" && !r.ClusterScoped {
		segments = append(segments, "namespaces", r.Namespace)
	}

	if r.Resource != "" { // Skip explicit resource for namespaces.
		segments = append(segments, r.Resource)
	}

	return segments
}

func (r *apiResource) String() string {
	return fmt.Sprintf("%s.%s `%s'", strings.ToLower(r.GVK.Kind), r.GVK.GroupVersion().String(), maybeNamespaced(r.Name, r.Namespace))
}

func (r *apiResource) GroupVersionResource() schema.GroupVersionResource {
	return r.GVK.GroupVersion().WithResource(r.Resource)
}

func (r *apiResource) Path() string {
	return path.Join(r.resourceSegments()...)
}

func (r *apiResource) PathWithName() string {
	p := r.Path()

	if r.Name != "" {
		p = path.Join(p, r.Name)
	}

	return p
}

func (r *apiResource) PathWithSubresource() string {
	p := r.PathWithName()

	if r.Subresource != "" {
		p = path.Join(p, r.Subresource)
	}

	return p
}
