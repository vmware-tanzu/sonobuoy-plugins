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

package reports

import (
	"gopkg.in/yaml.v2"
	"io"
)

type SonobuoyResultsItem struct {
	Name     string                 `json:"name" yaml:"name"`
	Status   string                 `json:"status" yaml:"status,omitempty"`
	Metadata map[string]string      `json:"meta,omitempty" yaml:"meta,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty" yaml:"details,omitempty"`
	Items    []SonobuoyResultsItem  `json:"items,omitempty" yaml:"items,omitempty"`
}

type SonobuoyItemGenerator interface {
	GenerateSonobuoyItem() SonobuoyResultsItem
}

func WriteSonobuoyReport(w io.Writer, s SonobuoyItemGenerator) error {
	item := s.GenerateSonobuoyItem()
	j, err := yaml.Marshal(item)
	if err != nil {
		return err
	}
	_, err = w.Write(j)
	return err
}
