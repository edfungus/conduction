package distributor

import "github.com/edfungus/conduction/model"

//Distributor is used for internal or inter-broker communication
type Distributor interface {
	Send(msg *model.Message)
	Messages() chan *DistributorMessage
	Acknowledge(msg *DistributorMessage)
	Errors() chan error
	Close()
}

// DistributorMessage is the message coming from a distributor
type DistributorMessage struct {
	Message   *model.Message
	topic     string
	partition int32
	offset    int64
}
