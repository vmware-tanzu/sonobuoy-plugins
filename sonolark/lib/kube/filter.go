// Copyright 2020 Cruise LLC
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
	yaml "gopkg.in/yaml.v2"
)

// filterYaml will deep copy m and remove the element at the yamlPath.
func filterYaml(m yaml.MapSlice, yamlPath ...string) yaml.MapSlice {
	var out yaml.MapSlice
	for _, item := range m {
		if f, ok := item.Key.(string); ok && f == yamlPath[0] {
			// path match found, skip element
			if len(yamlPath) == 1 {
				continue
			}

			// path match found, recurse into children
			if mm, ok := item.Value.(yaml.MapSlice); ok && len(yamlPath) > 1 {
				item = yaml.MapItem{
					Key:   item.Key,
					Value: filterYaml(mm, yamlPath[1:]...),
				}
			}
		}

		out = append(out, item)
	}
	return out
}

func filterEmpty(m yaml.MapSlice) yaml.MapSlice {
	var out yaml.MapSlice
	for _, item := range m {
		if value, ok := item.Value.(yaml.MapSlice); ok {
			// empty value, skip item
			if len(value) == 0 {
				continue
			}

			value = filterEmpty(value)

			// empty value after filtering, skip item
			if len(value) == 0 {
				continue
			}

			item = yaml.MapItem{
				Key:   item.Key,
				Value: value,
			}
		}

		out = append(out, item)
	}
	return out
}
