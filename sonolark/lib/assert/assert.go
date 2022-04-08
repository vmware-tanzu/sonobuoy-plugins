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
	"errors"
	"fmt"
	"strings"

	"github.com/k14s/starlark-go/starlark"
	"github.com/kylelemons/godebug/pretty"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template/core"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yamlmeta"
)

const (
	defaultErrorMsg = "Not equal:\n\n\t\t\t(expected type: $4)\n$1\n\n(was type: $5)\n$2"
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
	if args.Len() != 3 && args.Len() != 2 {
		return starlark.None, fmt.Errorf("expected 2 or 3 arguments")
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
		errMsg := defaultErrorMsg
		if len(args) == 3 {
			errMsg = args.Index(2).String()
		}
		return starlark.None, errors.New(getFmtStringFromArgs(errMsg, expected, actual))
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

// getFmtStringFromArgs takes an arbitrary fmt string and replaces the values
// $1, $2, $3, $4, and $5 with v1, v2, diff(v1, v2), v1.Type, and v2.Type respectively.
// The keywords (e.g. $1) can be provided any number of times in any order (or not at all).
func getFmtStringFromArgs(input string, v1, v2 starlark.Value) string {
	// Only calc diff if necessary.
	diff := ""
	if i3 := strings.Index(input, "$3"); i3 >= 0 {
		diff = pretty.Compare(v1, v2)
	}

	rep := strings.NewReplacer(
		"$1", fmt.Sprint(v1),
		"$2", fmt.Sprint(v2),
		"$3", fmt.Sprint(diff),
		"$4", v1.Type(),
		"$5", v2.Type(),
		`\n`, "\n",
		`\t`, "\t",
	)

	// Run it through twice to resolve any newlines/tabs that get placed into the diff.
	return rep.Replace(rep.Replace(input))
}
