package main

import (
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)



func checkNodes(check Check) (CheckResult, error) {
	result := CheckResult{}

	nodes, err := getNodes(check.Node)
	if err != nil {
		return CheckResult{Fail: true, Msgs: []string{err.Error()}}, err
	}

	var wantCPU,wantMemory *resource.Quantity

	if len(check.Node.CPU)>0 {
		want, err := resource.ParseQuantity(check.Node.CPU)
		if err != nil {
			result.Fail=true
			result.Msgs = append(result.Msgs, fmt.Sprintf("failed to parse desired CPU value %q: %w", check.Node.CPU, err))
		}
		wantCPU=&want
	}

	if len(check.Node.CPU)>0 {
		want, err := resource.ParseQuantity(check.Node.Memory)
		if err != nil {
			result.Fail=true
			result.Msgs = append(result.Msgs, fmt.Sprintf("failed to parse desired memory value %q: %w.", check.Node.Memory, err))
		}
		wantMemory=&want
	}

	passedAll :=0
	for _,n := range nodes.Items{
		nodeFailed:=false
		if wantMemory!=nil{
			if n.Status.Capacity.Memory().Cmp(*wantMemory)<0{
				result.Msgs=append(result.Msgs,fmt.Sprintf("node %q failed to meet the desired memory: wanted %v but have %v.",n.ObjectMeta.Name, wantMemory.String(), n.Status.Capacity.Memory().String()))
				nodeFailed=true
			}
		}
		if wantCPU!=nil{
			if n.Status.Capacity.Cpu().Cmp(*wantCPU)<0{
				result.Msgs=append(result.Msgs,fmt.Sprintf("node %q failed to meet the desired CPU: wanted %v but have %v.",n.ObjectMeta.Name, wantCPU.String(), n.Status.Capacity.Cpu().String()))
				nodeFailed=true
			}
		}
		if !nodeFailed{
			passedAll+=1
		}
	}

	// If count == 0 then we assume they want _all_ labeled nodes to meet criteria.
	switch {
	case check.Node.Count == 0 && len(nodes.Items) < check.Node.Count:
		result.Fail = true
		result.Msgs = append(result.Msgs, fmt.Sprintf("expected all nodes labeled %q to match the criteria but only %v of %v did.", check.Node.Label, passedAll, len(nodes.Items)))
	case check.Node.Count > 0 && passedAll < check.Node.Count:
		result.Fail = true
		result.Msgs = append(result.Msgs, fmt.Sprintf("expected %v node(s) labeled %q to match the criteria but only %v of %v did.", check.Node.Count, check.Node.Label, passedAll, len(nodes.Items)))
	}

	return result,nil
}

func getNodes(check CheckTypeNode) (*corev1.NodeList, error) {
	// Do some sanitation on the label value? Need to make sure we don't allow injection attacks.
	b, err := runCmd(fmt.Sprintf("kubectl get nodes -o json -l %v",check.Label))
	if err != nil {
		return nil, err
	}
	var nodeList *corev1.NodeList
	err = json.Unmarshal(b, &nodeList)
	return nodeList, err
}
