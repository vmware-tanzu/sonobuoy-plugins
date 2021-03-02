package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/vmware-tanzu/sonobuoy-plugins/reliability-scanner/api/v1alpha1/namespace/labels"
	"github.com/vmware-tanzu/sonobuoy-plugins/reliability-scanner/api/v1alpha1/pod/probes"
	"github.com/vmware-tanzu/sonobuoy-plugins/reliability-scanner/api/v1alpha1/pod/qos"
	"github.com/vmware-tanzu/sonobuoy-plugins/reliability-scanner/internal"
)

func scan() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
	logger.WithField("component", "main")
	logger.WithField("phase", "config")
	cfg := os.Getenv("CONFIG")
	f, err := os.Create("./config.yaml")
	if err != nil {
		logger.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		logger.Fatal(err)
	}

	var c internal.ReliabilityConfig

	err = viper.Unmarshal(&c)
	if err != nil {
		logger.Fatal(err)
	}

	runner := &internal.Runner{
		Context:  context.Background(),
		Results:  make(chan internal.ReportItem),
		Logger:   logger,
		Complete: make(chan struct{}),
	}
	for _, checkCfg := range c.Checks {
		switch checkCfg.Kind {
		case "v1alpha1/pod/probes":
			querier, err := probes.NewQuerier(&probes.QuerierSpec{})
			if err != nil {
				runner.Logger.WithFields(logrus.Fields{
					"kind":  checkCfg.Spec["kind"],
					"name":  checkCfg.Spec["name"],
					"phase": "add",
				}).Error(err)
			}
			querier.AddtoRunner(runner)
		case "v1alpha1/pod/qos":
			includeDetail, err := strconv.ParseBool(checkCfg.Spec["include_detail"])
			if err != nil {
				runner.Logger.WithFields(logrus.Fields{
					"kind":  checkCfg.Spec["kind"],
					"name":  checkCfg.Spec["name"],
					"phase": "add",
				}).Error(err)
			}
			querier, err := qos.NewQuerier(&qos.QuerierSpec{
				IncludeDetail:          includeDetail,
				MinimumDesiredQOSClass: checkCfg.Spec["minimum_desired_qos_class"],
			})
			if err != nil {
				runner.Logger.WithFields(logrus.Fields{
					"kind":  checkCfg.Spec["kind"],
					"name":  checkCfg.Spec["name"],
					"phase": "add",
				}).Error(err)
			}
			querier.AddtoRunner(runner)
		case "v1alpha1/namespace/labels":
			includeLabels, err := strconv.ParseBool(checkCfg.Spec["include_labels"])
			if err != nil {
				runner.Logger.WithFields(logrus.Fields{
					"check_name": checkCfg.Spec["Name"],
					"phase":      "add",
				}).Error(err)
			}
			querier, err := labels.NewQuerier(&labels.QuerierSpec{
				Key:           checkCfg.Spec["key"],
				IncludeLabels: includeLabels,
			})
			if err != nil {
				runner.Logger.WithFields(logrus.Fields{
					"check_name": checkCfg.Spec["Name"],
					"phase":      "add",
				}).Error(err)
			}
			querier.AddtoRunner(runner)
		}
	}
	logger.Info(fmt.Sprintf("Configured %s checks.", len(c.Checks)))
	runner.Run()
	report := runner.BuildReport(len(c.Checks), reportName)
	err = runner.WriteReport(report, filePath)
	if err != nil {
		logger.Error(err)
	}
	logger.Info("Reliability Scan Complete.")
	for {
	}
}
