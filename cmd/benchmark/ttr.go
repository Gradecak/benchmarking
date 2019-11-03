package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gradecak/benchmark/pkg/collector"
	"github.com/gradecak/benchmark/pkg/provenance"
	"github.com/gradecak/fission-workflows/pkg/provenance/graph"
	"github.com/gradecak/fission-workflows/pkg/types"
	stan "github.com/nats-io/stan.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prom2json"
	log "github.com/sirupsen/logrus"

	"strconv"
	"sync"
	// "strings"
)

type TTRExperiment struct {
	// exp constants
	url          string
	expLabel     string
	GraphSize    int
	MaxQPS       int
	QPSIntervals int

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
	var (
		maxQPS      int
		qpsInterval int
		graphSize   int
		ok          bool
	)

	if maxQPS, ok = cnf.ExpParams["maxQPS"].(int); !ok {
		return nil, errors.New("Invalid maxQPS value in config")
	}

	if qpsInterval, ok = cnf.ExpParams["qpsInterval"].(int); !ok {
		return nil, errors.New("Invalid qps interval value in config")
	}

	if graphSize, ok = cnf.ExpParams["graphSize"].(int); !ok {
		return nil, errors.New("Invalid graph size value in config")
	}

	return TTRExperiment{
		url:          cnf.Url,
		expLabel:     cnf.ExpLabel,
		GraphSize:    graphSize,
		MaxQPS:       maxQPS,
		QPSIntervals: qpsInterval,
		natsURL:      cnf.Url,
		natsCluster:  "test-cluster",
		collector:    cnf.collector,
	}, nil
}

func (exp TTRExperiment) warmup(qps int, ss *simpleService) error {
	ticker := time.NewTicker(time.Duration(1e9 / qps)) //consentInjector.Revoke(cId)
	// start collecting data
	bracketTimer, can := context.WithDeadline(context.TODO(), time.Now().Add(WARMUP_DURATION))
	defer can()
	wg := sync.WaitGroup{}
	func() {
		for {
			select {
			case <-ticker.C:
				wg.Add(1)
				go func(wg *sync.WaitGroup) {
					defer wg.Done()
					ss.revoke()
				}(&wg)
			case <-bracketTimer.Done():
				return
			}
		}
	}()
	wg.Wait()
	return nil
}

func (exp TTRExperiment) Run(c Context) (interface{}, error) {

	var (
		output  = []TTRResult{}
		cIds    = []string{}
		provGen = provenance.NewGenerator()
	)

	//event stream
	provInjector, err := NewEventInjector(exp.natsCluster, exp.natsURL, "PROVENANCE")
	consentInjector, err := NewEventInjector(exp.natsCluster, exp.natsURL, "CONSENT")
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(c)

	//prime the graph

	log.Info("Priming Consent Graph")
	for j := 0; j < exp.GraphSize; j++ {
		cId := uuid.New().String()
		cIds = append(cIds, cId)
		provEvent := provGen.NewProv(&provenance.Cnf{
			ID:     cId,
			NTasks: 1,
			TaskNodes: &graph.Node{
				Type: graph.Node_TASK,
				Op:   graph.Node_WRITE,
				Meta: "http://localhost:9999/done",
			},
		})
		err := provInjector.InjectProv(provEvent)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}
	log.Info("Sleeping for 2 minutes before begingin experiment")
	time.Sleep(time.Minute*1 + time.Second*30)
	// create and start the dummy http service to emulate data deletion endpoint
	ss := newSimpleService(consentInjector, cIds)
	go ss.run(":9999", ctx)

	for qps := exp.QPSIntervals; qps <= exp.MaxQPS; qps += exp.QPSIntervals {
		// collector state variables
		stateChan := make(chan *collector.DataPoint,
			(BRACKET_DURATION/time.Second)/exp.collector.GetRate()+1000)

		collectorContext, collectorCancel := context.WithCancel(c)
		// prime the provenance graph
		log.Infof("Warming up brakcet  %d", qps)
		// warmup the QPS Bracket
		exp.warmup(qps, ss)
		log.Infof("Starting Bracket %d", qps)
		ticker := time.NewTicker(time.Duration(1e9 / qps)) //consentInjector.Revoke(cId)
		// start collecting data
		bracketTimer, can := context.WithDeadline(c, time.Now().Add(BRACKET_DURATION))
		go exp.collector.Collect(collectorContext, stateChan)
		wg := sync.WaitGroup{}
		func() {
			for {
				select {
				case <-ticker.C:
					wg.Add(1)
					go func(wg *sync.WaitGroup) {
						defer wg.Done()
						ss.revoke()
					}(&wg)
				case <-bracketTimer.Done():
					return
				}
			}
		}()

		wg.Wait()
		can()
		collectorCancel()
		//results to output
		log.Infof("Collecting results for %v bracket...\n", qps)
		for r := range stateChan {
			for _, fam := range r.Data {
				output = append(output,
					TTRResult{qps, r.TimeStamp, fam, exp.expLabel})
			}
		}
		log.Info("Done")
	}
	cancel()
	return output, nil
}

// ************************************************************ //
//         Simple Web server used for measuring TTR             //
// ************************************************************ //

type simpleService struct {
	ids           []string
	index         int
	indexMu       *sync.Mutex
	outstanding   uint
	outstandingMu *sync.RWMutex
	injector      *eventInjector
}

func newSimpleService(i *eventInjector, cids []string) *simpleService {
	log.Infof("Number of ids %v", len(cids))
	log.Infof("%v \n %v \n %v", cids[0], cids[1], cids[2])
	return &simpleService{
		ids:           cids,
		index:         0,
		indexMu:       &sync.Mutex{},
		outstanding:   0,
		outstandingMu: &sync.RWMutex{},
		injector:      i,
	}
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

func (ss *simpleService) revoke() {
	//pop from list of available
	ss.indexMu.Lock()
	consentID := ss.ids[ss.index]
	ss.index = ss.index + 1
	if ss.index == len(ss.ids) {
		ss.index = 0
	}
	ss.indexMu.Unlock()
	t := time.Now().UnixNano()
	query := consentID + "/$/" + strconv.FormatInt(t, 10)
	ss.injector.Revoke(query)
}

func (ss *simpleService) Done(sumVec prometheus.Summary) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		unixString := vars["consentID"]
		unixInt, err := strconv.ParseInt(unixString, 10, 64)
		if err != nil {
			panic(err)
		}
		startTime := time.Unix(0, unixInt)
		sumVec.Observe(float64(time.Since(startTime)))
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

func (i *eventInjector) InjectProv(pg *graph.Provenance) error {
	buf, err := proto.Marshal(pg)
	if err != nil {
		return err
	}
	i.Publish(i.prefix, buf)
	return nil
}
