/*
Copyright 2022 the Sonobuoy Project contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ph "github.com/vmware-tanzu/sonobuoy-plugins/plugin-helper"
	"gopkg.in/yaml.v2"

	"github.com/vmware-tanzu/sonobuoy/pkg/client/results"
	"github.com/vmware-tanzu/sonobuoy/pkg/plugin/driver/job"
	"github.com/vmware-tanzu/sonobuoy/pkg/plugin/manifest"
)

const (
	donefile = "done"
)

// rootCmd represents the base command when called without any subcommands
func getRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sonobuoy-post",
		Short: "Post-processor for Sonobuoy plugins",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return waitForDone()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := os.Getenv("SONOBUOY_RESULTS_DIR")

			// First we have to convert to the common yaml format.
			// WIP just assuming junit and hardcoding the conversion
			pName := "myplugin"
			format := "junit"

			m := manifest.Manifest{SonobuoyConfig: manifest.SonobuoyConfig{PluginName: pName, Driver: "job", ResultFormat: format}}
			p := job.NewPlugin(m, "", "", "", "", nil)
			items, err:=results.ProcessDir(p,"", dir, results.JunitProcessFile, results.FileOrExtension([]string{}, ".xml"))
			if err != nil {
				logrus.Errorf("Error processing plugin %v: %v", p.GetName(), err)
				return err
			}
			if len(items)==0{
				return errors.New("did not get any results when processing results")
			}

			// Save existing yaml so we can apply ytt transform to it.
			results := results.Item{
				Name:     p.GetName(),
				Metadata: map[string]string{results.MetadataTypeKey: results.MetadataTypeSummary},
			}

			results.Items = append(results.Items, items...)
			SaveYAML(results)

			// now shell out to ytt
			c := exec.Command("/usr/bin/ytt","--debug","--dangerous-allow-all-symlink-destinations", fmt.Sprintf("-f=%v/sonobuoy_results.yaml",ph.GetResultsDir()),fmt.Sprintf("-f=%v/ytt-transform.yaml", os.Getenv("SONOBUOY_CONFIG_DIR")),fmt.Sprintf("--output-files=%v", ph.GetResultsDir()))
			b,err := c.CombinedOutput()
			if err != nil {
				logrus.Trace(string(b))
				logrus.Error(err)
				return err
			}
			logrus.Trace(string(b))
			logrus.Trace("Done with processing")

			return nil
		},
	}
}

func getResultsFileName()string{
	return filepath.Join(os.Getenv("SONOBUOY_RESULTS_DIR"), "sonobuoy_results.yaml")
}

func SaveYAML(item results.Item) error {
	resultsFile := getResultsFileName()
	if err := os.MkdirAll(filepath.Dir(resultsFile), 0755); err != nil {
		return errors.Wrap(err, "error creating plugin directory")
	}

	outfile, err := os.Create(resultsFile)
	if err != nil {
		return errors.Wrap(err, "error creating results file")
	}
	defer outfile.Close()

	enc := yaml.NewEncoder(outfile)
	defer enc.Close()
	err = enc.Encode(item)
	return errors.Wrap(err, "error writing to results file")
}

func waitForDone()  error{
	logrus.WithField("waitfile", donefile).Info("Waiting for waitfile")
	ticker := time.NewTicker(time.Duration(1) * time.Second)
	donefilePath := filepath.Join(ph.GetResultsDir(),donefile)

	for {
		select {
		case <-ticker.C:
			if resultFile, err := ioutil.ReadFile(donefilePath); err == nil {
				resultFile = bytes.TrimSpace(resultFile)
				logrus.WithField("resultFile", string(resultFile)).Info("Detected done file, continuing with post-processing...")
				if err := os.Remove(donefilePath); err != nil {
					logrus.Errorf("Failed to remove donefile; postprocessing may end in race: %v", err)
				}
				return nil
			}
		}
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	logrus.SetLevel(logrus.TraceLevel)
	err := getRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

