package storage

import (
	"strings"

	"github.com/cayleygraph/cayley/quad"
	"github.com/satori/go.uuid"
)

// Key is the type to identify elements in the graph
type Key struct {
	UUID uuid.UUID `json:"uuid"`
}

// NewRandomKey returns a random Key based on random numbers (RFC 4122)
func NewRandomKey() Key {
	return Key{
		UUID: uuid.NewV4(),
	}
}

// NewKeyFromQuadIRI returns a Key from a Quad Value
func NewKeyFromQuadIRI(v quad.IRI) (Key, error) {
	uuid, err := uuid.FromString(removeIDBrackets(quad.StringOf(v)))
	if err != nil {
		return Key{}, err
	}
	return Key{
		UUID: uuid,
	}, nil
}

// NewKeyFromQuadValue returns a Key from a Quad Value
func NewKeyFromQuadValue(v quad.Value) (Key, error) {
	uuid, err := uuid.FromString(removeIDBrackets(quad.StringOf(v)))
	if err != nil {
		return Key{}, err
	}
	return Key{
		UUID: uuid,
	}, nil
}

// NewKeyFromString returns a Key from a Quad Value
func NewKeyFromString(v string) (Key, error) {
	uuid, err := uuid.FromString(v)
	if err != nil {
		return Key{}, err
	}
	return Key{
		UUID: uuid,
	}, nil
}

// QuadIRI returns key in Quad IRI format
func (k Key) QuadIRI() quad.IRI {
	return quad.IRI(k.UUID.String()).Short()
}

// QuadValue return key in Quad Value format. The <, > makes the type quad.IRI instead of quad.String
func (k Key) QuadValue() quad.Value {
	return quad.StringToValue("<" + k.UUID.String() + ">")
}

// String returns key in string format
func (k Key) String() string {
	return k.UUID.String()
}

// Equals returns whether or not the Keys are equal
func (k Key) Equals(k2 Key) bool {
	return uuid.Equal(k.UUID, k2.UUID)
}

func removeIDBrackets(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		switch r {
		case '>':
			return true
		case '<':
			return true
		default:
			return false
		}
	})
}
