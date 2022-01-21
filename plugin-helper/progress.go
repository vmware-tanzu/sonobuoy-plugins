package plugin_helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/sonobuoy/pkg/plugin"
)

const (
	SonobuoyProgressPortEnvKey = "SONOBUOY_PROGRESS_PORT"
)

type ProgressReporter struct {
	total, completed int64
	failures, errors []string
	c                *http.Client
	port             string
}

// NewProgressReporter will initialize a progress reporter which expects the given number of tests. If
// it fails to generate a reporter, it will return the empty reporter which executes noops.
func NewProgressReporter(total int64) ProgressReporter {
	progressPort := os.Getenv(SonobuoyProgressPortEnvKey)
	if progressPort == "" {
		logrus.Tracef("No %v env var set; no progress updates will be sent.", SonobuoyProgressPortEnvKey)
		return ProgressReporter{}
	}
	logrus.Tracef("ProgressReporter created with %v total tests expected. Will send requests to localhost:%v", total, progressPort)
	return ProgressReporter{total: total, c: &http.Client{Timeout: 30 * time.Second}, port: progressPort}
}

// StartTest will send a progress update indicating the start of the given test.
func (r *ProgressReporter) StartTest(name string) {
	r.SendMessage(fmt.Sprintf("Test started: %v", name))
}

// StopTest will increase the tests counts and send an update message accordingly.
func (r *ProgressReporter) StopTest(name string, failed, skipped bool, err error) {
	msg := ""
	if failed {
		// Completed count not incremented when failing tests added.
		r.failures = append(r.failures, name)
		msg = fmt.Sprintf("Test failed: %v", name)
	} else if skipped {
		r.completed += 1
		msg = fmt.Sprintf("Test skipped: %v", name)
	} else if err != nil {
		r.completed += 1
		r.errors = append(r.errors, name)
		msg = fmt.Sprintf("Test errored: %v %v", name, err.Error())
	} else {
		r.completed += 1
		msg = fmt.Sprintf("Test completed: %v", name)
	}
	r.SendMessage(msg)
}

// SendMessage should be used for sending arbitrary messages. This method waits for a response,
// use SendMessageAsync for an asynchronous call.
func (r *ProgressReporter) SendMessage(msg string) error {
	if r.c == nil {
		logrus.Warnln("Progress update attempted but no client available.")
		return nil
	}

	update := plugin.ProgressUpdate{
		Timestamp: time.Time{},
		Message:   msg,
		Total:     r.total,
		Completed: r.completed,
		Errors:    r.errors,
		Failures:  r.failures,
	}
	b, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal progress update: %w", err)
	}

	resp, err := r.c.Post(fmt.Sprintf("http://localhost:%v/progress", r.port), "", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to POST progress update: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP Status from progress update: %v (%v)", resp.Status, resp.StatusCode)
	}
	return nil
}

func (r *ProgressReporter) SendMessageAsync(msg string) {
	go func() {
		err := r.SendMessage(msg)
		if err != nil {
			logrus.Errorf("Failed to send progress update: %v", err)
		}
	}()
}
