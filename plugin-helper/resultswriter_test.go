package plugin_helper

import (
	sono "github.com/vmware-tanzu/sonobuoy/pkg/client/results"
)

func ExampleWriterDoneToStdout() {
	// Note: The spacing looks weird in the output here just because of tabbing and yaml encoding.
	// The real point of the test is to ensure that a writer without a resultsDir will write to stdout.
	w := &SonobuoyResultsWriter{}
	w.Data = sono.Item{Name: "suite", Status: "shouldAggregateToPassed", Items: []sono.Item{{Name: "t1", Status: "passed"}}}
	w.Done(false)
	//Output: name: suite
	//status: passed
	//items:
	//- name: t1
	//   status: passed
}
