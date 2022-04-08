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

package sonobuoy

import (
	"context"
	"errors"
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	"github.com/k14s/starlark-go/starlarkstruct"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/carvel-ytt/pkg/template/core"
	sono "github.com/vmware-tanzu/sonobuoy-plugins/plugin-helper"
	"github.com/vmware-tanzu/sonobuoy-plugins/sonolark/lib/shared"
)

const (
	EnvKeySonobuoy          = "SONOBUOY"
	EnvKeySonobuoyConfigDir = "SONOBUOY_CONFIG_DIR"

	WriterCtxKey         = "sonoWriter"
	ProgressWriterCtxKey = "sonoProgressWriter"
	CurrentTestCtxKey    = "sonoCurrentTest"

	testStatusPassed  = "passed"
	testStatusFailed  = "failed"
	testStatusSkipped = "skipped"
	testStatusError   = "error"
)

var (
	API = starlark.StringDict{
		"sonobuoy": &starlarkstruct.Module{
			Name: "sonobuoy",
			Members: starlark.StringDict{
				"startSuite": starlark.NewBuiltin("sonobuoy.startSuite", core.ErrWrapper(sonobuoyModule{}.StartSuite)),
				"startTest":  starlark.NewBuiltin("sonobuoy.startTest", core.ErrWrapper(sonobuoyModule{}.StartTest)),
				"passTest":   starlark.NewBuiltin("sonobuoy.passTest", core.ErrWrapper(sonobuoyModule{}.PassTest)),
				"failTest":   starlark.NewBuiltin("sonobuoy.failTest", core.ErrWrapper(sonobuoyModule{}.FailTest)),
				"done":       starlark.NewBuiltin("sonobuoy.done", core.ErrWrapper(sonobuoyModule{}.Done)),
			},
		},
	}
)

type sonobuoyModule struct{}

// StartSuite initializes sonobuoy helpers and places them in the Go context for future invocations.
func (b sonobuoyModule) StartSuite(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() > 1 {
		return starlark.None, fmt.Errorf("expected at most one argument: test count")
	}

	count := int64(-1)
	if args.Len() > 0 {
		var err error
		count, err = core.NewStarlarkValue(args.Index(0)).AsInt64()
		if err != nil {
			return starlark.None, err
		}
	}
	StartSuite(thread, count)
	return starlark.None, nil
}

func StartSuite(thread *starlark.Thread, count int64) {
	w, pw := sono.NewDefaultSonobuoyResultsWriter(), sono.NewProgressReporter(count)
	shared.SetGoCtxWithValues(thread,
		WriterCtxKey, &w,
		ProgressWriterCtxKey, &pw,
	)
}

func (b sonobuoyModule) StartTest(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected atleast one argument: suite name, [test count]")
	}

	testName, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	_, _, pw := getSonobuoyHelpers(thread)
	pw.StartTest(testName)
	shared.SetGoCtxWithValues(thread, CurrentTestCtxKey, testName)

	return starlark.None, nil
}

func markTestComplete(thread *starlark.Thread, failed, skipped bool, err error, msg string) {
	ctx, w, pw := getSonobuoyHelpers(thread)
	maybeTestName := ctx.Value(CurrentTestCtxKey)
	testName := ""
	if maybeTestName != nil {
		testName = maybeTestName.(string)
	}
	if len(testName) == 0 {
		logrus.Warnf("Attempting to mark current test as complete (failed=%v skipped=%v err=%v msg=%v) but there is no currently executing test.", failed, skipped, err, msg)
		return
	}
	pw.StopTest(testName, failed, skipped, err)
	result := testStatusPassed
	switch {
	case failed:
		result = testStatusFailed
	case skipped:
		result = testStatusSkipped
	case err != nil:
		result = testStatusError
	}

	w.AddTest(testName, result, err, msg)

	// Clear out current test.
	shared.SetGoCtxWithValues(thread, CurrentTestCtxKey, "")
}

func (b sonobuoyModule) SkipTest(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected 1 argument: progress message for Sonobuoy")
	}

	msg, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	markTestComplete(thread, false, true, nil, msg)
	return starlark.None, nil
}

func (b sonobuoyModule) FailTest(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected 1 argument: progress message for Sonobuoy")
	}

	msg, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	FailTest(thread, msg)
	return starlark.None, nil
}

func FailTest(thread *starlark.Thread, msg string) {
	markTestComplete(thread, true, false, nil, msg)
}

func (b sonobuoyModule) PassTest(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	msg := ""
	var err error
	switch {
	case args.Len() == 1:
		msg, err = core.NewStarlarkValue(args.Index(0)).AsString()
		if err != nil {
			return starlark.None, err
		}
	case args.Len() > 1:
		if args.Len() > 1 {
			return starlark.None, fmt.Errorf("expected at most 1 argument: progress message for Sonobuoy")
		}
	}

	markTestComplete(thread, false, false, nil, msg)
	return starlark.None, nil
}

func (b sonobuoyModule) update(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return starlark.None, fmt.Errorf("expected 1 argument: progress message for Sonobuoy")
	}

	msg, err := core.NewStarlarkValue(args.Index(0)).AsString()
	if err != nil {
		return starlark.None, err
	}

	_, _, pw := getSonobuoyHelpers(thread)
	pw.SendMessage(msg)

	return starlark.None, nil
}

func (b sonobuoyModule) Done(thread *starlark.Thread, f *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 0 {
		return starlark.None, fmt.Errorf("expected no arguments")
	}
	Done(thread)
	return starlark.None, nil
}

func Done(thread *starlark.Thread) {
	logrus.Trace("sonobuoy.Done called")
	ctx, w, pw := getSonobuoyHelpers(thread)
	maybeTest := ctx.Value(CurrentTestCtxKey)
	testName := ""
	if maybeTest != nil {
		testName = maybeTest.(string)
	}
	if len(testName) > 0 {
		logrus.Tracef("Found test %q still marked as currently running. Marking it as failed.", testName)
		markTestComplete(thread, true, false, errors.New("suite completed while test still running"), "suite completed while test still running")
	}
	pw.SendMessage("Suite completed.")
	w.Done(true)
}

func getSonobuoyHelpers(thread *starlark.Thread) (context.Context, *sono.SonobuoyResultsWriter, *sono.ProgressReporter) {
	ctx := shared.GetGoCtx(thread)
	w := ctx.Value(WriterCtxKey).(*sono.SonobuoyResultsWriter)
	pw := ctx.Value(ProgressWriterCtxKey).(*sono.ProgressReporter)

	return ctx, w, pw
}

func RunningViaSonobuoy(env map[string]string) bool {
	logrus.Tracef("Checking if env.SONOBUOY==true indicating running via Sonobuoy. Value is: %q", env[EnvKeySonobuoy])
	return env[EnvKeySonobuoy] == "true"
}
