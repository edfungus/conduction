package storage

import (
	"github.com/edfungus/conduction/messenger"
	"github.com/satori/go.uuid"
)

type Flows struct {
	Flows []Flow `json:"flows"`
}

type Flow struct {
	UUID        uuid.UUID       `json:"uuid,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Path        *messenger.Path `json:"path"`
}

// AddKeyToFlows adds/replace keys in the optional UUID field
func AddKeyToFlows(flows []Flow, keys []Key) Flows {
	combineList := Flows{Flows: []Flow{}}
	for i := range flows {
		flows[i].UUID = keys[i].UUID
		combineList.Flows = append(combineList.Flows, flows[i])
	}
	return combineList
}
