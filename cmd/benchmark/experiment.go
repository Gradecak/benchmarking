package main

import (
	"fmt"
	"io/ioutil"

	"github.com/gradecak/benchmark/pkg/collector"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var exps = map[string]ExperimentConstructor{
	"throughput":   setupThroughput,
	"wfSerial":     setupSerialExp,
	"ttr":          setupTTRExp,
	"ttf":          setupTTFExp,
	"provIngest":   setupProvExp,
	"policyUnique": setupPolicyUniqueExp,
	"policy":       setupPolicyExp,
}

type ExperimentConstructor = func(*ExperimentConf) (Experiment, error)

type Experiment interface {
	Run(Context) (interface{}, error)
}

type ExperimentConf struct {
	Url           string
	ExpName       string                    `yaml:"expName"`
	ExpLabel      string                    `yaml:"expLabel"`
	WfSpec        string                    `yaml:"wfSpec"`
	OutputFile    string                    `yaml:"outputFile"`
	ExpParams     map[string]interface{}    `yaml:"expParams"`
	CollectorConf collector.CollectorConfig `yaml:"collector"`
	collector     *collector.Collector
}

func ConfigFromFile(filename string) (*ExperimentConf, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	e := &ExperimentConf{}
	err = yaml.Unmarshal(data, e)
	if err != nil {
		return nil, err
	}
	logrus.Infof("Parsed Config %+v", e)
	return e, nil
}

func GetExperiment(c *ExperimentConf) (Experiment, error) {
	// intialise state collector
	collector, err := collector.New(&c.CollectorConf)
	c.collector = collector
	if err != nil {
		return nil, err
	}
	// initialise experiment
	if expConstructor, ok := exps[c.ExpName]; ok {
		return expConstructor(c)
	}
	return nil, fmt.Errorf("Cannot find experiment %v", c.ExpName)
}
