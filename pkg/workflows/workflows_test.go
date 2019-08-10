package workflows

import (
	// "fmt"
	"github.com/gradecak/fission-workflows/pkg/types"
	"testing"
)

// func TestWorkflow(t *testing.T) {
// 	wf := NewWorkflow(2, 2, &WorkflowConfig{})
// 	fmt.Printf("%+v", wf)
// }

func TestPercentMultizone(t *testing.T) {
	for i := 1; i < 11; i++ {
		wf := NewWorkflow(1, 10, &WorkflowConfig{
			PercentMultienvTasks: float32(i) / float32(10),
		})
		n := numMz(wf)
		expected := int(10 * (float32(i) / float32(10)))
		if n != expected {
			t.Errorf("incorrect number of me tasks expected %v was %v", expected, n)
		}
	}

}

func numMz(wfSpec *types.WorkflowSpec) int {
	ret := 0
	tIds := wfSpec.TaskIds()
	for _, tId := range tIds {
		taskSpec := wfSpec.TaskSpec(tId)
		// if task is multizone give it a zone constraint
		if taskSpec.GetExecConstraints().GetMultiZone() {
			ret++
		}
	}
	return ret
}
