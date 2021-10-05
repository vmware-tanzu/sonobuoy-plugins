package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-version"
	appsv1 "k8s.io/api/apps/v1"
)

func checkDeployment(c Check) (CheckResult, error) {
	check := c.Deployment
	wantVer, err := version.NewVersion(check.Version)
	if err != nil {
		return CheckResult{Fail: true, Msgs: []string{fmt.Sprintf("failed to parse desired verstion %v as a semver value: %w", check.Version, err)}}, err
	}

	// Will probably change this logic; getting the deployments in go instead of kubectl is annoying
	// but getting just the annotation is very narrow/brittle and wont evolve well.
	dep, err := getDeployment(check.Name)
	if err != nil {
		return CheckResult{Fail: true, Msgs: []string{err.Error()}}, err
	}
	if dep==nil{
		return CheckResult{Fail: true,Msgs: []string{fmt.Sprintf("failed to find any deployments with name %v", check.Name)}},nil
	}

	val := dep.Annotations[check.Annotation]
	// Right now just hard coded for semver check but would need to check eq/DNE/semver/regexp etc
	haveVer, err := version.NewVersion(val)
	if err != nil {
		err = fmt.Errorf("annotation %q has value %q which failed to parse using semver: %w", check.Annotation, val, err)
		return CheckResult{Fail: true, Msgs: []string{err.Error()}}, err
	}

	if haveVer.LessThan(wantVer) {
		return CheckResult{
			Fail: true,
			Msgs: []string{fmt.Sprintf("wanted version >= %v but got %v", wantVer.String(), haveVer.String())},
		}, nil
	}

	return CheckResult{}, nil
}

func getDeployment(name string) (*appsv1.Deployment, error) {
	// Do some sanitation on the value? Need to make sure we don't allow injection attacks.
	b, err := runCmd(fmt.Sprintf("kubectl get deployments -A -o json|jq '.items[]|select(.metadata.name==\"%v\")'", name))
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, nil
	}
	var dep *appsv1.Deployment
	err = json.Unmarshal(b, &dep)
	return dep, err
}
