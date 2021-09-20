package main

type Metadata struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        CheckType `json:"type"`
	Optional    bool      `json:"optional"`
}

type CheckType string

const (
	// Should match the json tag for clarity to user.
	K8sVersionCheck = CheckType("k8s_version")
	ProviderCheck   = CheckType("provider")
	NodeCheck       = CheckType("node")
	DeploymentCheck = CheckType("deployment")
)

type Checker func(Check) (CheckResult,error)

var (
	typeFuncLookup = map[CheckType]Checker{
		K8sVersionCheck: Checker(checkK8sVersion),
		ProviderCheck   : Checker(checkProvider),
		NodeCheck       : Checker(checkNodes),
		DeploymentCheck : Checker(checkDeployment),
}
)

type CheckList []Check

type Check struct {
	Meta       Metadata                   `json:"meta"`
	K8sVersion CheckTypeKubernetesVersion `json:"k8s_version"`
	Provider   CheckTypeProvider          `json:"provider"`
	Node       CheckTypeNode              `json:"node"`
	Deployment CheckTypeDeployment        `json:"deployment"`
}

type CheckResult struct {
	Fail bool
	Msgs []string
}

type CheckTypeKubernetesVersion struct {
	Version string `json:"version"`
	Exact   bool   `json:"exact"`
}

type CheckTypeProvider struct {
	In    []string `json:"in"`
	NotIn []string `json:"not_in"`
}

type CheckTypeNode struct {
	Label  string `json:"label"`
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
	Count  int    `json:"count"`
}

// TODO(jschnake): Search by name? Annotation? Label for pods it makes?
type CheckTypeDeployment struct {
	Name       string `json:"name,omitempty"`
	Annotation string `json:"annotation,omitempty"`
	Version    string `json:"version,omitempty"`
}
