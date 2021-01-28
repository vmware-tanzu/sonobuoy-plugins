package internal

import (
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Item defines an Item within a ReportItem
type Item struct {
	Name    string                 `yaml:"name"`
	Status  string                 `yaml:"status"`
	Details map[string]interface{} `yaml:"details"`
}

// ReportItem defines a set of Items
type ReportItem struct {
	Name   string `yaml:"name"`
	Status string `yaml:"status"`
	Meta   struct {
		File string `yaml:"file"`
		Type string `yaml:"type"`
	} `yaml:"meta"`
	Items []Item `yaml:"items"`
}

// Report is the overall output of the program to be returned to Sonobuoy
type Report struct {
	Name   string `yaml:"name"`
	Status string `yaml:"status"`
	Meta   struct {
		Type string `yaml:"type"`
	} `yaml:"meta"`
	Items []ReportItem
}

// BuildReport given data returns a Report ready for writing
func (runner *Runner) BuildReport(cc int, name string) *Report {
	runner.Logger.WithFields(log.Fields{
		"component": "runner",
		"phase":     "report",
	}).Info("building")

	report := Report{
		Name:   name,
		Status: "passed",
	}
	for len(report.Items) < cc {
		item := <-runner.Results
		for _, check := range item.Items {
			if check.Status != "passed" {
				item.Status = "failed"
			}
		}
		if item.Status != "passed" {
			report.Status = "failed"
		}
		report.Items = append(report.Items, item)
	}
	return &report
}

// WriteReport writes a Report out to a path on disk
func (runner *Runner) WriteReport(report *Report, path string) error {
	resultsFilepath := "/tmp/results/done"

	runner.Logger.WithFields(log.Fields{
		"component": "runner",
		"phase":     "report",
	}).Info(fmt.Sprintf("writing to %s\n", resultsFilepath))

	out, err := yaml.Marshal(report)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, out, 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(resultsFilepath, []byte(path), 0644)
	if err != nil {
		return err
	}
	return nil
}
