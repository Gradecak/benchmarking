package main

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

type ThroughputExp struct {
	throughputBrackets []int  // number of runs to perform per client
	wfID               string // Id of workflow to execute
	client             *FWClient
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
	t := &ThroughputExp{}
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
	t.client = NewFWClient(cnf.Url)
	wfID, err := t.client.setupWF(Context{}, cnf.WfSpec)
	if err != nil {
		return nil, err
	}
	t.wfID = wfID
	return t, nil

}

func (t ThroughputExp) Run(ctx Context) ([][]string, error) {
	output := [][]string{[]string{"latency", "qps"}}
	for _, throughput := range t.throughputBrackets {
		tick := time.NewTicker(time.Duration(1e9 / throughput))
		resultChan := make(chan *Result, throughput*3)
		c, ca := context.WithDeadline(ctx, time.Now().Add(3*time.Second))
		wg := sync.WaitGroup{}
		defer ca()

		// make the invocation
		func() {
			for {
				select {
				case <-tick.C:
				case <-c.Done():
					return
				}
				wg.Add(1)
				go func(w *sync.WaitGroup) {
					defer w.Done()
					r, err := t.client.Invoke(ctx, t.wfID)
					if err != nil {
						logrus.Error("Invocation Returned Error")
						return
					}
					resultChan <- r
				}(&wg)
			}
		}()
		wg.Wait()
		close(resultChan)

		for r := range resultChan {
			output = append(output, []string{strconv.FormatInt(int64(r.duration), 10), strconv.Itoa(throughput)})
		}

	}
	return output, nil
}
