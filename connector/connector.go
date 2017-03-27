package connector

type Connector interface {
	Requests() <-chan Request
	Respond(path string, payload []byte) error
	Close() error
}

// Request encapsulates the origin and contents of an incoming request
type Request struct {
	Path    string
	Payload []byte
}
