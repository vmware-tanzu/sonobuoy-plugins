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

package env

import (
	"os"
	"strings"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
)

const (
	// EnvPrefix is the prefix each env var key should have in order to be exposed to the
	// script. The value will be all lowercase letters.
	EnvPrefix = "SONOLARK_"
)

// NewAPI returns a starlark.StringDict with the "env" module with all the constants derived from
// env var values of the form: SONOLARK_<name>=<value>. When exposed to the script; it can be invoked
// by env.<name>
func NewAPI() starlark.StringDict {
	m := &starlarkstruct.Module{
		Name:    "env",
		Members: starlark.StringDict{},
	}
	for _, keyval := range os.Environ() {
		parts := strings.SplitN(keyval, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if !strings.HasPrefix(parts[0], EnvPrefix) {
			continue
		}
		key := strings.ToLower(strings.TrimPrefix(parts[0], EnvPrefix))
		m.Members[key] = starlark.String(parts[1])
	}
	return starlark.StringDict{"env": m}
}
