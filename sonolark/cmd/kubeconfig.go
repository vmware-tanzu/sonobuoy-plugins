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

package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/sonobuoy-plugins/sonolark/lib/sonobuoy"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getInClusterConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	return config
}

func getClusterConfig(kubeconfigPath string, env map[string]string) *rest.Config {
	if sonobuoy.RunningViaSonobuoy(env) {
		return getInClusterConfig()
	}
	return getOutOfClusterConfig(kubeconfigPath)
}

func getOutOfClusterConfig(kubeconfigPath string) *rest.Config {
	// Use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err.Error())
	}

	// Adjust QPS/Burst so that the queries execute as quickly as possible.
	config.QPS = float32(rest.DefaultQPS * 10)
	config.Burst = rest.DefaultBurst * 10

	return config
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	logrus.SetLevel(logrus.TraceLevel)
	err := getRootCmd(getEnvs()).Execute()
	if err != nil {
		os.Exit(1)
	}
}
