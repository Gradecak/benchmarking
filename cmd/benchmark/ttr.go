package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/gradecak/benchmark/pkg/collector"
	"github.com/gradecak/fission-workflows/pkg/types"
	stan "github.com/nats-io/stan.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prom2json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type TTRExperiment struct {
	// exp constants
	wfSpecPath    string
	url           string
	expLabel      string
	maxGraphSize  int
	graphStepSize int

	// Event stream config
	natsURL     string
	natsCluster string

	// Data Collector
	collector *collector.Collector
}

type TTRResult struct {
	GraphSize int
	Timestamp time.Time
	State     *prom2json.Family
	ExpLabel  string
}

func setupTTRExp(cnf *ExperimentConf) (Experiment, error) {
	return TTRExperiment{
		wfSpecPath:    cnf.WfSpec,
		url:           cnf.Url,
		expLabel:      cnf.ExpLabel,
		maxGraphSize:  100,
		graphStepSize: 10,
		natsURL:       "127.0.0.1",
		natsCluster:   "test-cluster",
		collector:     cnf.collector,
	}, nil
}

func (exp TTRExperiment) Run(c Context) (interface{}, error) {
	// output := []TTRResult{}
	// create and start the dummy http service to emulate data deletion endpoint
	ctx, cancel := context.WithCancel(c)
	ss := newSimpleService()
	go ss.run(":9999", ctx)

	// collector state variables
	// stateChan := make(chan *collector.DataPoint, exp.collector.GetRate()+10000)
	// collectorContext, collectorCancel := context.WithCancel(c)

	//event stream
	eventInjector, err := NewEventInjector(exp.natsCluster, exp.natsURL, "CONSENT")
	if err != nil {
		return nil, err
	}

	cIds := []string{}
	total := 0
	for i := exp.graphStepSize; i < exp.maxGraphSize; i += exp.graphStepSize {
		// prime the provenance graph
		log.Infof("Current Bracket %d", i)
		client, err := NewFWClient(exp.url)
		if err != nil {
			return nil, err
		}
		for j := 0; j < i; j++ {
			wfID, err := client.SetupWfFromFile(c, exp.wfSpecPath)
			if err != nil {
				return nil, err
			}
			cId := RandomString(10)
			cIds = append(cIds, cId)
			client.InvokeWithConsentID(c, wfID, cId)
		}

		// start collecting data
		// go exp.collector.Collect(collectorContext, stateChan)

		total += i
		ticker := time.NewTicker(time.Second / 2) //eventInjector.Revoke(cId)
		//start measuring
		for _, cId := range cIds {
			select {
			case <-ticker.C:
				eventInjector.Revoke(cId)
				ss.revoke(cId)
			}

			//results to output
			// for r := range stateChan {
			// 	for _, fam := range r.Data {
			// 		output = append(output,
			// 			TTRResult{i, r.TimeStamp, fam, exp.expLabel})
			// 	}
			// }
		}
	}

	time.Sleep(time.Minute)
	cancel()
	return nil, nil
}

// ************************************************************ //
//         Simple Web server used for measuring TTR             //
// ************************************************************ //

type simpleService struct {
	times map[string]time.Time
}

func newSimpleService() *simpleService {
	return &simpleService{make(map[string]time.Time)}
}

func (ss *simpleService) run(url string, c context.Context) {
	router := mux.NewRouter()

	recoveryTime := prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "policy",
		Subsystem: "recovery",
		Name:      "time",
		Help:      "Time for watchdog to rectify policy violation",
		Objectives: map[float64]float64{
			0:    0.0001,
			0.01: 0.0001,
			0.1:  0.0001,
			0.25: 0.0001,
			0.5:  0.0001,
			0.75: 0.0001,
			0.9:  0.0001,
			1:    0.0001,
		},
	})

	prometheus.Register(recoveryTime)
	server := http.Server{Addr: url, Handler: router}

	router.Handle("/done/{consentID}", ss.Done(recoveryTime))
	router.Handle("/metrics", prometheus.Handler()) //Metrics endpoint for scrapping

	go func() {
		<-c.Done()
		if err := server.Shutdown(c); err != nil {
			log.Fatal(err)
		}
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}

}

func (ss simpleService) revoke(consentID string) {
	ss.times[consentID] = time.Now()
}

func (ss simpleService) Done(sumVec prometheus.Summary) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Revoked")
		vars := mux.Vars(r)
		consentID := vars["consentID"]
		// record time that the revocation request came in
		if startTime, ok := ss.times[consentID]; ok {
			sumVec.Observe(float64(time.Since(startTime)))
		}
		greet := fmt.Sprintf("Hello %s \n", consentID)
		w.Write([]byte(greet))
	}
}

//**********************************************************************
// Simple Event Injector for NATS
//**********************************************************************

type eventInjector struct {
	prefix string
	stan.Conn
}

func NewEventInjector(clusterName, natsURL, prefix string) (*eventInjector, error) {
	conn, err := stan.Connect(clusterName, RandomString(4), stan.NatsURL(natsURL))
	if err != nil {
		return nil, err
	}
	return &eventInjector{prefix, conn}, nil
}

func (i *eventInjector) Revoke(consentID string) error {
	consentMsg := &types.ConsentMessage{
		ID: consentID,
		Status: &types.ConsentStatus{
			Status: types.ConsentStatus_Status(1),
		},
	}

	buf, err := proto.Marshal(consentMsg)
	if err != nil {
		return err
	}
	i.Publish(i.prefix, buf)
	return nil
}
