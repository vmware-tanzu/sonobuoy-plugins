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

	"github.com/k14s/starlark-go/starlark"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// resourceQuantityFn returns a starlark.Value that represents
func resourceQuantityFn(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	var v string
	if err := starlark.UnpackPositionalArgs(b.Name(), args, nil, 1, &v); err != nil {
		return nil, err
	}

	q, err := resource.ParseQuantity(v)
	if err != nil {
		return nil, fmt.Errorf("%v: failed to parse quantity string: %v", b.Name(), err)
	}

	un, err := runtime.DefaultUnstructuredConverter.ToUnstructured(q)
	if err != nil {
		return nil, fmt.Errorf("<%v>: failed to convert '%v' to unstructured JSON: %v", b.Name(), q, err)
	}

	return ValueFromNestedMap(un)
}

// fromStringFn converts Stalark integer to string *intstr.IntOrString
func fromStringFn(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	var v string
	if err := starlark.UnpackPositionalArgs(b.Name(), args, nil, 1, &v); err != nil {
		return nil, err
	}

	p := intstr.FromString(v)
	return starlark.String(p.String()), nil
}

// fromIntFn converts Stalark integer to integer *intstr.IntOrString
func fromIntFn(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	var v int
	if err := starlark.UnpackPositionalArgs(b.Name(), args, nil, 1, &v); err != nil {
		return nil, err
	}

	p := intstr.FromInt(v)
	return starlark.String(p.String()), nil
}
