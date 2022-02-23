package main

// Result stores a suite of tests and their results in a form which serializes
// to sonobuoy's expected manual results format.
// https://sonobuoy.io/docs/main/results/#manual-results-format
type Result struct {
	Name   string
	Status string
	Meta   map[string]string
	Tests  []Result `yaml:"items,omitempty"`
}

// Test stores a suite of test definitions. When the input yaml file is read,
// it is serialized to this format, containing the name of the suite, the names
// of the individual tests, and the commands they run.
type Test struct {
	Name  string
	Tests []struct {
		Name string
		Cmd  string `yaml:",flow"`
	}
}

const (
	resDir  = "/tmp/sonobuoy/results/"
	resFile = "sonoshell.yaml"
)
