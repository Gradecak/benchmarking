package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type TTRExperiment struct {
	wfSpecPath    string
	url           string
	expLabel      string
	maxGraphSize  int
	graphStepSize int
}

func setupTTRExp(cnf *ExperimentConf) (Experiment, error) {
	return TTRExperiment{cnf.WfSpec, cnf.Url, cnf.ExpLabel, 100, 10}, nil
}

func (exp TTRExperiment) Run(c Context) (interface{}, error) {
	// start the dummy service to emulate data deletion
	ctx, cancel := context.WithCancel(c)
	ss := newSimpleService()
	go ss.run(":9999", ctx)

	cIds := []string{}
	total := 0
	for i := 1; i < exp.maxGraphSize; i += exp.graphStepSize {

		// prime the provenance graph
		client, err := NewFWClient(exp.url)
		if err != nil {
			return nil, err
		}
		for j := 0; j < i; j++ {
			wfID, err := client.setupWF(c, exp.wfSpecPath)
			if err != nil {
				return nil, err
			}
			cIds[total+j] = RandomString(10)
			client.InvokeWithConsentID(c, wfID, cIds[total+j])
		}

		// TODO start collector

		total += i
		//start measuring
		for _, cId := range cIds {
			// TODO send consent revocation event
			ss.revoke(cId)
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
	return &simpleService{(make(map[string]time.Time))}
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
