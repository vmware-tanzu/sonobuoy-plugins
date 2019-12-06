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
	"os"

	"github.com/vmware-tanzu/sonobuoy-plugins/who-can/pkg/whocan"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

func NewWhoCanCommand() *cobra.Command {
	var subjectsReport string
	var resourcesReport string
	var sonobuoyReport string
	var restQPS float32
	var restBurst int

	cmds := &cobra.Command{
		Use:   "who-can",
		Short: "Creates reports of who can perform actions in your cluster",
		Long:  "TODO enter long description here",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhoCanPlugin(subjectsReport, resourcesReport, sonobuoyReport, restQPS, restBurst)
		},
	}

	cmds.ResetFlags()
	cmds.Flags().StringVar(&subjectsReport, "subjects-report", "", "Generate a JSON report of the results by subject at the given path")
	cmds.Flags().StringVar(&resourcesReport, "resources-report", "", "Generate a JSON report of the results by resource at the given path")
	cmds.Flags().StringVar(&sonobuoyReport, "sonobuoy-report", "", "Generate a Sonobuoy results report by subject at the given path")
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
