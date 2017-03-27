package connector

type Connector interface {
	Requests() <-chan Request
	Respond(path string, payload []byte) error
	AddPath(path Path) error
	RemovePath(path Path) error
	Close() error
}

// Request encapsulates the origin and contents of an incoming request
type Request struct {
	Path    string
	Payload []byte
}

type Path struct {
	Method   string
	PathName string
}
