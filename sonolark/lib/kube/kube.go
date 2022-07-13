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

// Package kube implements "kube" starlark built-in which renders and applies
// Kubernetes objects.
// Modifications copyright (C) 2022 the Sonobuoy project contributors

package kube

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cruise-automation/isopod/pkg/addon"
	log "github.com/golang/glog"
	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/yaml"
)

const (
	namespaceResrc = "namespace"
	apiGroupKW     = "api_group"
)

// kubePackage implements Kubernetes package that can be imported by plugin
// code.
type kubePackage struct {
	dClient     discovery.DiscoveryInterface
	dynClient   dynamic.Interface
	httpClient  *http.Client
	config      *rest.Config
	dryRun      bool
	force       bool
	diff        bool
	diffFilters []string
	// host:port of the master endpoint.
	Master string
}

// KubeNoop returns a new stringDict with noop methods.
func KubeNoop() starlark.StringDict {
	return starlark.StringDict{
		"kube": &starlarkstruct.Module{
			Name: "kube",
			Members: starlark.StringDict{
				kubeDeleteMethod:           starlark.NewBuiltin("kube."+kubeDeleteMethod, NoOp),
				kubeResourceQuantityMethod: starlark.NewBuiltin("kube."+kubeResourceQuantityMethod, NoOp),
				kubePutMethod:              starlark.NewBuiltin("kube."+kubePutMethod, NoOp),
				kubeExistsMethod:           starlark.NewBuiltin("kube."+kubeExistsMethod, NoOp),
				kubeGetMethod:              starlark.NewBuiltin("kube."+kubeGetMethod, NoOp),
				kubeFromStrMethod:          starlark.NewBuiltin("kube."+kubeFromStrMethod, NoOp),
				kubeFromIntMethod:          starlark.NewBuiltin("kube."+kubeFromIntMethod, NoOp),
			},
		},
	}
}

func NoOp(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.None, nil
}

// New returns a new StringDict.
func New(
	addr string,
	d discovery.DiscoveryInterface,
	dynC dynamic.Interface,
	c *http.Client,
	config *rest.Config,
	dryRun, force, diff bool,
	diffFilters []string,
) starlark.StringDict {

	pkg := &kubePackage{
		dClient:     d,
		dynClient:   dynC,
		httpClient:  c,
		config:      config,
		Master:      addr,
		dryRun:      dryRun,
		force:       force,
		diff:        diff,
		diffFilters: diffFilters,
	}

	return starlark.StringDict{
		"kube": &starlarkstruct.Module{
			Name: "kube",
			Members: starlark.StringDict{
				kubeDeleteMethod:           starlark.NewBuiltin("kube."+kubeDeleteMethod, pkg.kubeDeleteFn),
				kubeResourceQuantityMethod: starlark.NewBuiltin("kube."+kubeResourceQuantityMethod, resourceQuantityFn),
				kubePutMethod:              starlark.NewBuiltin("kube."+kubePutMethod, pkg.kubePutFn),
				kubeExistsMethod:           starlark.NewBuiltin("kube."+kubeExistsMethod, pkg.kubeExistsFn),
				kubeGetMethod:              starlark.NewBuiltin("kube."+kubeGetMethod, pkg.kubeGetFn),
				kubeFromStrMethod:          starlark.NewBuiltin("kube."+kubeFromStrMethod, fromStringFn),
				kubeFromIntMethod:          starlark.NewBuiltin("kube."+kubeFromIntMethod, fromIntFn),
				kubeDiffMethod:             starlark.NewBuiltin("kube."+kubeDiffMethod, kubeDiffFn),
				kubePortForwardMethod:      starlark.NewBuiltin("kube."+kubePortForwardMethod, pkg.kubePortForwardTestFn),
			},
		},
	}
}

const (
	kubeDeleteMethod           = "delete"
	kubeFromIntMethod          = "from_int"
	kubeFromStrMethod          = "from_str"
	kubeGetMethod              = "get"
	kubeExistsMethod           = "exists"
	kubePutMethod              = "put"
	kubePutYamlMethod          = "put_yaml"
	kubeResourceQuantityMethod = "resource_quantity"
	kubeDiffMethod             = "diff"
	kubePortForwardMethod      = "portforward"
)

