package main

import (
	"github.com/gradecak/benchmark/pkg/benchmark"
	"github.com/gradecak/benchmark/pkg/experiments"
	"github.com/sirupsen/logrus"
)

func main() {
	conf := benchmark.Config{
		Runs:    1,
		Clients: 100,
		URL:     "http://35.204.112.154/fission-function/",
	}

	exp := benchmark.Experiment{
		Fn:          "privacy2",
		Interpreter: experiments.NewZoneExperiment(2, "NL"),
	}

	conf.SetExp(exp)

	logrus.Info("Starting experiment...")
	err := conf.RunExperiment()
	if err != nil {
		logrus.Error("Ooopsie Doopsie")
	}
	logrus.Info("Finished experiment...")
}
