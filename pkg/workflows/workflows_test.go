package workflows

import (
	"fmt"
	"testing"
)

func TestWorkflow(t *testing.T) {
	wf := NewWorkflow(2, 2)
	fmt.Printf("%+v", wf)
}
