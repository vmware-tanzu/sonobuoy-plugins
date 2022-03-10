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
	"path/filepath"

	"github.com/vmware-tanzu/sonobuoy-plugins/sonolark/lib/sonobuoy"
)

var (
	envKeys = []string{sonobuoy.EnvKeySonobuoy, sonobuoy.EnvKeySonobuoyConfigDir}
)

// getEnvs grabs a series of keys of interest and just stores them in a map to pass around to
// decouple functions from the runtime env.
func getEnvs() map[string]string {
	envs := map[string]string{}
	for _, key := range envKeys {
		envs[key] = os.Getenv(key)
	}
	return envs
}

func getScriptName(env map[string]string) string {
	return filepath.Join(env[sonobuoy.EnvKeySonobuoyConfigDir], defaultScriptName)
}
