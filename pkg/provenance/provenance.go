package provenance

import (
	"github.com/google/uuid"
	"github.com/gradecak/fission-workflows/pkg/provenance/graph"
	"math/rand"
	"time"
)

type Cnf struct {
	ID          string
	NTasks      int
	Predecessor int64
	TaskNodes   *graph.Node
}

type Gen struct {
	rand *rand.Rand
}

func NewGenerator() *Gen {
	return &Gen{rand.New(rand.NewSource(time.Now().Unix()))}
}

func (g *Gen) NewRandomProv(nTasks int) *graph.Provenance {
	return g.NewProv(&Cnf{
		ID:     uuid.New().String(),
		NTasks: nTasks,
	})
}

func (g *Gen) NewProv(cnf *Cnf) *graph.Provenance {
	provenance := &graph.Provenance{
		Nodes:          make(map[int64]*graph.Node),
		WfTasks:        make(map[int64]*graph.IDs),
		WfPredecessors: make(map[int64]*graph.IDs),
		Executed:       make(map[string]int64),
	}
	wfID := g.rand.Int63()

	// add workflow tasks
	wfTasks := &graph.IDs{[]int64{}}
	for i := 0; i < cnf.NTasks; i++ {
		var (
			id   int64
			node *graph.Node
		)
		if cnf.TaskNodes != nil {
			id = g.rand.Int63()
			node = cnf.TaskNodes
		} else {
			id, node = g.genTaskNode()
		}

		provenance.Nodes[id] = node
		wfTasks.Ids = append(wfTasks.Ids, id)
	}
	provenance.WfTasks[wfID] = wfTasks
	provenance.Executed[cnf.ID] = wfID
	return provenance
}

func (g *Gen) genTaskNode() (int64, *graph.Node) {
	nodeID := g.rand.Int63()
	node := &graph.Node{}

	node.Type = graph.Node_TASK
	node.Op = graph.Node_READ
	node.Task = "asdf"
	return nodeID, node
}
