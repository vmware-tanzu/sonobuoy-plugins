/*
Copyright 2022 the Sonobuoy Project contributors.

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

package assert

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template/core"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
)

func NoOp(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.None, nil
}

// Fail is a slightly modified copy from ytt so that we can provide a custom failure message.
func Fail(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected exactly one argument")
	}

	val, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	return starlark.None, fmt.Errorf("fail: %s", val)
}

// Equals is a slightly modified copy of yttlibrary.assertLibrary.Equals so we can provide our own failure message.
func Equals(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 3 {
		return starlark.None, fmt.Errorf("expected three arguments")
	}

	expected := args.Index(0)
	if _, notOk := expected.(starlark.Callable); notOk {
		return starlark.None, fmt.Errorf("expected argument not to be a function, but was %T", expected)
	}

	actual := args.Index(1)
	if _, notOk := actual.(starlark.Callable); notOk {
		return starlark.None, fmt.Errorf("expected argument not to be a function, but was %T", actual)
	}

	expectedString, err := assertAsString(expected)
	if err != nil {
		return starlark.None, err
	}

	actualString, err := assertAsString(actual)
	if err != nil {
		return starlark.None, err
	}

	if expectedString != actualString {
		return starlark.None, fmt.Errorf(args.Index(2).String(), args.Index(0), args.Index(1))
	}

	return starlark.None, nil
}

// assertAsString is a copy from ytt unexported code to support the custom assert method.
func assertAsString(value starlark.Value) (string, error) {
	starlarkValue, err := core.NewStarlarkValue(value).AsGoValue()
	if err != nil {
		return "", err
	}
	yamlString, err := assertYamlEncode(starlarkValue)
	if err != nil {
		return "", err
	}
	return yamlString, nil
}

// assertYamlEncode is a copy from ytt unexported code to support the custom assert method.
func assertYamlEncode(goValue interface{}) (string, error) {
	var docSet *yamlmeta.DocumentSet

	switch typedVal := goValue.(type) {
	case *yamlmeta.DocumentSet:
		docSet = typedVal
	case *yamlmeta.Document:
		// Documents should be part of DocumentSet by the time it makes it here
		panic("Unexpected document")
	default:
		docSet = &yamlmeta.DocumentSet{Items: []*yamlmeta.Document{{Value: typedVal}}}
	}

	valBs, err := docSet.AsBytes()
	if err != nil {
		return "", err
	}

	return string(valBs), nil
}
