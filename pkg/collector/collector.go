package collector

import (
	"context"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/prom2json"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const (
	DEFAULT_RATE = time.Second * 5
)

type Collector struct {
	targets []*Collection
	client  *http.Client
	rate    time.Duration
}

type Collection struct {
	Endpoint string
	Interest []string
}

type CollectorConfig struct {
	Collectors   []*Collection `yaml:"collectors"`
	SamplingRate string        `yaml:"sampling_rate"`
}

type DataPoint struct {
	TimeStamp time.Time
	Data      []*prom2json.Family
}

func New(conf *CollectorConfig) (*Collector, error) {
	sampleRate, err := time.ParseDuration(conf.SamplingRate)
	if err != nil {
		return nil, err
	}
	return &Collector{conf.Collectors, &http.Client{}, sampleRate}, nil
}

func (c *Collector) GetRate() time.Duration {
	return c.rate
}

func (c *Collector) Collect(ctx context.Context, out chan *DataPoint) error {
	tick := time.NewTicker(c.rate)
	for {
		// rate controll
		select {
		case <-tick.C:
		case <-ctx.Done():
			logrus.Info("Closing collector chanel")
			close(out)
			return nil
		}
		logrus.Info("Collecting State...")
		// Sample
		res := &DataPoint{TimeStamp: time.Now()}
		for _, target := range c.targets {
			resp, err := c.client.Get(target.Endpoint)
			if err != nil {
				logrus.Warn(err.Error())
			}
			parser := &expfmt.TextParser{}
			mfs, err := parser.TextToMetricFamilies(resp.Body)
			resp.Body.Close()
			if err != nil {
				logrus.Warn(err.Error())
				// return err
			}
			// extract relevant information
			for _, label := range target.Interest {
				if mf, ok := mfs[label]; ok {
					res.Data = append(res.Data, prom2json.NewFamily(mf))
					logrus.Info(mf)
				} else {
					logrus.Warnf("Label %v not found in collected data\n", label)
				}
			}
		}
		out <- res
	}
}

func (c *Collector) CollectUntilStable(ctx context.Context, out chan *DataPoint) error {
	var (
		prev   float64 = 0
		tick           = time.NewTicker(c.rate)
		stable         = 0
	)

	for {
		// rate controll
		select {
		case <-tick.C:

		case <-ctx.Done():
			logrus.Info("Closing collector chanel")
			close(out)
			return nil
		}
		if stable == 3 {
			close(out)
			return nil
		}
		logrus.Info("Collecting State...")
		// Sample
		res := &DataPoint{TimeStamp: time.Now()}

		for _, target := range c.targets {
			resp, err := c.client.Get(target.Endpoint)
			if err != nil {
				logrus.Warn(err.Error())
			}
			parser := &expfmt.TextParser{}
			mfs, err := parser.TextToMetricFamilies(resp.Body)
			resp.Body.Close()
			if err != nil {
				logrus.Warn(err.Error())
				// return err
			}
			// extract relevant information
			for _, label := range target.Interest {
				if mf, ok := mfs[label]; ok {
					if mf.GetName() == "dispatcher_enforcment_count" {
						logrus.Infof("previous %v ~~ %v current (%v stable)", prev, mf.GetMetric()[0].GetCounter().GetValue(), stable)
						if prev == mf.GetMetric()[0].GetCounter().GetValue() {
							stable++
						} else {
							prev = mf.GetMetric()[0].GetCounter().GetValue()
						}
					}
					res.Data = append(res.Data, prom2json.NewFamily(mf))
				} else {
					logrus.Warnf("Label %v not found in collected data\n", label)
				}
			}
		}
		out <- res
	}
}
