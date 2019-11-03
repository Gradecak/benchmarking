package main

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/gradecak/benchmark/pkg/clients"
	"github.com/gradecak/benchmark/pkg/collector"
	"github.com/gradecak/benchmark/pkg/provenance"
	"github.com/prometheus/prom2json"
	"github.com/sirupsen/logrus"
)

type ProvIngestExp struct {
	maxQPS      int
	QPSinterval int

	url   string
	label string

	collector *collector.Collector
	rand      *rand.Rand
}

type ProvResult struct {
	Timestamp time.Time
	QPS       int
	State     *prom2json.Family
	ExpLabel  string
}

func setupProvExp(cnf *ExperimentConf) (Experiment, error) {
	var maxQPS, qpsInterval int
	var ok bool
	if maxQPS, ok = cnf.ExpParams["maxQPS"].(int); !ok {
		return nil, errors.New("Invalid maxQPS value in config")
	}
	if qpsInterval, ok = cnf.ExpParams["qpsInterval"].(int); !ok {
		return nil, errors.New("Invalid qps interval value in config")
	}

	return &ProvIngestExp{
		maxQPS:      maxQPS,
		QPSinterval: qpsInterval,
		url:         cnf.Url,
		label:       cnf.ExpLabel,
		collector:   cnf.collector,
		rand:        rand.New(rand.NewSource(time.Now().Unix())),
	}, nil

}

func (exp ProvIngestExp) Run(ctx Context) (interface{}, error) {
	output := []ProvResult{}
	provGen := provenance.NewGenerator()
	for i := exp.QPSinterval; i < exp.maxQPS; i += exp.QPSinterval {
		//setup collector data reciever
		collectorChan := make(chan *collector.DataPoint, 2000)
		cCtx, cancel := context.WithCancel(ctx)

		//warmup experiment treatment
		logrus.Infof("Warming up %v QPS bracket...", i)
		err := exp.warmupBracket(i, provGen)
		if err != nil {
			panic(err)
		}

		//initialise client
		nats, err := clients.NewNatsClient(&clients.NATSConf{exp.url, "test-cluster"})
		if err != nil {
			return nil, err
		}

		//start the experiment
		c, _ := context.WithDeadline(ctx, time.Now().Add(BRACKET_DURATION))
		ticker := time.NewTicker(time.Duration(1e9 / i))
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			exp.collector.CollectUntilStable(cCtx, collectorChan)
			wg.Done()
		}(&wg)
		logrus.Info("Starting Experiment...")
		func() {
			for {
				select {
				case <-ticker.C:
					wg.Add(1)
					go func(wg *sync.WaitGroup) {
						defer wg.Done()
						p := provGen.NewRandomProv(3)
						nats.PublishProto("PROVENANCE", p)
					}(&wg)
				case <-c.Done():
					return
				}
			}
		}()
		logrus.Info("Waiting for lads to finish...")
		wg.Wait()
		cancel()
		logrus.Infof("Grouping results for %v throughput bracket \n", i)

		// aggregate collected data
		for r := range collectorChan {
			for _, fam := range r.Data {
				output = append(output, ProvResult{r.TimeStamp, i, fam, exp.label})
			}
		}

	}
	return output, nil
}

func (exp ProvIngestExp) warmupBracket(qps int, gen *provenance.Gen) error {
	nats, err := clients.NewNatsClient(&clients.NATSConf{exp.url, "test-cluster"})
	if err != nil {
		return err
	}
	ticker := time.NewTicker(time.Duration(1e9 / qps))
	c, _ := context.WithDeadline(context.TODO(), time.Now().Add(WARMUP_DURATION))
	for {
		select {
		case <-ticker.C:
			go func() {
				p := gen.NewRandomProv(3)
				nats.PublishProto("PROVENANCE", p)
			}()
		case <-c.Done():
			return nil
		}
	}
}
