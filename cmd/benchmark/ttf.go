package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gradecak/benchmark/pkg/workflows"
	"github.com/gradecak/fission-workflows/pkg/types"
	"github.com/sirupsen/logrus"
)

var Zones = []string{}

type TTFExperiment struct {
	url      string
	sfURL    string
	proxyURL string
	expLabel string

	//wf spec config
	numTasks  int
	runsPerWf int
}

type TTFResult struct {
	Timestamp         time.Time
	ExpLabel          string
	TimeToFailure     time.Duration
	RunsBeforeFailure int
	PercentDFTasks    float32
	NZones            int
}

func setupTTFExp(cnf *ExperimentConf) (Experiment, error) {
	exp := &TTFExperiment{
		url: cnf.Url,
		//TODO replace with url from config file
		// sfURL:     "http://127.0.0.1:9000",
		// numTasks:  10,
		// runsPerWf: 100,
		expLabel: cnf.ExpLabel,
	}

	// parse simfission url
	sfUrl, ok := cnf.ExpParams["simfaasUrl"].(string)
	if !ok {
		return nil, errors.New("simfission url not found in config")
	}
	exp.sfURL = sfUrl

	//parse proxy url
	proxyURL, ok := cnf.ExpParams["proxyUrl"].(string)
	if !ok {
		return nil, errors.New("proxy url not found in config")
	}
	exp.proxyURL = proxyURL

	//parse numTasks
	nTasks, ok := cnf.ExpParams["numTasks"].(int)
	if !ok {
		return nil, errors.New("num tasks not found in config")
	}
	exp.numTasks = nTasks

	//parse zones
	zones, ok := cnf.ExpParams["zones"].([]interface{})
	if !ok {
		return nil, errors.New("cannot find zones in config")
	} else {
		for _, z := range zones {
			if zone, ok := z.(string); ok {
				Zones = append(Zones, zone)
			} else {
				logrus.Info("Pee pee poo poo")
			}
		}
	}
	logrus.Infof("Parsed Zones %v", Zones)

	//parse runsPerWf
	runs, ok := cnf.ExpParams["runsPerWf"].(int)
	if !ok {
		return nil, errors.New("cannpt find number of runs in config")
	}
	exp.runsPerWf = runs

	return exp, nil
}

func (exp TTFExperiment) Run(c Context) (interface{}, error) {
	output := []TTFResult{}
	// create and start the simfission proxy
	ctx, cancel := context.WithCancel(c)
	proxy := newProxy(exp.sfURL)
	go proxy.run(":9999", ctx)
	time.Sleep(time.Second)

	//experiment
	client, err := NewFWClient(exp.url)
	if err != nil {
		panic(err)
	}
	for i := 1; i < exp.numTasks+1; i += 1 {
		percentDFTasks := float32(i) / float32(exp.numTasks)
		wfSpec := workflows.NewWorkflow(1, exp.numTasks, &workflows.WorkflowConfig{
			PercentMultienvTasks: percentDFTasks,
			TaskRuntime:          "0.2",
		})
		wfId, err := client.SetupWfFromSpec(Context{}, wfSpec)
		if err != nil {
			panic(err)
		}

		for j := 0; j < exp.runsPerWf; j++ {
			res, err := proxy.invokeWf(wfId, wfSpec, client)
			if err != nil {
				logrus.Fatal(err)
				return nil, err
			}
			res.PercentDFTasks = percentDFTasks
			output = append(output, *res)
		}
	}

	//shut down proxy
	cancel()
	return output, nil
}

// ************************************************************ //
//               Simple proxy for simfass requests              //
// ************************************************************ //

type simfaasProxy struct {
	simfissionURL       string
	rand                *rand.Rand
	stat                *proxyStat
	previousConstraints map[string]map[string]string
}

type proxyStat struct {
	wfId            string
	failed          bool
	taskZones       map[string]string //map tasks to zones
	succInvocations int
	start           time.Time
}

func newProxy(sfUrl string) *simfaasProxy {
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	return &simfaasProxy{sfUrl, rand, nil, make(map[string]map[string]string)}
}

func (p *simfaasProxy) invokeWf(wfId string, wfSpec *types.WorkflowSpec, client *FWClient) (*TTFResult, error) {
	if p.stat == nil {
		logrus.Info("Starting new round")
		p.stat = &proxyStat{wfId, false, nil, 0, time.Now()}
		if constraints, ok := p.previousConstraints[wfId]; ok {
			p.stat.taskZones = constraints
		} else {
			constraints = p.genMzConstraints(wfSpec)
			logrus.Infof("Generated Constraits %v", constraints)
			p.previousConstraints[wfId] = constraints
			p.stat.taskZones = constraints
		}
	} else {
		logrus.Warn("stat not cleared from previous experiment run")
	}

	//keep going until we fail
	for !p.stat.failed {
		_, err := client.Invoke(Context{}, wfId)
		if !p.stat.failed {
			p.stat.succInvocations++
		}
		if err != nil {
			return nil, err
		}
	}

	result := &TTFResult{
		Timestamp:         time.Now(),
		TimeToFailure:     time.Since(p.stat.start),
		RunsBeforeFailure: p.stat.succInvocations,
		NZones:            len(Zones),
	}
	// clear stat before next run
	p.stat = nil
	return result, nil
}

func (p *simfaasProxy) genMzConstraints(wfSpec *types.WorkflowSpec) map[string]string {
	ret := make(map[string]string)
	tIds := wfSpec.TaskIds()
	for _, tId := range tIds {
		taskSpec := wfSpec.TaskSpec(tId)
		// if task is multizone give it a zone constraint
		if taskSpec.GetExecConstraints().GetMultiZone() {
			ret[taskSpec.GetFunctionRef()] = Zones[p.rand.Intn(len(Zones))]
		}
	}
	return ret
}

func (p *simfaasProxy) run(url string, ctx context.Context) {
	router := mux.NewRouter()
	server := http.Server{Addr: url, Handler: router}

	router.PathPrefix("/").Handler(p.Forward())
	go func() {
		logrus.Info("Starting Proxy...")
		<-ctx.Done()
		if err := server.Shutdown(ctx); err != nil {
			logrus.Fatal(err)
		}
		logrus.Info("Shutting Down Proxy...")
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logrus.Fatal(err)
	}
	logrus.Info("Finished with proxy")
}

func (p *simfaasProxy) Forward() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := r.URL
		// only do fun stuff when Fission Workflows is attempting to run
		// the task, everything else we just forward to simfaas
		if strings.Contains(url.String(), "fission-function") && !p.stat.failed {
			//get the fn name
			fnName := url.Path[strings.LastIndex(url.Path, "/")+1:]
			// if task has zone
			if i := strings.LastIndex(fnName, "-"); i != -1 {
				baseName := fnName[:i]
				if zone, ok := p.stat.taskZones[baseName]; ok {
					if fnName[i+1:] != zone {
						logrus.Info("Failed")
						p.stat.failed = true
					}
				}
			} else {
				// if task has no zone, check if it should
				if _, ok := p.stat.taskZones[fnName]; ok {
					logrus.Info("Failed")
					p.stat.failed = true
				}
			}

		}
		resp, err := http.Get(fmt.Sprintf("%s/%s", p.simfissionURL, url))
		if err != nil {
			logrus.Error("Failed to contact simfaas (reason %v)", err.Error())
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logrus.Error(err)
		}
		w.Write(body)
	}
}
