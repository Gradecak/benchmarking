package main

import (
	"context"
	"errors"
	"github.com/gradecak/benchmark/pkg/collector"
	"github.com/prometheus/prom2json"
	"github.com/sirupsen/logrus"
	// "strconv"
	// "sync"
	"time"
)

const (
	BRACKET_DURATION = time.Minute * 1
)

type ThroughputExp struct {
	throughputBrackets []int  // number of runs to perform per client
	wfID               string // Id of workflow to execute
	url                string
	collector          *collector.Collector
	expLabel           string
}

type ThroughputResult struct {
	ThroughputBracket int
	Timestamp         time.Time
	State             *prom2json.Family
	ExpLabel          string
}

func parseThroughputBrackets(brackets []interface{}) ([]int, error) {
	ret := []int{}
	for _, bracket := range brackets {
		intB, ok := bracket.(int)
		if !ok {
			return nil, errors.New("Cannot convert throughput bracket to integer")
		}
		ret = append(ret, intB)
	}
	return ret, nil
}

func setupThroughput(cnf *ExperimentConf) (Experiment, error) {
	t := &ThroughputExp{
		expLabel:  cnf.ExpLabel,
		collector: cnf.collector,
		url:       cnf.Url,
	}
	// parse throughput brackets
	throughputs, ok := cnf.ExpParams["throughput"]
	if !ok {
		return nil, errors.New("Cannot find throughput treatments in experiment config")
	}
	tb, ok := throughputs.([]interface{})
	if !ok {
		return nil, errors.New("Throughputs must be list of integers")
	}
	treatments, err := parseThroughputBrackets(tb)
	if err != nil {
		return nil, err
	}
	t.throughputBrackets = treatments

	client, err := NewFWClient(cnf.Url)
	if err != nil {
		return nil, err
	}
	wfID, err := client.setupWF(Context{}, cnf.WfSpec)
	if err != nil {
		return nil, err
	}
	t.wfID = wfID
	return t, nil
}

func (t ThroughputExp) Run(ctx Context) (interface{}, error) {
	// output := make([]ThroughputResult, len(t.throughputBrackets))
	output := []ThroughputResult{}
	for _, throughput := range t.throughputBrackets {
		logrus.Infof("Starting invocations for %v for throughput bracket\n", throughput)
		tick := time.NewTicker(time.Duration(1e9 / throughput))
		stateChan := make(chan *collector.DataPoint, BRACKET_DURATION/collector.DEFAULT_RATE)
		c, ca := context.WithDeadline(ctx, time.Now().Add(BRACKET_DURATION))
		defer ca()
		// make the invocation
		func() {
			// start collecting FW state information
			go t.collector.Collect(c, stateChan)
			// start simulating workload
			client, err := NewFWClient(t.url)
			if err != nil {
				logrus.Fatal(err.Error())
				return
			}
			for {
				select {
				case <-tick.C:
				case <-c.Done():
					return
				}
				go func() {
					_, err := client.Invoke(ctx, t.wfID)
					if err != nil {
						logrus.Error(err)
						return
					}
				}()
			}
		}()

		logrus.Infof("Collecting results for %v throughput bracket...\n", throughput)
		for r := range stateChan {
			for _, fam := range r.Data {
				output = append(output, ThroughputResult{throughput, r.TimeStamp, fam, t.expLabel})
			}
		}
	}
	return output, nil
}
