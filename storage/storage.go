package storage

type Storage interface {
	Put(key string, value *Destination) error
	Get(key string) (*Destination, error)
	Delete(key string) error
	Close()
}

// Destination is the information we wish to store
// This in the future may be part of connectors instead
type Destination struct {
	Method string
	Path   string
}
