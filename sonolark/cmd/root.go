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
	"context"
	"errors"
	"net/http"
	"path/filepath"

	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template/core"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yttlibrary/overlay"
	"github.com/vmware-tanzu/sonobuoy-plugins/sonolark/lib/assert"
	"github.com/vmware-tanzu/sonobuoy-plugins/sonolark/lib/env"
	"github.com/vmware-tanzu/sonobuoy-plugins/sonolark/lib/log"
	"github.com/vmware-tanzu/sonobuoy-plugins/sonolark/lib/shared"
	"github.com/vmware-tanzu/sonobuoy-plugins/sonolark/lib/sonobuoy"
	"k8s.io/client-go/util/homedir"

	"github.com/k14s/starlark-go/starlark"
	"github.com/vmware-tanzu/carvel-ytt/pkg/yttlibrary"
	"github.com/vmware-tanzu/sonobuoy-plugins/sonolark/lib/kube"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

const (
	defaultScriptName = "script.star"
)

var (
	ignoreDiffFields = []string{"metadata.managedFields"}
)

type runInput struct {
	KubeConfigPath string
	Filename       string
	LogLevel       log.LevelFlagType
}

// rootCmd represents the base command when called without any subcommands
func getRootCmd(env map[string]string) *cobra.Command {
	in := runInput{}
	root := &cobra.Command{
		Use:   "sonolark",
		Short: "Sonolark is a tool which allows users to easily build scripts using the Starlark language on top of our library of useful functions including assertions, Kubernetes API access, and more.",
		RunE: func(cmd *cobra.Command, args []string) error {

			thread := &starlark.Thread{}
			shared.SetGoCtx(thread, context.Background())

			// Automatically start/end suite.
			sonobuoy.StartSuite(thread, -1)
			defer sonobuoy.Done(thread)

			predeclared, err := getLibraryFuncs(in.KubeConfigPath, env)
			if err != nil {
				return nil
			}

			_, err = starlark.ExecFile(thread, in.Filename, nil, *predeclared)
			if err != nil {
				if evalErr, ok := err.(*starlark.EvalError); ok {
					sonobuoy.FailTest(thread, evalErr.Backtrace())
					return errors.New(evalErr.Backtrace())
				}
				sonobuoy.FailTest(thread, err.Error())
				return err
			}
			return nil
		},
	}

	root.Flags().StringVarP(&in.Filename, "file", "f", getDefaultScriptName(env), "The name of the script to run")
	root.Flags().Var(&in.LogLevel, "level", "The Log level. One of {panic, fatal, error, warn, info, debug, trace}")
	if home := homedir.HomeDir(); home != "" {
		root.Flags().StringVar(&in.KubeConfigPath, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		root.Flags().StringVar(&in.KubeConfigPath, "kubeconfig", "", "absolute path to the kubeconfig file")
	}

	root.AddCommand(NewCmdVersion())
	return root
}

func getLibraryFuncs(kubeconfigPath string, currentEnv map[string]string) (*starlark.StringDict, error) {
	predeclared := starlark.StringDict{
		"sonobuoy": sonobuoy.API["sonobuoy"],
		"env":      env.NewAPI()["env"],

		// ytt
		"assert":  yttlibrary.AssertAPI["assert"],
		"regexp":  yttlibrary.RegexpAPI["regexp"],
		"md5":     yttlibrary.MD5API["md5"],
		"sha256":  yttlibrary.SHA256API["sha256"],
		"base64":  yttlibrary.Base64API["base64"],
		"json":    yttlibrary.JSONAPI["json"],
		"toml":    yttlibrary.TOMLAPI["toml"],
		"yaml":    yttlibrary.YAMLAPI["yaml"],
		"url":     yttlibrary.URLAPI["url"],
		"ip":      yttlibrary.IPAPI["ip"],
		"struct":  yttlibrary.StructAPI["struct"],
		"module":  yttlibrary.ModuleAPI["module"],
		"overlay": overlay.API["overlay"],
		"version": yttlibrary.VersionAPI["version"],
	}

	// Custom overrides so that we can make custom error messages.
	yttlibrary.AssertAPI["assert"].(*starlarkstruct.Module).Members["equals"] = starlark.NewBuiltin("assert.equals", core.ErrWrapper(assert.Equals))
	yttlibrary.AssertAPI["assert"].(*starlarkstruct.Module).Members["fail"] = starlark.NewBuiltin("assert.fail", core.ErrWrapper(assert.Fail))

	// Kubernetes API access via kube.*
	c := getClusterConfig(kubeconfigPath, currentEnv)
	dC := discovery.NewDiscoveryClientForConfigOrDie(c)
	t, err := rest.TransportFor(c)
	if err != nil {
		return nil, err
	}
	dynC, err := dynamic.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	predeclared["kube"] = kube.New(c.Host, dC, dynC, &http.Client{Transport: t}, true, false, false, ignoreDiffFields)["kube"]

	return &predeclared, nil
}
