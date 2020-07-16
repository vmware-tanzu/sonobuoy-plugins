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

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/vmware-tanzu/sonobuoy-plugins/cluster-inventory/pkg/inventory"
	"github.com/vmware-tanzu/sonobuoy-plugins/cluster-inventory/pkg/reports"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewRunCommand() *cobra.Command {
	var sonobuoyReport string
	var jsonReport string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the cluster inventory and produce reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := GetKubeClient()
			if err != nil {
				return errors.Wrap(err, "creating Kubernetes Client")
			}

			collector := inventory.NewCollector(client)
			results, err := collector.Run()
			if err != nil {
				return fmt.Errorf("error running inventory: %q", err)
			}

			if sonobuoyReport != "" {
				f, err := os.Create(sonobuoyReport)
				if err != nil {
					return errors.Wrap(err, "opening sonobuoy report file")
				}
				if err = reports.WriteSonobuoyReport(f, results); err != nil {
					return errors.Wrap(err, "writing sonobuoy report")
				}
			}

			if jsonReport != "" {
				f, err := os.Create(jsonReport)
				if err != nil {
					return errors.Wrap(err, "opening json report file")
				}

				b, err := json.Marshal(results)
				if _, err := fmt.Fprintln(f, string(b)); err != nil {
					return errors.Wrap(err, "writing json report")
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&sonobuoyReport, "sonobuoy-report", "", "Generate a Sonobuoy results report at the given path")
	cmd.Flags().StringVar(&jsonReport, "json-report", "", "Generate a JSON report at the given path")

	return cmd
}
