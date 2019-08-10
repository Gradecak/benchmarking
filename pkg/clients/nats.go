package clients

import (
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	stan "github.com/nats-io/stan.go"
)

type NATS struct {
	stan.Conn
}

type NATSConf struct {
	Url     string
	Cluster string
}

func NewNatsClient(cnf *NATSConf) (*NATS, error) {
	uuid, err := uuid.NewRandom()
	conn, err := stan.Connect(cnf.Cluster, uuid.String(), stan.NatsURL(cnf.Url))
	return &NATS{conn}, err
}

func (n *NATS) PublishProto(prefix string, msg proto.Message) error {
	buf, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return n.Publish(prefix, buf)
}
