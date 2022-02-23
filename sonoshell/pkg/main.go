package main

import (
	"io/fs"
	"log"
	"os"
	"os/exec"

	pluginhelper "github.com/vmware-tanzu/sonobuoy-plugins/plugin-helper"
	"gopkg.in/yaml.v3"
)

// Delegate the parsing of the command to bash by writing the test's cmd to a tmpfile and executing that
func (t *Test) MakeTestPairs() (pairs map[string]string) {
	pairs = make(map[string]string)
	for _, test := range t.Tests {
		f, err := os.CreateTemp("", test.Name)
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
func (results *Result) RunTests(pairs map[string]string) {
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
		log.Printf("Status of \"%s\": %s\n", name, cur.Status)
		// Need to put the output in the directory so that we can send it off
		f, err := os.CreateTemp("", name)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := f.Write(output); err != nil {
			log.Fatal(err)
		}
		cur.Meta["type"] = "file"
		cur.Meta["file"] = f.Name()

		results.Tests = append(results.Tests, cur)
	}
}

// Marshal the test results to a yaml file for sonobuoy to consume
func (results Result) emit() {
	if err := os.MkdirAll(resDir, fs.ModeDir|fs.ModePerm); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(resDir + resFile)
	if err != nil {
		log.Fatal(err)
	}
	output, err := yaml.Marshal(results)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	if _, err := f.Write(output); err != nil {
		log.Fatal(err)
	}
	log.Printf("Wrote output to %s", f.Name())
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
	//fmt.Printf("---\n%v\n", testspec) // DEBUG

	test_pairs := testspec.MakeTestPairs()

	//fmt.Println(test_pairs) // DEBUG

	results := Result{Name: testspec.Name, Meta: make(map[string]string)}
	results.Meta["type"] = "summary"

	results.RunTests(test_pairs)

	// Write the overall status; if any tests fail, report a failure
	results.Status = "passed"
	for _, t := range results.Tests {
		if t.Status != "passed" {
			results.Status = "failed"
		}
	}

	results.emit()
	if err := pluginhelper.Done(); err != nil {
		log.Fatal(err)
	}
}
