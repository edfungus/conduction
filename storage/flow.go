package storage

import "github.com/edfungus/conduction/messenger"

type Flow struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Path        *messenger.Path `json:"path"`
}