// setMetadata sets metadata fields on the obj.
func (m *kubePackage) setMetadata(name, namespace string, obj runtime.Object) error {
	a := meta.NewAccessor()

	objName, err := a.Name(obj)
	if err != nil {
		return err
	}
	if objName != "" && objName != name {
		return fmt.Errorf("name=`%s' argument does not match object's .metadata.name=`%s'", name, objName)
	}
	if err := a.SetName(obj, name); err != nil {
		return err
	}

	if namespace != "" { // namespace is optional argument.
		objNs, err := a.Namespace(obj)
		if err != nil {
			return err
		}
		if objNs != "" && objNs != namespace {
			return fmt.Errorf("namespace=`%s' argument does not match object's .metadata.namespace=`%s'", namespace, objNs)
		}

		if err := a.SetNamespace(obj, namespace); err != nil {
			return err
		}
	}

	ls, err := a.Labels(obj)
	if err != nil {
		return err
	}
	if ls == nil {
		ls = map[string]string{}
	}
	if err := a.SetLabels(obj, ls); err != nil {
		return err
	}

	as, err := a.Annotations(obj)
	if err != nil {
		return err
	}
	if as == nil {
		as = map[string]string{}
	}

	return a.SetAnnotations(obj, as)
}

func getResourceAndName(resArg starlark.Tuple) (resource, name string, err error) {
	resourceArg, ok := resArg[0].(starlark.String)
	if !ok {
		err = errors.New("expected string for resource")
		return
	}
	resource = string(resourceArg)
	nameArg, ok := resArg[1].(starlark.String)
	if !ok {
		err = errors.New("expected string for resource name")
		return
	}
	name = string(nameArg)
	return
}

// kubePutFn is entry point for `kube.put' callable.
func (m *kubePackage) kubePutFn(t *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name, namespace string
	data := &starlark.List{}
	unpacked := []interface{}{
		"name", &name,
		"data", &data,
		"namespace?", &namespace,
	}
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, unpacked...); err != nil {
		return nil, fmt.Errorf("<%v>: %v", b.Name(), err)
	}

	val, err := m.Apply(t, name, namespace, data)
	if err != nil {
		return nil, fmt.Errorf("<%v>: %v", b.Name(), err)
	}

	return val, nil
}

// kubeDeleteFn is entry point for `kube.delete' callable.
// TODO(dmitry-ilyevskiy): Return Status object from the response as Starlark dict.
func (m *kubePackage) kubeDeleteFn(t *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("<%v>: positional args not supported by `kube.delete': %v", b.Name(), args)
	}

	if len(kwargs) < 1 {
		return nil, fmt.Errorf("<%v>: expected at least <resource>=<name>", b.Name())
	}

	resource, name, err := getResourceAndName(kwargs[0])
	if err != nil {
		return nil, fmt.Errorf("<%v>: %s", b.Name(), err.Error())
	}

	// If resource is not namespace itself (special case) attempt to parse
	// namespace out of the arg value.
	var namespace string
	if resource != namespaceResrc {
		ss := strings.Split(name, "/")
		if len(ss) > 1 {
			namespace = ss[0]
			name = ss[1]
		}
	}

	// Optional api_group argument.
	var apiGroup starlark.String
	var foreground starlark.Bool
	for _, kv := range kwargs[1:] {
		switch string(kv[0].(starlark.String)) {
		case apiGroupKW:
			var ok bool
			if apiGroup, ok = kv[1].(starlark.String); !ok {
				return nil, fmt.Errorf("<%v>: expected string value for `%s' arg, got: %s", b.Name(), apiGroupKW, kv[1].Type())
			}
		case "foreground":
			var ok bool
			if foreground, ok = kv[1].(starlark.Bool); !ok {
				return nil, fmt.Errorf("<%v>: expected string value for `foreground' arg, got: %s", b.Name(), kv[1].Type())
			}
		default:
			return nil, fmt.Errorf("<%v>: expected `api_group' or `foreground', got: %v=%v", b.Name(), kv[0], kv[1])
		}
	}

	r, err := newResource(m.dClient, name, namespace, string(apiGroup), resource, "")
	if err != nil {
		return nil, fmt.Errorf("<%v>: failed to map resource: %v", b.Name(), err)
	}

	ctx := t.Local(addon.GoCtxKey).(context.Context)
	if err := m.kubeDelete(ctx, r, bool(foreground)); err != nil {
		return nil, fmt.Errorf("<%v>: %v", b.Name(), err)
	}

	return starlark.None, nil
}

