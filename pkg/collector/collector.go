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
	Output   string
}

type DataPoint struct {
	TimeStamp time.Time
	Data      []*prom2json.Family
}

func New(endpoints []*Collection, rate time.Duration) *Collector {
	return &Collector{endpoints, &http.Client{}, rate}
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
			if err != nil {
				logrus.Warn(err.Error())
				// return err
			}
			// extract relevant information
			for _, label := range target.Interest {
				if mf, ok := mfs[label]; ok {
					res.Data = append(res.Data, prom2json.NewFamily(mf))
				} else {
					logrus.Warnf("Label %v not found in collected data\n", label)
				}
			}
		}
		out <- res
	}
}
