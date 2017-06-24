package router

import "github.com/edfungus/conduction/messenger"

type Flow struct {
	Name        string
	Description string
	Path        *messenger.Path
}