// kubeGetFn is an entry point for `kube.get` built-in.
func (m *kubePackage) kubeGetFn(t *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("<%v>: positional args not supported: %v", b.Name(), args)
	}

	if len(kwargs) < 1 {
		return nil, fmt.Errorf("<%v>: expected <resource>=<name>", b.Name())
	}

	resource, name, err := getResourceAndName(kwargs[0])
	if err != nil {
		return nil, fmt.Errorf("<%v>: %s", b.Name(), err.Error())
	}

	// If resource is not namespace itself (special case), attempt to parse
	// namespace out of the arg value.
	var namespace string
	if resource != namespaceResrc {
		ss := strings.Split(name, "/")
		if len(ss) > 1 {
			namespace = ss[0]
			name = ss[1]
		}
	}

	// Optional api_group argument.
	var apiGroup starlark.String
	var wait = 30 * time.Second
	var wantJSON bool
	//wip
	_ = wantJSON
	for _, kv := range kwargs[1:] {
		switch string(kv[0].(starlark.String)) {
		case apiGroupKW:
			var ok bool
			if apiGroup, ok = kv[1].(starlark.String); !ok {
				return nil, fmt.Errorf("<%v>: expected string value for `%s' arg, got: %s", b.Name(), apiGroupKW, kv[1].Type())
			}
		case "wait":
			durStr, ok := kv[1].(starlark.String)
			if !ok {
				return nil, fmt.Errorf("<%v>: expected string value for `wait' arg, got: %s", b.Name(), kv[1].Type())
			}

			var err error
			if wait, err = time.ParseDuration(string(durStr)); err != nil {
				return nil, fmt.Errorf("<%v>: failed to parse duration value: %v", b.Name(), err)
			}
		case "json":
			bv, ok := kv[1].(starlark.Bool)
			if !ok {
				return nil, fmt.Errorf("<%v>: expected boolean value for `json' arg, got: %s", b.Name(), kv[1].Type())
			}
			wantJSON = bool(bv)
		default:
			return nil, fmt.Errorf("<%v>: expected one of [ api_group | wait | json ] args, got: %v=%v", b.Name(), kv[0], kv[1])
		}
	}

	r, err := newResource(m.dClient, name, namespace, string(apiGroup), resource, "")
	if err != nil {
		return nil, fmt.Errorf("<%v>: failed to map resource: %v", b.Name(), err)
	}

	ctx := t.Local(addon.GoCtxKey).(context.Context)
	obj, err := m.kubeGet(ctx, r, wait)
	if err != nil {
		return nil, fmt.Errorf("<%v>: failed to get %s%s `%s': %v", b.Name(), resource, maybeCore(string(apiGroup)), name, err)
	}

	bits, err := renderObj(obj, nil, true, m.diffFilters)
	if err != nil {
		panic(err)
	}
	return starlark.String(string(bits)), nil
}

// kubeExistsFn is an entry point for `kube.exists` built-in.
func (m *kubePackage) kubeExistsFn(t *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 0 {
		return starlark.False, fmt.Errorf("<%v>: positional args not supported: %v", b.Name(), args)
	}

	if len(kwargs) < 1 {
		return starlark.False, fmt.Errorf("<%v>: expected <resource>=<name>", b.Name())
	}

	resource, name, err := getResourceAndName(kwargs[0])
	if err != nil {
		return nil, fmt.Errorf("<%v>: %s", b.Name(), err.Error())
	}

	// If resource is not namespace itself (special case), attempt to parse
	// namespace out of the arg value.
	var namespace string
	if resource != namespaceResrc {
		ss := strings.Split(name, "/")
		if len(ss) > 1 {
			namespace = ss[0]
			name = ss[1]
		}
	}

	// Optional api_group argument.
	var apiGroup starlark.String
	var wait time.Duration
	for _, kv := range kwargs[1:] {
		switch string(kv[0].(starlark.String)) {
		case apiGroupKW:
			var ok bool
			if apiGroup, ok = kv[1].(starlark.String); !ok {
				return starlark.False, fmt.Errorf("<%v>: expected string value for `%s' arg, got: %s", b.Name(), apiGroupKW, kv[1].Type())
			}
		case "wait":
			durStr, ok := kv[1].(starlark.String)
			if !ok {
				return starlark.False, fmt.Errorf("<%v>: expected string value for `wait' arg, got: %s", b.Name(), kv[1].Type())
			}

			var err error
			if wait, err = time.ParseDuration(string(durStr)); err != nil {
				return starlark.False, fmt.Errorf("<%v>: failed to parse duration value: %v", b.Name(), err)
			}
		default:
			return starlark.False, fmt.Errorf("<%v>: expected one of [ api_group | wait ] args, got: %v=%v", b.Name(), kv[0], kv[1])
		}
	}

	r, err := newResource(m.dClient, name, namespace, string(apiGroup), resource, "")
	if err != nil {
		return starlark.False, fmt.Errorf("<%v>: failed to map resource: %v", b.Name(), err)
	}

	ctx := t.Local(addon.GoCtxKey).(context.Context)
	_, err = m.kubeGet(ctx, r, wait)
	if err == ErrNotFound {
		return starlark.False, nil
	} else if err != nil {
		return starlark.False, err
	}

	return starlark.True, nil
}

