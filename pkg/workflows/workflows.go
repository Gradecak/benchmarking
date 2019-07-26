package workflows

import (
	"fmt"
	"github.com/gradecak/fission-workflows/pkg/types"
)

const (
	TASK_NAME = "T%d"
)

type WorkflowConfig struct {
	//WF Specific Config
	Consent    bool
	Provenance bool

	//Task Specific
	ProvMeta      map[string]interface{}
	MultienvTasks bool
	//TODO task zone locking/hinting
}

func NewWorkflow(parallel, serial int, wfConf *WorkflowConfig) *types.WorkflowSpec {
	tasks := []map[string]*types.TaskSpec{}
	outputTask := fmt.Sprintf(TASK_NAME, parallel*serial)

	for i := 0; i < serial; i++ {
		tasks = append(tasks, map[string]*types.TaskSpec{})
		for j := 0; j < parallel; j++ {
			taskName := fmt.Sprintf(TASK_NAME, (i*parallel + (j + 1)))
			if i > 0 {
				tasks[i][taskName] = newTask(tasks[i-1], wfConf)
			} else {
				tasks[i][taskName] = newTask(map[string]*types.TaskSpec{}, wfConf)
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

	spec := &types.WorkflowSpec{
		OutputTask: outputTask,
		Tasks:      taskSpecs,
	}

	if wfConf != nil {
		dfSpec := &types.DataFlowSpec{
			ConsentCheck: wfConf.Consent,
			Provenance:   wfConf.Provenance,
		}
		spec.Dataflow = dfSpec
	}

	return spec
}

func newTask(deps map[string]*types.TaskSpec, wfConf *WorkflowConfig) *types.TaskSpec {

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
