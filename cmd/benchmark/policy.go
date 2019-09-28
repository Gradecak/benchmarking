package main

import (
	"context"
	"github.com/prometheus/prom2json"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/gradecak/benchmark/pkg/collector"
	"github.com/gradecak/benchmark/pkg/workflows"
)

type PolicyExp struct {
	throughput      int
	warmupDuration  time.Duration
	bracketDuration time.Duration
	url             string
	expLabel        string
	collector       *collector.Collector
}

type PolicyResult struct {
	ExpLabel  string
	State     *prom2json.Family
	Timestamp time.Time
}

func setupPolicyExp(cnf *ExperimentConf) (Experiment, error) {
	return &PolicyExp{
		throughput:      6,
		warmupDuration:  time.Duration(time.Minute),
		bracketDuration: time.Duration(time.Second * 30),
		url:             cnf.Url,
		expLabel:        "test",
		collector:       cnf.collector,
	}, nil
}

func (exp PolicyExp) Run(c Context) (interface{}, error) {
	var (
		output            = []PolicyResult{}
		ticker            = time.NewTicker(time.Duration(1e9 / exp.throughput))
		ctx, cancel       = context.WithDeadline(c, time.Now().Add(exp.bracketDuration))
		colCtx, colCancel = context.WithCancel(c)
		stateChan         = make(chan *collector.DataPoint, (BRACKET_DURATION/time.Second)/exp.collector.GetRate()+1000)
		wfPool            = make([]string, 100)
	)

	// get FW client
	client, err := NewFWClient(exp.url)
	if err != nil {
		logrus.Fatal(err.Error())
		return nil, err
	}

	// setup pool of workflows spec
	for i, _ := range wfPool {
		wfSpec := workflows.NewWorkflow(1, 3, &workflows.WorkflowConfig{
			TaskRuntime:    "1s",
			RandomTaskName: true,
		})
		wfId, err := client.SetupWfFromSpec(c, wfSpec)
		if err != nil {
			logrus.Panic(err)
		}
		wfPool[i] = wfId
	}
	wfIndex := 0

	//run
	go exp.collector.Collect(colCtx, stateChan)
	func() {
		for {
			select {
			case <-ticker.C:
				_, err := client.Invoke(Context{}, wfPool[wfIndex])
				wfIndex++
				if err != nil {
					logrus.Error(err)
					return
				}

			case <-ctx.Done():
				return
			}
		}
	}()
	colCancel()
	cancel()
	for r := range stateChan {
		for _, fam := range r.Data {
			output = append(
				output,
				PolicyResult{exp.expLabel, fam, r.TimeStamp},
			)
		}
	}
	return output, nil
}
