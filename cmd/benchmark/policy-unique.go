package main

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/prom2json"
	"github.com/sirupsen/logrus"

	"github.com/gradecak/benchmark/pkg/collector"
	"github.com/gradecak/benchmark/pkg/workflows"
)

type PolicyUniqueExp struct {
	throughput   int
	poolSize     int
	url          string
	simfaasUrl   string
	expLabel     string
	maxColdStart time.Duration
	collector    *collector.Collector
}

type PolicyUniqueResult struct {
	ExpLabel  string
	State     *prom2json.Family
	ColdStart time.Duration
	Timestamp time.Time
}

func setupPolicyUniqueExp(cnf *ExperimentConf) (Experiment, error) {
	return &PolicyUniqueExp{
		throughput:   6,
		poolSize:     10,
		url:          "127.0.0.1:5555",
		simfaasUrl:   "http://127.0.0.1:9000",
		maxColdStart: time.Second * 20,
		expLabel:     "test",
		collector:    cnf.collector,
	}, nil
}

func (exp PolicyUniqueExp) SetColdStart(duration time.Duration) error {
	req, err := http.NewRequest("GET", exp.simfaasUrl+"/set/cold-start", nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("cs", duration.String())
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 200 {
		return errors.New("Server Error")
	}
	return nil
}

func (exp PolicyUniqueExp) Run(c Context) (interface{}, error) {
	var (
		output = []PolicyUniqueResult{}
		ticker = time.NewTicker(time.Duration(1e9 / exp.throughput))
		wfPool = make([]string, exp.poolSize)
	)

	// get FW client
	client, err := NewFWClient(exp.url)
	if err != nil {
		logrus.Fatal(err.Error())
		return nil, err
	}

	//run
	for coldStart := time.Second; coldStart <= exp.maxColdStart; coldStart = coldStart + time.Second {

		// setup pool of workflow specs
		for i, _ := range wfPool {
			wfSpec := workflows.NewWorkflow(1, 3, &workflows.WorkflowConfig{
				TaskRuntime:    "1",
				RandomTaskName: true,
				// PercentMultienvTasks: 1.0,
			})
			wfId, err := client.SetupWfFromSpec(c, wfSpec)
			if err != nil {
				logrus.Panic(err)
			}
			wfPool[i] = wfId
		}
		wfIndex := 0
		err := exp.SetColdStart(coldStart)
		if err != nil {
			logrus.Errorf("Error setting cold start (reason %v)", err)
			return nil, err
		}
		colCtx, colCancel := context.WithCancel(c)
		stateChan := make(chan *collector.DataPoint, 1000)
		go exp.collector.Collect(colCtx, stateChan)
		wg := &sync.WaitGroup{}
		// run experiment and collect data
		logrus.Infof("Running experiment with %v cold start", coldStart)
		func() {
			for {
				select {
				case <-ticker.C:
					wg.Add(1)
					go func(wg *sync.WaitGroup, i int) {
						defer wg.Done()
						_, err := client.Invoke(Context{}, wfPool[i])
						if err != nil {
							logrus.Error(err)
							return
						}
					}(wg, wfIndex)
					wfIndex++
					if wfIndex == exp.poolSize {
						return
					}
				}
			}
		}()

		wg.Wait()
		colCancel()
		for r := range stateChan {
			for _, fam := range r.Data {
				output = append(
					output,
					PolicyUniqueResult{exp.expLabel, fam, coldStart, r.TimeStamp},
				)
			}
		}
	}
	return output, nil
}
