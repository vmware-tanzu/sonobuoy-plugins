package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"

	pluginhelper "github.com/vmware-tanzu/sonobuoy-plugins/plugin-helper"
	sono "github.com/vmware-tanzu/sonobuoy/pkg/client/results"
)

const (
	defaultInputFile = "input.json"
	pluginInputDir   = "/tmp/sonobuoy/config"
)


func main() {
	// Debug
	logrus.SetLevel(logrus.TraceLevel)
	inputFile := defaultInputFile
	if os.Getenv("SONOBUOY_K8S_VERSION") != "" {
		inputFile = filepath.Join(pluginInputDir, inputFile)
	}
	inputB, err := os.ReadFile(inputFile)
	if err != nil {
		panic(err)
	}

	var checks CheckList
	if err := json.Unmarshal(inputB, &checks); err != nil {
		panic(err)
	}

	w := pluginhelper.NewDefaultSonobuoyResultsWriter()
	p := pluginhelper.NewProgressReporter(int64(len(checks)))

	// Refactor this with an interface for those methods so we can just save a map between
	// type and method which will return `result, error` and avoid this whole switch here.
	for _, check := range checks {
		f, ok := typeFuncLookup[check.Meta.Type]
		if !ok {
			fmt.Fprintf(os.Stderr, "Unknown check type: %v", check.Meta.Type)
		}
		p.StartTest(check.Meta.Name)
		res, err := f(check)

		logrus.Tracef("Completed test %q, result: %v\n", check.Meta.Name, failToStatus(res.Fail))
		w.AddTest(check.Meta.Name, failToStatus(res.Fail), err, "")
		p.StopTest(check.Meta.Name, res.Fail, false, err)
	}

	if err := w.Done(true); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func failToStatus(failed bool) string {
	if failed {
		return sono.StatusFailed
	}
	return sono.StatusPassed
}

func runCmd(cmdText string) ([]byte, error) {
	c := exec.Command("/bin/bash", "-c", cmdText)
	logrus.Traceln(c.String(), c.Args)
	out, err := c.CombinedOutput()
	logrus.Traceln("Command:", c.String())
	logrus.Traceln("Output:", string(out))
	if err != nil {
		logrus.Traceln("Error returned:", err.Error())
	}
	return bytes.TrimSpace(out), err
}

func runFilterCmd(inputFile, filter string) ([]byte, error) {
	o, err := runCmd(fmt.Sprintf(`%v %v`, filter, string(inputFile)))
	return o, err
}
