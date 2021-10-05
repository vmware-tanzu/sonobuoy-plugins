package main

func checkProvider(check Check) (CheckResult, error) {
	v, err := getProvider()
	if err != nil {
		return CheckResult{Fail: true}, err
	}

	return CheckResult{Fail: !in(v, check.Provider.In) || in(v,check.Provider.NotIn)}, nil
}

func in(s string, list []string) bool {
	for _,v:=range list{
		if v==s {
			return true
		}
	}
	return false
}

// getProvider should return a string representing the cluster provider. Currently, it relies only the
// providerID field of the node.spec and I feel fairly condfident not every provider sets that as expected.
// TODO(jschnake): Go through intended providers and test them each one one.
func getProvider() (string, error) {
	b, err := runCmd("kubectl get nodes -o json|jq '.items[]|.spec.providerID' -r| cut -d : -f1|sort|uniq")
	return string(b), err
}
