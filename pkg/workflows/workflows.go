package workflows

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/gradecak/fission-workflows/pkg/types"
	"github.com/gradecak/fission-workflows/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
)

const (
	TASK_NAME = "T%d"
)

type WorkflowConfig struct {
	//WF Specific Config
	Consent    bool
	Provenance bool

	//Task Specifica
	ProvMeta             map[string]interface{}
	PercentMultienvTasks float32
	TaskRuntime          string

	RandomTaskName bool
}

func NewWorkflow(parallel, serial int, wfConf *WorkflowConfig) *types.WorkflowSpec {
	tasks := []map[string]*types.TaskSpec{}
	outputTask := fmt.Sprintf(TASK_NAME, parallel*serial)
	numTasks := parallel * serial

	numMZTasks := 0
	if wfConf != nil && wfConf.PercentMultienvTasks != 0.0 {
		numMZTasks = int(float32(numTasks) * wfConf.PercentMultienvTasks)
	}

	t := 1
	for i := 0; i < serial; i++ {
		tasks = append(tasks, map[string]*types.TaskSpec{})
		// check how many tasks in the workflow need to be multienv
		// create the tasks
		for j := 0; j < parallel; j++ {
			var taskName string
			if wfConf.RandomTaskName {
				taskName = uuid.New().String()
				outputTask = taskName
			} else {
				taskName = fmt.Sprintf(TASK_NAME, (i*parallel + (j + 1)))
			}

			if i > 0 {
				tasks[i][taskName] = newTask(tasks[i-1], taskName)
			} else {
				tasks[i][taskName] = newTask(map[string]*types.TaskSpec{}, taskName)
			}

			//check if the task should be multizone we use this
			// technique such that the multizone tasks will be
			// spaced out evenly throughout the task spec
			if numMZTasks != 0 &&
				(i*parallel+(j+1))%int(numTasks/numMZTasks) == 0 &&
				t <= numMZTasks {
				t++
				tasks[i][taskName].ExecConstraints = &types.TaskDataflowSpec{
					MultiZone: true,
				}
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

	if wfConf.TaskRuntime != "" {
		addTaskRuntime(taskSpecs, wfConf.TaskRuntime)
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

func newTask(deps map[string]*types.TaskSpec, fnName string) *types.TaskSpec {

	d := map[string]*types.TaskDependencyParameters{}
	for dep, _ := range deps {
		d[dep] = &types.TaskDependencyParameters{}
	}

	ts := &types.TaskSpec{
		FunctionRef: fnName,
		Requires:    d,
		Await:       int32(len(d)),
	}

	return ts
}

func addTaskRuntime(tasks map[string]*types.TaskSpec, runtime string) {
	time, err := typedvalues.Wrap(runtime)
	if err != nil {
		logrus.Error(err)
	}
	if rt, err := typedvalues.Wrap(map[string]*typedvalues.TypedValue{"runtime": time}); err == nil {
		for _, task := range tasks {
			task.Inputs = map[string]*typedvalues.TypedValue{
				"query": rt,
			}
		}
	} else {
		logrus.Error(err)
	}

}
