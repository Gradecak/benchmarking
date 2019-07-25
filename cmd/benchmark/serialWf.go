package main

import (
	"context"
	"errors"
	"github.com/gradecak/benchmark/pkg/collector"
	"github.com/gradecak/benchmark/pkg/workflows"
	"github.com/prometheus/prom2json"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type SerialExp struct {
	maxWorkflowLength int
	interval          int
	startLength       int
	collector         *collector.Collector
	expLabel          string
	qps               int
	//
	url string
}

type SerialResult struct {
	WfLength  int
	Timestamp time.Time
	State     *prom2json.Family
	ExpLabel  string
}

func setupSerialExp(cnf *ExperimentConf) (Experiment, error) {
	exp := &SerialExp{
		expLabel:  cnf.ExpLabel,
		collector: cnf.collector,
		url:       cnf.Url,
	}
	// parse length
	maxLen, ok := cnf.ExpParams["maxLen"]
	if !ok {
		return nil, errors.New("Cannot find max serial length in config")
	}
	exp.maxWorkflowLength = maxLen.(int)
	//parse intervals
	intervals, ok := cnf.ExpParams["intervals"]
	if !ok {
		return nil, errors.New("Cannot find intervals in config")
	}
	exp.interval = intervals.(int)

	//parse intervals
	startValue, ok := cnf.ExpParams["startValue"]
	if !ok {
		return nil, errors.New("Cannot find startValue in config")
	}
	exp.startLength = startValue.(int)

	//parse qps
	qps, ok := cnf.ExpParams["qps"]
	if !ok {
		return nil, errors.New("Cannot find qps in config")
	}
	exp.qps = qps.(int)

	return exp, nil
}

func (exp SerialExp) warmup(wfId string) error {
	client, err := NewFWClient(exp.url)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(time.Duration(1e9 / exp.qps))
	ctx, ca := context.WithDeadline(context.TODO(), time.Now().Add(WARMUP_DURATION))
	defer ca()
	for {
		select {
		case <-ticker.C:
			go func() {
				_, err := client.Invoke(Context{}, wfId)
				if err != nil {
					logrus.Error(err)
					return
				}

			}()
		case <-ctx.Done():
			return nil
		}
	}
	return nil
}

func (exp SerialExp) Run(ctx Context) (interface{}, error) {
	output := []SerialResult{}

	for wfLength := exp.startLength; wfLength < exp.maxWorkflowLength; wfLength += exp.interval {
		//create workflow
		logrus.Infof("Creating Wf Serial Length: %d", wfLength)
		client, err := NewFWClient(exp.url)
		if err != nil {
			logrus.Fatal(err.Error())
			return nil, err
		}
		spec := workflows.NewWorkflow(1, wfLength)
		wfId, err := client.SetupWfFromSpec(ctx, spec)
		if err != nil {
			logrus.Fatal("Cannot setup workflow length (%d)", wfLength)
		}

		//setup Collector
		collectorChan := make(chan *collector.DataPoint,
			(int(BRACKET_DURATION.Seconds())*exp.qps)/int(exp.collector.GetRate().Seconds())+100)
		colCtx, cancel := context.WithCancel(ctx)

		//warmup for given workflow size
		exp.warmup(wfId)

		// start experiment and collector
		ticker := time.NewTicker(time.Duration(1e9 / exp.qps))
		c, _ := context.WithDeadline(ctx, time.Now().Add(BRACKET_DURATION))
		go exp.collector.Collect(colCtx, collectorChan)
		wg := sync.WaitGroup{}
		func() {
			for {
				select {
				case <-ticker.C:
					wg.Add(1)
					go func(wg *sync.WaitGroup) {
						defer wg.Done()
						_, err := client.Invoke(Context{}, wfId)
						if err != nil {
							logrus.Error(err)
							return
						}
					}(&wg)
				case <-c.Done():
					return
				}
			}
		}()

		logrus.Info("Waiting for the Boysh to finish...")
		wg.Wait()
		cancel()

		logrus.Infof("Collecting results for %s wf length...\n", wfLength)
		for r := range collectorChan {
			for _, fam := range r.Data {
				output = append(output,
					SerialResult{wfLength, r.TimeStamp, fam, exp.expLabel})
			}
		}

	}

	return output, nil
}
