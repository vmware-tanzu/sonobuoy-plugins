package internal

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
)

var (
	CheckStartMsg    string = "starting"
	CheckWriteMsg    string = "writing-result"
	CheckCompleteMsg string = "complete"
)

// QuerierConfig provides generic configuration options for a Querier
type QuerierConfig struct {
	Context  context.Context
	Results  chan ReportItem
	Logger   *log.Logger
	Complete chan struct{}
}

// Querier is a gather of information that can be ran by the Runner.
type Querier interface {
	Start(*QuerierConfig)
}

// Runner runs the included Queriers, based on the configured Checks provided.
type Runner struct {
	Config   *ReliabilityConfig
	Context  context.Context
	Queriers []Querier
	Results  chan ReportItem
	Complete chan struct{}
	Logger   *log.Logger
}

// Run runs the Runner
func (runner Runner) Run() {
	if len(runner.Config.Checks) < 1 {
		runner.Logger.WithFields(log.Fields{
			"component": "runner",
			"phase":     "run",
		}).Error("no checks configured")
		os.Exit(1)
	}

	runner.Logger.WithFields(log.Fields{
		"component": "runner",
		"phase":     "run",
	}).Info("waiting for checks to complete")

	for _, querier := range runner.Queriers {
		go querier.Start(&QuerierConfig{
			Context:  runner.Context,
			Results:  runner.Results,
			Logger:   runner.Logger,
			Complete: make(chan struct{}),
		})
	}
}