var decodeFn = Codecs.UniversalDeserializer().Decode

func decode(raw []byte) (runtime.Object, *schema.GroupVersionKind, error) {
	obj, gvk, err := decodeFn(raw, nil, nil)
	if err == nil {
		return obj, gvk, nil
	}
	if !runtime.IsNotRegisteredError(err) {
		return nil, nil, err
	}

	// When the input is already a json, this just returns it as-is.
	j, err := yaml.YAMLToJSON(raw)
	if err != nil {
		return nil, nil, err
	}

	return unstructured.UnstructuredJSONScheme.Decode(j, nil, nil)
}

// parseHTTPResponse parses response body to extract runtime.Object
// and HTTP return code.
// Returns details message on success and error on failure (includes HTTP
// response codes not in 2XX).
func parseHTTPResponse(r *http.Response) (obj runtime.Object, details string, err error) {
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read body (response code: %d): %v", r.StatusCode, err)
	}

	log.V(2).Infof("Response raw data: %s", raw)
	obj, gvk, err := decode(raw)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse json object (response code: %d): %v", r.StatusCode, err)
	}

	if r.StatusCode < 200 || r.StatusCode >= 300 {
		return nil, "", fmt.Errorf("%s (response code: %d)", apierrors.FromObject(obj).Error(), r.StatusCode)
	}

	if s, ok := obj.(*metav1.Status); ok {
		d := s.Details
		if d == nil {
			return obj, s.Message, nil
		}
		return obj, fmt.Sprintf("%s%s `%s", d.Kind, d.Group, d.Name), nil
	}

	if in, ok := obj.(metav1.Object); ok {
		return obj, fmt.Sprintf("%s%s `%s'", strings.ToLower(gvk.Kind), maybeCore(gvk.Group), maybeNamespaced(in.GetName(), in.GetNamespace())), nil
	}
	if _, ok := obj.(metav1.ListInterface); ok {
		return obj, fmt.Sprintf("%s%s'", strings.ToLower(gvk.Kind), maybeCore(gvk.Group)), nil
	}
	return nil, "", fmt.Errorf("returned object does not implement `metav1.Object` or `metav1.ListInterface`: %v", obj)
}

// kubePeek checks if object by url exists in Kubernetes.
func (m *kubePackage) kubePeek(ctx context.Context, url string) (obj runtime.Object, found bool, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}

	log.V(1).Infof("GET to %s", url)

	resp, err := m.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, false, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, false, nil
	}

	obj, _, err = parseHTTPResponse(resp)
	if err != nil {
		return nil, false, err
	}
	return obj, true, nil
}

var ErrUpdateImmutable = errors.New("cannot update immutable. Use -force to delete and recreate")

func ErrImmutableRessource(attribute string, obj runtime.Object) error {
	return fmt.Errorf("failed to update %s of resource %s: %w", attribute, obj.GetObjectKind().GroupVersionKind().String(), ErrUpdateImmutable)
}

