/*
Copyright 2022 the Sonobuoy Project contributors

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

// Package buildinfo holds build-time information.
// This is a separate package so that other packages can import it without
// worrying about introducing circular dependencies.
package buildinfo

import (
	_ "embed"
)

//go:generate bash get_version.sh
//go:embed version.txt
var version string

// GitSHA is the actual commit that is being built.
//go:generate bash get_version.sh
//go:embed gitsha.txt
var GitSHA string

type InfoObj struct {
	Version string `json:"version"`
	GitSHA  string `json:"git_sha"`
}

var Info = InfoObj{
	Version: version,
	GitSHA:  GitSHA,
}
