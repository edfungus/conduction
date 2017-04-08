package main

import "github.com/edfungus/conduction/pb"

type Storage interface {
	PathListen(path pb.Path) (bool, error)                          // Check if path is being listened to
	GetFlow(path pb.Path) (map[string]pb.Flow, error)               // Gets flows from a path. Could be multiple therefore the uuid is important to identify
	AddFlow(flow pb.Flow) (string, error)                           // Adds a flow. Will traverse through dependent flows and add if not yet made
	UpdateFlowFull(uuid string) error                               // Updates everythig including all subsequent dependent flows
	UpdateFlowProperties(uuid string, flow pb.Flow) error           // Updates everything except dependent flows (doesn't even look at it)
	UpdateFlowDepedent(uuid string, dependentFlows []pb.Flow) error // Updates only the dependent flow order/relationship (doesn't look inside dependent flows)
	DeleteFlow(uuid string) error                                   // Deletes the flow and removes from all relationships. Path may stay only if another flow uses it
}
