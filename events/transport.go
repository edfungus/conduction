package events

//Transport is used for internal or inter-broker communication.. bad name maybe rename disperse?
type Transport interface {
	Send(msg *Message)
	Messages() chan *Message
	Errors() chan error
	Close()
}
