package benchmark

import (
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type ResultInterpreter interface {
	Parse(chan Result) // list of result
	Save(string)       // save to csv file
}

type Experiment struct {
	Fn          string
	Interpreter ResultInterpreter
}

type Config struct {
	Runs    int    // number of runs to perform per client
	Clients int    // number of clients to concurrently execute function
	URL     string // Fission router url
	exp     Experiment
}

type Result struct {
	Response string
	Time     time.Duration
}

func (conf *Config) SetExp(e Experiment) {
	conf.exp = e
}

func (conf Config) RunExperiment() error {
	var wg sync.WaitGroup
	wg.Add(conf.Clients)

	results := make(chan Result, conf.Clients*conf.Runs)

	logrus.Info("Starting clients")
	for i := 0; i < conf.Clients; i++ {
		go func(url string, fn string, runs int, rc chan Result) {
			defer wg.Done()

			//intialise http client
			cl := NewFissionClient(url)

			for j := 0; j < runs; j++ {
				// make the query
				start := time.Now()
				result, err := cl.execFn(fn)
				if err != nil {
					logrus.Errorf("Error executing target function for client: %v reason: %v", i, err)
				}
				taken := time.Now().Sub(start)
				rc <- Result{result, taken}
			}

		}(conf.URL, conf.exp.Fn, conf.Runs, results)
	}

	// Close channel when all tasks have resolved
	go func() {
		wg.Wait()
		close(results)
	}()

	// collect results and feed them to the experiment specific result
	// interpreter
	logrus.Info("All clients finished, processing results")
	conf.exp.Interpreter.Parse(results)
	//conf.exp.Interpreter.Save("test.csv")
	return nil
}
