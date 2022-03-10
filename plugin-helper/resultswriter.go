package plugin_helper

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	sono "github.com/vmware-tanzu/sonobuoy/pkg/client/results"
	"gopkg.in/yaml.v2"
)

const (
	defaultOutputFileName = "sonobuoy_results.yaml"
)

// SonobuoyResultsWriter will keep track of result items in memory and will
// write them to the .ResultsDir/.OutputFile. If the ResultsDir is empty,
// the data will be dumped to stdout instead.
type SonobuoyResultsWriter struct {
	ResultsDir string
	OutputFile string
	Data       sono.Item
}

func NewDefaultSonobuoyResultsWriter() SonobuoyResultsWriter {
	return NewSonobuoyResultsWriter(os.Getenv("SONOBUOY_RESULTS_DIR"), defaultOutputFileName)
}

func NewSonobuoyResultsWriter(resultsDir, outputFile string) SonobuoyResultsWriter {
	w := SonobuoyResultsWriter{
		ResultsDir: resultsDir,
		OutputFile: outputFile,
		Data:       sono.Item{Items: []sono.Item{}},
	}
	return w
}

func (w *SonobuoyResultsWriter) AddTest(
	testName string,
	result string,
	err error,
	output string,
) {
	i := sono.Item{
		Name:   testName,
		Status: result,
	}
	if len(output) > 0 {
		if i.Details == nil {
			i.Details = map[string]interface{}{}
		}
		i.Details[sono.MetadataDetailsOutput] = output
	}
	if err != nil {
		if i.Details == nil {
			i.Details = map[string]interface{}{}
		}
		i.Details[sono.MetadataDetailsFailure] = err.Error()
	}
	w.Data.Items = append(w.Data.Items, i)
}

func (w *SonobuoyResultsWriter) Done(writeDoneFile bool) error {
	w.Data.Status = sono.AggregateStatus(w.Data.Items...)

	out := os.Stdout
	if len(w.ResultsDir) > 0 {
		# Ensure ResultsDir already exists
		if err := MkdirAll(w.ResultsDir, fs.ModeDir|fs.ModePerm); err != nil {
			return errors.Wrap(err, "error creating results directory")
		}
		outfile, err := os.Create(filepath.Join(w.ResultsDir, w.OutputFile))
		if err != nil {
			return errors.Wrap(err, "error creating results file")
		}
		defer outfile.Close()
		out = outfile
	}

	enc := yaml.NewEncoder(out)
	defer enc.Close()
	if err := enc.Encode(w.Data); err != nil {
		return errors.Wrap(err, "error writing to results file")
	}
	if writeDoneFile {
		return Done()
	}
	return nil
}
