package main

import (
	"context"
	"errors"
	"github.com/gradecak/benchmark/pkg/collector"
	"github.com/prometheus/prom2json"
	"github.com/sirupsen/logrus"
	// "strconv"
	"sync"
	"time"
)

const (
	BRACKET_DURATION = time.Minute * 1
	WARMUP_DURATION  = time.Second * 20
)

type ThroughputExp struct {
	maxThroughput       int
	throughputIntervals int
	wfID                string // Id of workflow to execute
	url                 string
	collector           *collector.Collector
	expLabel            string
}

type ThroughputResult struct {
	ThroughputBracket int
	Timestamp         time.Time
	State             *prom2json.Family
	ExpLabel          string
}

func setupThroughput(cnf *ExperimentConf) (Experiment, error) {
	t := &ThroughputExp{
		expLabel:            cnf.ExpLabel,
		collector:           cnf.collector,
		url:                 cnf.Url,
		maxThroughput:       0,
		throughputIntervals: 0,
	}
	// parse throughput brackets
	maxThroughput, ok := cnf.ExpParams["maxThroughput"]
	if !ok {
		return nil, errors.New("Cannot find throughput treatments in experiment config")
	}
	if maxThroughput, ok := maxThroughput.(int); ok {
		t.maxThroughput = maxThroughput
	}
	throughputIntervals, ok := cnf.ExpParams["throughputIntervals"]
	if !ok {
		return nil, errors.New("Cannot find throughput treatments in experiment config")
	}
	if throughputIntervals, ok := throughputIntervals.(int); ok {
		t.throughputIntervals = throughputIntervals
	}

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

func (t ThroughputExp) warmup(throughput int) error {
	client, err := NewFWClient(t.url)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(time.Duration(1e9 / throughput))
	ctx, ca := context.WithDeadline(context.TODO(), time.Now().Add(WARMUP_DURATION))
	defer ca()
	for {
		select {
		case <-ticker.C:
			go func() {
				_, err := client.Invoke(Context{}, t.wfID)
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

func (t ThroughputExp) Run(ctx Context) (interface{}, error) {
	output := []ThroughputResult{}

	for throughput := t.throughputIntervals; throughput < t.maxThroughput; throughput += t.throughputIntervals {
		logrus.Infof("Warming up for throughput bracket (%v)\n", throughput)
		t.warmup(throughput)
		logrus.Infof("Starting experiment (%v)\n", throughput)
		// setup collector buffer
		stateChan := make(chan *collector.DataPoint,
			(BRACKET_DURATION/time.Second)/t.collector.GetRate()+1000)
		collectorContext, ccCancel := context.WithCancel(ctx)
		// start collecting FW state information
		go t.collector.Collect(collectorContext, stateChan)
		// get FW client
		client, err := NewFWClient(t.url)
		if err != nil {
			logrus.Fatal(err.Error())
			return nil, err
		}

		c, b := context.WithDeadline(ctx, time.Now().Add(BRACKET_DURATION))
		ticker := time.NewTicker(time.Duration(1e9 / throughput))
		wg := sync.WaitGroup{}
		func() {
			for {
				select {
				case <-ticker.C:
					wg.Add(1)
					go func(wg *sync.WaitGroup) {
						defer wg.Done()
						_, err := client.Invoke(Context{}, t.wfID)
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

		//get the FW client

		// } else { // dont do anything just collect state
		//	time.Sleep(time.Minute)
		// }
		logrus.Info("Waiting for Lads to finish...")
		wg.Wait()
		// stop the collector and process the results
		ccCancel()
		b()
		logrus.Infof("Collecting results for %v throughput bracket...\n", throughput)
		for r := range stateChan {
			for _, fam := range r.Data {
				output = append(output, ThroughputResult{throughput, r.TimeStamp, fam, t.expLabel})
			}
		}
	}
	return output, nil
}
