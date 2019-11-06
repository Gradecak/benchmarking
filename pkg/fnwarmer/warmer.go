package fnwarmer

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gradecak/fission-workflows/pkg/types"
)

// "github.com/gradecak/fission-workflows/pkg/types"
// "github.com/gradecak/fission-workflows/pkg/types/typedvalues"

type Warmer struct {
	client *http.Client
	url    string
}

func New(url string) *Warmer {
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	return &Warmer{
		client: &http.Client{},
		url:    url,
	}
}

var zones = []string{
	"nl",
	"de",
	"fr",
	"ir",
	"au",
}

func (w *Warmer) WarmupTasks(spec *types.WorkflowSpec) error {
	for _, taskSpec := range spec.Tasks {
		var fnRef string
		if taskSpec.GetExecConstraints().GetMultiZone() {
			fnRef = addMzSuffix(taskSpec.GetFunctionRef())
		} else {
			fnRef = taskSpec.GetFunctionRef()
		}
		err := w.warmupFn(fnRef)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Warmer) warmupFn(fnRef string) error {
	resp, err := w.client.Get(fmt.Sprintf("%s/fission-function/%s", w.url, fnRef))
	resp.Body.Close()
	return err
}

func addMzSuffix(fnRef string) string {
	zone := zones[rand.Intn(len(zones))]
	return fmt.Sprintf("%s-%s", fnRef, zone)
}
