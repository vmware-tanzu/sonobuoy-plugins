package plugin_helper

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vmware-tanzu/sonobuoy/pkg/tarball"
)

const(
	SonobuoyResultsDirKey="SONOBUOY_RESULTS_DIR"
	DoneFileName = "done"
	DefaultTarballName = "results.tar.gz"
)

// Done will tar up the results directory, write the done file which instructs Sonobuoy to
// submit results to the aggregator.
func Done() error{
	dir := GetResultsDir()
	outputFile := filepath.Join(dir, DefaultTarballName)
	if err := tarball.DirToTarball(dir, outputFile,true ); err!=nil{
		return fmt.Errorf("failed to tar up entire results directory: %w",err)
	}
	return WriteDone(outputFile)
}

func GetResultsDir() string{
	return os.Getenv(SonobuoyResultsDirKey)
}

func WriteDone(resultsPath string) error{
	if err:=os.WriteFile(filepath.Join(GetResultsDir(), DoneFileName), []byte(DefaultTarballName), 0666); err !=nil{
		return fmt.Errorf("failed write done file: %w",err)
	}
	return nil
}
