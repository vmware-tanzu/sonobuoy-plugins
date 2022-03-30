package main

import (
	"log"
	"os"
	"os/exec"
	"regexp"

	pluginhelper "github.com/vmware-tanzu/sonobuoy-plugins/plugin-helper"
	"gopkg.in/yaml.v3"
)

// Delegate the parsing of the command to bash by writing the test's cmd to a tmpfile and executing that
func (t *Test) MakeTestPairs() (pairs map[string]string) {
	pairs = make(map[string]string)

	// Ensure generated filenames don't contain special characters
	specials := regexp.MustCompile(`\W+`)

	for _, test := range t.Tests {
		fname := specials.ReplaceAllString(test.Name, "_")
		f, err := os.CreateTemp("", fname)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := f.Write([]byte(test.Cmd)); err != nil {
			log.Fatal(err)
		}
		pairs[test.Name] = f.Name()
		log.Printf("Pair added: ")
		log.Println(test.Name, "=", pairs[test.Name])
	}

	return
}

// Execute the tests stored in the given set of name: cmd pairs and store the results
func (results *Result) RunTests(pairs map[string]string, w *pluginhelper.SonobuoyResultsWriter) {
	p := pluginhelper.NewProgressReporter(int64(len(pairs)))
	for name, cmd := range pairs {
		cur := Result{Name: name, Meta: make(map[string]string)}
		p.StartTest(name)
		cmd := exec.Command("bash", cmd)
		output, err := cmd.CombinedOutput()
		//TODO: This currently doesn't distinguish between a test failing and
		//something else going wrong, so we currently don't meaningfully report if
		//some other error happened.
		if err != nil {
			cur.Status = "failed"
			p.StopTest(name, true, false, nil)
		} else {
			cur.Status = "passed"
			p.StopTest(name, false, false, nil)
		}
		w.AddTest(name, cur.Status, nil, string(output))
		log.Printf("Status of \"%s\": %s\n", name, cur.Status)

		results.Tests = append(results.Tests, cur)
	}
}

func main() {
	if len(os.Args[1:]) == 0 {
		log.Fatal("Must pass a path to an input yaml")
	}

	// Read in a yaml as input and convert it to a suite of tests
	yml, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	testspec := Test{}
	if err := yaml.Unmarshal(yml, &testspec); err != nil {
		log.Fatal(err)
	}

	testPairs := testspec.MakeTestPairs()

	results := Result{Name: testspec.Name, Meta: make(map[string]string)}
	results.Meta["type"] = "summary"

	w := pluginhelper.NewSonobuoyResultsWriter(resDir, resFile)
	results.RunTests(testPairs, &w)

	// Write the overall status; if any tests fail, report a failure
	results.Status = "passed"
	for _, t := range results.Tests {
		if t.Status != "passed" {
			results.Status = "failed"
		}
	}

	if err := w.Done(true); err != nil {
		log.Fatal(err)
	}
	if err := pluginhelper.Done(); err != nil {
		log.Fatal(err)
	}
}