// mergeObjects merges the fields from the live object to the new
// object such as resource version and clusterIP.
// TODO(jon.yucel): Instead of selectively picking fields, holisticly
// solving this problem requires three-way merge implementation.
func mergeObjects(live, obj runtime.Object) error {
	// Service's clusterIP needs to be re-set to the value provided
	// by controller or mutation will be denied.
	if liveSvc, ok := live.(*corev1.Service); ok {
		svc := obj.(*corev1.Service)
		svc.Spec.ClusterIP = liveSvc.Spec.ClusterIP

		gotPort := liveSvc.Spec.HealthCheckNodePort
		wantPort := svc.Spec.HealthCheckNodePort
		// If port is set (non-zero) and doesn't match the existing port (also non-zero), error out.
		if wantPort != 0 && gotPort != 0 && wantPort != gotPort {
			return ErrImmutableRessource(".spec.healthCheckNodePort", obj)
		}
		svc.Spec.HealthCheckNodePort = gotPort
	}

	if liveClusterRoleBinding, ok := live.(*rbacv1.ClusterRoleBinding); ok {
		clusterRoleBinding := obj.(*rbacv1.ClusterRoleBinding)
		if liveClusterRoleBinding.RoleRef.APIGroup != clusterRoleBinding.RoleRef.APIGroup ||
			liveClusterRoleBinding.RoleRef.Kind != clusterRoleBinding.RoleRef.Kind ||
			liveClusterRoleBinding.RoleRef.Name != clusterRoleBinding.RoleRef.Name {
			return ErrImmutableRessource("roleRef", obj)
		}
	}

	// Set metadata.resourceVersion for updates as required by
	// Kubernetes API (http://go/k8s-concurrency).
	if gotRV := live.(metav1.Object).GetResourceVersion(); gotRV != "" {
		obj.(metav1.Object).SetResourceVersion(gotRV)
	}

	return nil
}

