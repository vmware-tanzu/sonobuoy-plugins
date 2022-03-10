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

package shared

import (
	"context"

	"github.com/k14s/starlark-go/starlark"
)

func GetGoCtx(thread *starlark.Thread) context.Context {
	return thread.Local(GoCtxKey).(context.Context)
}

func SetGoCtx(thread *starlark.Thread, ctx context.Context) {
	thread.SetLocal(GoCtxKey, ctx)
}

func SetGoCtxWithValues(thread *starlark.Thread, keyValPairs ...interface{}) {
	len := len(keyValPairs)
	ctx := GetGoCtx(thread)
	for i := 0; i < len; i += 2 {
		key := keyValPairs[i]
		val := keyValPairs[i+1]
		ctx = context.WithValue(ctx, key, val)
	}
	SetGoCtx(thread, ctx)
}
