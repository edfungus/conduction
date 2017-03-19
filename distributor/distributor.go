package distributor

//Distributor is used for internal or inter-broker communication
type Distributor interface {
	Send(msg *Message)
	Messages() chan *DistributorMessage
	Acknowledge(msg *DistributorMessage)
	Errors() chan error
	Close()
}

// DistributorMessage is the message coming from a distributor
type DistributorMessage struct {
	Message   *Message
	topic     string
	partition int32
	offset    int64
}
