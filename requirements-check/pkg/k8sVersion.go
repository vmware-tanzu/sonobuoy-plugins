package main

import (
	"fmt"

	version "github.com/hashicorp/go-version"
)

func checkK8sVersion(check Check) (CheckResult, error) {
	v, err := getK8sVersion()
	if err != nil {
		return CheckResult{Fail:true}, err
	}

	serverVer, err := version.NewVersion(v)
	if err != nil {
		return CheckResult{Fail:true}, fmt.Errorf("failed to parse server version %q: %w",v,err)
	}
	wantVersion, err := version.NewVersion(check.K8sVersion.Version)
	if err != nil {
		return CheckResult{Fail:true}, fmt.Errorf("failed to parse server version %q: %w",check.K8sVersion.Version,err)
	}

	if check.K8sVersion.Exact {
		return CheckResult{Fail:serverVer.Equal(wantVersion)}, nil
	}

	return CheckResult{Fail: serverVer.LessThan(wantVersion)}, nil
}

func getK8sVersion() (string, error) {
	b, err := runCmd("kubectl version -o json|jq .serverVersion.gitVersion -r")
	return string(b), err
}
