package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var exps = map[string]ExperimentConstructor{
	"throughput": setupThroughput,
}

type ExperimentConstructor = func(*ExperimentConf) (Experiment, error)

type Experiment interface {
	Run(Context) ([][]string, error)
}

type ExperimentConf struct {
	Url        string
	ExpName    string                 `yaml:"expName"`
	WfSpec     string                 `yaml:"wfSpec"`
	OutputFile string                 `yaml:"outputFile"`
	ExpParams  map[string]interface{} `yaml:"expParams"`
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

func GetExperiment(expName string, c *ExperimentConf) (Experiment, error) {
	if expConstructor, ok := exps[expName]; ok {
		return expConstructor(c)
	}
	return nil, fmt.Errorf("Cannot find experiment %v", expName)
}
