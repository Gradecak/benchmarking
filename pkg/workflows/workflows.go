package workflows

import (
	"fmt"
	"github.com/gradecak/fission-workflows/pkg/types"
)

const (
	TASK_NAME = "T%d"
)

func NewWorkflow(parallel, serial int) *types.WorkflowSpec {
	tasks := []map[string]*types.TaskSpec{}
	outputTask := fmt.Sprintf(TASK_NAME, parallel*serial)

	for i := 0; i < serial; i++ {
		tasks = append(tasks, map[string]*types.TaskSpec{})
		for j := 0; j < parallel; j++ {
			taskName := fmt.Sprintf(TASK_NAME, (i*parallel + (j + 1)))
			if i > 0 {
				tasks[i][taskName] = newTask(tasks[i-1])
			} else {
				tasks[i][taskName] = newTask(map[string]*types.TaskSpec{})
			}
		}
	}

	//flatten task
	taskSpecs := map[string]*types.TaskSpec{}
	for _, t := range tasks {
		for k, v := range t {
			taskSpecs[k] = v
		}
	}

	return &types.WorkflowSpec{
		// ApiVersion: "",
		OutputTask: outputTask,
		Tasks:      taskSpecs,
	}
}

func newTask(deps map[string]*types.TaskSpec) *types.TaskSpec {

	d := map[string]*types.TaskDependencyParameters{}
	for dep, _ := range deps {
		d[dep] = &types.TaskDependencyParameters{}
	}

	ts := &types.TaskSpec{
		FunctionRef: "taskFn",
		Requires:    d,
		Await:       int32(len(d)),
	}
	return ts
}
