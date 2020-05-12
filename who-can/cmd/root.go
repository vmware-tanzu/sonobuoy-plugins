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
	"fmt"
	"os"

	"github.com/vmware-tanzu/sonobuoy-plugins/who-can/pkg/whocan"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

const subjectsReportFlag = "subjects-report"
const resourcesReportFlag = "resources-report"
const sonobuoyReportFlag = "sonobuoy-report"

func NewWhoCanCommand() *cobra.Command {
	var subjectsReport string
	var resourcesReport string
	var sonobuoyReport string
	var restQPS float32
	var restBurst int

	cmds := &cobra.Command{
		Use:   "who-can",
		Short: "Creates reports of who can perform actions in your cluster",
		Long:  "who-can iterates over all resources in your Kubernetes cluster and produces reports detailing which subjects have RBAC permissions to perform actions against those resources",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if subjectsReport == "" && resourcesReport == "" && sonobuoyReport == "" {
				return fmt.Errorf("No output file specified. Must provide at least one of --%v, --%v, or --%v", subjectsReportFlag, resourcesReportFlag, sonobuoyReportFlag)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhoCanPlugin(subjectsReport, resourcesReport, sonobuoyReport, restQPS, restBurst)
		},
	}

	cmds.ResetFlags()
	cmds.Flags().StringVar(&subjectsReport, subjectsReportFlag, "", "Generate a JSON report of the results by subject at the given path")
	cmds.Flags().StringVar(&resourcesReport, resourcesReportFlag, "", "Generate a JSON report of the results by resource at the given path")
	cmds.Flags().StringVar(&sonobuoyReport, sonobuoyReportFlag, "", "Generate a Sonobuoy results report by subject at the given path")
	cmds.Flags().Float32Var(&restQPS, "qps", 100.0, "QPS for Kubernetes REST client")
	cmds.Flags().IntVar(&restBurst, "burst", 50, "Burst for Kubernetes REST client")
	return cmds
}

func runWhoCanPlugin(subjectsReport string, resourcesReport string, sonobuoyReport string, restQPS float32, restBurst int) error {
	restConfig, err := getRESTConfig(restQPS, restBurst)
	if err != nil {
		return errors.Wrap(err, "getting REST Config")
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return errors.Wrap(err, "creating Kubernetes client")
	}

	checker, err := whocan.NewChecker(client, restConfig)
	if err != nil {
		return errors.Wrap(err, "creating who-can checker")
	}

	runner := whocan.NewRunner(checker)

	resources, err := getAPIResources(client)
	if err != nil {
		return errors.Wrap(err, "getting resources")
	}

	whoCanConfig, err := whocan.LoadConfigFromEnv()
	if err != nil {
		return errors.Wrap(err, "getting who-can config")
	}

	namespaces := whoCanConfig.Namespaces
	if len(namespaces) == 0 {
		namespaces = []string{"default"}
	}

	results, err := runner.Run(namespaces, resources)
	if err != nil {
		// TODO write error sonobuoy report if there is an error running who-can
		return errors.Wrap(err, "running who can")
	}

	if subjectsReport != "" {
		f, err := os.Create(subjectsReport)
		if err != nil {
			return errors.Wrap(err, "opening subjects report file")
		}
		err = results.WriteSubjectsReport(f)
		if err != nil {
			return errors.Wrap(err, "writing subjects report file")
		}
	}

	if resourcesReport != "" {
		f, err := os.Create(resourcesReport)
		if err != nil {
			return errors.Wrap(err, "opening resources report file")
		}
		err = results.WriteResourcesReport(f)
		if err != nil {
			return errors.Wrap(err, "writing resources report file")
		}
	}

	if sonobuoyReport != "" {
		f, err := os.Create(sonobuoyReport)
		if err != nil {
			return errors.Wrap(err, "opening sonobuoy report file")
		}
		err = results.WriteSonobuoyReport(f)
		if err != nil {
			return errors.Wrap(err, "writing sonobuoy report file")
		}
	}

	return nil
}
