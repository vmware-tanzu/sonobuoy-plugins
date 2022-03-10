package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"

	sono "github.com/vmware-tanzu/sonobuoy/pkg/client/results"
	pluginhelper "github.com/vmware-tanzu/sonobuoy-plugins/plugin-helper"
)

const (
	defaultInputFile = "input.json"
	pluginInputDir="/tmp/sonobuoy/config"
)


func main() {
	// Debug
	logrus.SetLevel(logrus.TraceLevel)
	inputFile:=defaultInputFile
	if os.Getenv("SONOBUOY_K8S_VERSION")!=""{
		inputFile=filepath.Join(pluginInputDir,inputFile)
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

	// Refactor this with an interface for those methods so we can just save a map between
	// type and method which will return `result, error` and avoid this whole switch here.
	for _, check := range checks {
		f,ok:=typeFuncLookup[check.Meta.Type]
		if !ok{
			fmt.Fprintf(os.Stderr, "Unknown check type: %v", check.Meta.Type)
		}
		res,err := f(check)

		logrus.Tracef("Completed test %q, result: %v\n", check.Meta.Name, failToStatus(res.Fail))
		w.AddTest(check.Meta.Name, failToStatus(res.Fail), err,"")
	}

	if err := w.Done(); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func failToStatus(failed bool) string{
	if failed{
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
