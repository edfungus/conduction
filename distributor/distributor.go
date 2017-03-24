package distributor

import "github.com/edfungus/conduction/model"

//Distributor is used for internal or inter-broker communication
type Distributor interface {
	Send(msg *model.Message) error
	Messages() chan *ReceivedMessage
	Acknowledge(msg *ReceivedMessage)
	Errors() chan error
	Close()
}

// ReceivedMessage is the message coming from a distributor
type ReceivedMessage struct {
	Message   *model.Message
	Topic     string
	Partition int32
	Offset    int64
}
