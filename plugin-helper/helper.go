package plugin_helper

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/sonobuoy/pkg/tarball"
)

const (
	SonobuoyResultsDirKey = "SONOBUOY_RESULTS_DIR"
	DoneFileName          = "done"
	DefaultTarballName    = "results.tar.gz"
)

// Done will tar up the results directory, write the done file which instructs Sonobuoy to
// submit results to the aggregator.
func Done() error {
	dir := GetResultsDir()
	if len(dir) == 0 {
		logrus.Warnf("No %v set, no results directory will be archived and no 'done file' will be written.", SonobuoyResultsDirKey)
		return nil
	}

	outputFile := filepath.Join(dir, DefaultTarballName)
	logrus.Tracef("Tarring up directory: %v", dir)
	if err := tarball.DirToTarball(dir, outputFile, true); err != nil {
		return fmt.Errorf("failed to tar up entire results directory: %w", err)
	}
	logrus.Trace("Writing done file...")
	if err := WriteDone(outputFile); err != nil {
		return err
	}
	logrus.Trace("Done file written without error.")
	return nil
}

func GetResultsDir() string {
	return os.Getenv(SonobuoyResultsDirKey)
}

func WriteDone(resultsPath string) error {
	if err := os.WriteFile(filepath.Join(GetResultsDir(), DoneFileName), []byte(resultsPath), 0666); err != nil {
		return fmt.Errorf("failed write done file: %w", err)
	}
	return nil
}
