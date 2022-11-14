package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
		Config:   &c,
		Context:  context.Background(),
		Results:  make(chan internal.ReportItem),
		Logger:   logger,
		Complete: make(chan struct{}),
	}
	err = initializeQueriers(runner)
	if err != nil {
		logger.Fatal(err)
	}

	runner.Run()
	report := runner.BuildReport(len(c.Checks), reportName)
	err = runner.WriteReport(report, os.Getenv("SONOBUOY_RESULTS_DIR"))
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	logger.Info("Reliability Scan Complete.")
	for {
	}
}
