package fnwarmer

import (
	"testing"

	"github.com/gradecak/benchmark/pkg/workflows"
)

func TestWarmer(t *testing.T) {
	wfSpec := workflows.NewWorkflow(1, 3, &workflows.WorkflowConfig{
		TaskRuntime:          "1",
		RandomTaskName:       true,
		PercentMultienvTasks: 1,
	})
	warmer := New("http://127.0.0.1:9000")
	warmer.WarmupTasks(wfSpec)

}