// maybeRecreate can be called to check if a resource can be updated or
// is immutable and needs recreation.
// It evaluates if resource should be forcefully recreated. In that case
// the resource will be deleted and recreated. If the -force flag is not
// enabled and an immutable resource should be updated, an error is thrown
// and no resources will get deleted.
func maybeRecreate(ctx context.Context, live, obj runtime.Object, m *kubePackage, r *apiResource) error {
	err := mergeObjects(live, obj)
	if errors.Is(errors.Unwrap(err), ErrUpdateImmutable) && m.force {
		if m.dryRun {
			fmt.Fprintf(os.Stdout, "\n\n**WARNING** %s %s is immutable and will be deleted and recreated.\n", strings.ToLower(r.GVK.Kind), maybeNamespaced(r.Name, r.Namespace))
		}
		// kubeDelete() already properly handles a dry run, so the resource won't be deleted if -force is set, but in dry run mode
		if err := m.kubeDelete(ctx, r, true); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// kubeUpdate creates or overwrites object in Kubernetes.
// Path is computed based on msg type, name and (optional) namespace (these must
// not conflict with name and namespace set in object metadata).
func (m *kubePackage) kubeUpdate(ctx context.Context, r *apiResource, obj runtime.Object) error {
	uri := r.PathWithName()
	live, found, err := m.kubePeek(ctx, m.Master+uri)
	if err != nil {
		return err
	}

	method := http.MethodPut
	if found {
		// Reset uri in case subresource update is requested.
		uri = r.PathWithSubresource()
		if err := maybeRecreate(ctx, live, obj, m, r); err != nil {
			return err
		}
	} else { // Object doesn't exist so create it.
		if r.Subresource != "" {
			return errors.New("parent resource does not exist")
		}

		method = http.MethodPost
		uri = r.Path()
	}

	//wip
	bs, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	url := m.Master + uri
	req, err := http.NewRequest(method, url, bytes.NewReader(bs))
	if err != nil {
		return err
	}

	log.V(1).Infof("%s to %s", method, url)

	if log.V(2) {
		s, err := renderObj(obj, &r.GVK, bool(log.V(3)) /* If --v=3, only return JSON. */, m.diffFilters)
		if err != nil {
			return fmt.Errorf("failed to render :live object for %s: %v", r.String(), err)
		}

		log.Infof("%s:\n%s", r.String(), s)
	}

	if m.diff {
		if err := printUnifiedDiff(os.Stdout, live, obj, r.GVK, maybeNamespaced(r.Name, r.Namespace), m.diffFilters); err != nil {
			return err
		}
	}

	if m.dryRun {
		return printUnifiedDiff(os.Stdout, live, obj, r.GVK, maybeNamespaced(r.Name, r.Namespace), m.diffFilters)
	}

	resp, err := m.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	_, rMsg, err := parseHTTPResponse(resp)
	if err != nil {
		return err
	}

	actionMsg := "created"
	if method == http.MethodPut {
		actionMsg = "updated"
	}
	log.Infof("%s %s", rMsg, actionMsg)

	return nil
}

// kubeDelete deletes namespace/name resource in Kubernetes.
// Attempts to deduce GroupVersionResource from apiGroup (optional) and resource
// strings. Fails if multiple matches found.
func (m *kubePackage) kubeDelete(_ context.Context, r *apiResource, foreground bool) error {
	var c dynamic.ResourceInterface = m.dynClient.Resource(r.GroupVersionResource())
	if r.Namespace != "" {
		c = c.(dynamic.NamespaceableResourceInterface).Namespace(r.Namespace)
	}

	delPolicy := metav1.DeletePropagationBackground
	if foreground {
		delPolicy = metav1.DeletePropagationForeground
	}

	log.V(1).Infof("DELETE to %s", m.Master+r.PathWithName())

	if m.dryRun {
		return nil
	}

	if err := c.Delete(context.TODO(), r.Name, metav1.DeleteOptions{
		PropagationPolicy: &delPolicy,
	}); err != nil {
		return err
	}

	log.Infof("%v deleted", r)

	return nil
}

// waitRetryInterval is a duration between consecutive get retries.
const waitRetryInterval = 500 * time.Millisecond

var ErrNotFound = errors.New("not found")

// kubeGet attempts to read namespace/name resource from an apiGroup from API
// Server.
// If object is not present will retry every waitRetryInterval up to wait (only
// tries once if wait is zero).
func (m *kubePackage) kubeGet(ctx context.Context, r *apiResource, wait time.Duration) (runtime.Object, error) {
	url := m.Master + r.PathWithName()
	var waitDone <-chan time.Time
	if wait != 0 {
		waitDone = time.After(wait)
	}

	// retryInterval is zero so no delay before the first poll.
	var retryInterval time.Duration
	for {
		select {
		case <-time.After(retryInterval):
			retryInterval = waitRetryInterval
			obj, ok, err := m.kubePeek(ctx, url)
			if err != nil {
				return nil, err
			}
			if ok {
				return obj, nil
			}
			if waitDone == nil {
				return nil, ErrNotFound
			}

		case <-waitDone:
			return nil, ErrNotFound

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// not reachable
}

func (m *kubePackage) kubePortForwardTestFn(t *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var namespace, pod, testPath string
	var localPort, podPort int
	unpacked := []interface{}{
		"namespace", &namespace,
		"pod", &pod,
		"local_port", &localPort,
		"pod_port", &podPort,
		"test_path", &testPath,
	}
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, unpacked...); err != nil {
		return starlark.False, fmt.Errorf("<%v>: %v", b.Name(), err)
	}

	// check if test path is already reachable before port forward
	testUrl := fmt.Sprintf("http://localhost:%v%v", localPort, testPath)

	req, err := http.NewRequest(http.MethodGet, testUrl, nil)
	if err != nil {
		return starlark.False, err
	}

	ctx := t.Local(addon.GoCtxKey).(context.Context)
	_, err = m.httpClient.Do(req.WithContext(ctx))
	if err == nil {
		return starlark.False, fmt.Errorf("endpoint %v already in use", testUrl)
	}

	// port forward the pod
	roundTripper, upgrader, err := spdy.RoundTripperFor(m.config)
	if err != nil {
		return starlark.False, err
	}

	portForwardPath := fmt.Sprintf("/api/v1/namespaces/%v/pods/%v/portforward", namespace, pod)

	host := strings.Replace(m.Master, "https://", "", 1)
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &url.URL{Scheme: "https", Path: portForwardPath, Host: host})

	var errorBuffer, outBuffer bytes.Buffer
	errorWriter := bufio.NewWriter(&errorBuffer)
	outWriter := bufio.NewWriter(&outBuffer)

	var readyChannel chan struct{}
	portForwardStopChannel := make(chan struct{})

	portForward, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", localPort, podPort)}, portForwardStopChannel, readyChannel, outWriter, errorWriter)
	if err != nil {
		return starlark.False, err
	}

	go func() {
		err = portForward.ForwardPorts()
		if err != nil {
			log.Fatal(err)
		}
	}()
	defer close(portForwardStopChannel)

	// check if endpoint is accessible after short delay
	time.Sleep(time.Millisecond * 125)

	var returnErr error
	for retries := 0; retries < 5; retries++ {
		resp, err := m.httpClient.Do(req.WithContext(ctx))
		if err != nil {
			time.Sleep(time.Duration(retries+1) * time.Second)
			continue
		}

		if resp.StatusCode < 400 || resp.StatusCode >= 500 {
			return starlark.True, nil
		}

		returnErr = fmt.Errorf("unable to reach service at url %v received status code %v (%v)", testUrl, resp.StatusCode, resp.Status)
	}

	return starlark.False, returnErr
}
