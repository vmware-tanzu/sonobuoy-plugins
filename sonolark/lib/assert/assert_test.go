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

package assert

import (
	"testing"

	"github.com/k14s/starlark-go/starlark"
)

func TestGetFmtStringFromArgs(t *testing.T) {
	v1, v2 := starlark.String("a"), starlark.String("b")
	testcases := []struct {
		desc   string
		input  string
		expect string
	}{
		{
			desc:  "All subs work",
			input: "$1 $2 $3 $4 $5",
			expect: `"a" "b" -"a"
+"b" string string`,
		}, {
			desc:  "Default looks OK",
			input: defaultErrorMsg,
			expect: `Not equal:

			(expected type: string)
"a"

(was type: string)
"b"`,
		}, {
			desc:   "Can repeat subs",
			input:  "$1 $2 $1",
			expect: `"a" "b" "a"`,
		}, {
			desc:  "Order doesnt matter",
			input: "$3 $2 $1",
			expect: `-"a"
+"b" "b" "a"`,
		}, {
			desc:   "OK if no subs",
			input:  "static",
			expect: `static`,
		}, {
			desc:   "Dollars signs and numbers are still fine",
			input:  "$ 1",
			expect: `$ 1`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			out := getFmtStringFromArgs(tc.input, v1, v2)
			if out != tc.expect {
				t.Errorf("Expected '%v'\n but got '%v'", tc.expect, out)
			}
		})
	}
}
