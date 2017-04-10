package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"

	_ "github.com/lib/pq"

	"github.com/edfungus/conduction/pb"
)

type Storage interface {
	// PathListen(path pb.Path) (map[string]pb.Flow, error)                     // Check if path is being listened to and return associated flows
	// FindFlow(path pb.Path) (map[string]pb.Flow, error)                       // Gets flows from a path. Could be multiple therefore the id is important to identify
	// GetFlow(id string) (pb.Flow, error)                                      // Gets a flow based on the id
	SaveFlowFull(flow *pb.Flow) (string, error)                              // Adds/update a flow. Will traverse through dependent flows and add if not yet made and replace existing contents will given
	SaveFlowSingle(flow *pb.Flow, dependentFlowIDs []string) (string, error) // Adds/updates only the flow passed in. Will not traverse. Dependent flows provide by IDs not Flow
	// DeleteFlow(id string) error                                              // Deletes the flow and removes from all relationships. Path may stay only if another flow uses it
}

type CockroachStorage struct {
	url string
	db  *sql.DB
}

const (
	SCHEMA_FILE   string = "./database/schema.sql"
	DATABASE_USER string = "conductor"
	DATABASE_NAME string = "conduction"
	DATABASE_URL  string = "postgresql://%s@%s/%s?sslmode=disable"
)

func NewCockroachStorage(url string) (*CockroachStorage, error) {
	initdb, err := sql.Open("postgres", fmt.Sprintf(DATABASE_URL, "root", url, DATABASE_NAME))
	if err != nil {
		return nil, err
	}
	if err := initializeDatabase(initdb); err != nil {
		return nil, err
	}
	initdb.Close()

	db, err := sql.Open("postgres", fmt.Sprintf(DATABASE_URL, DATABASE_USER, url, DATABASE_NAME))
	if err != nil {
		return nil, err
	}
	cs := &CockroachStorage{
		url: url,
		db:  db,
	}
	return cs, nil
}

// for this flow, check all dependents, call AddFlow (aka recursive) to get the flow ids to add
// if this flow, see if id exists and check if it exist
// // if it exists call UpdateFlowFull (then return flow id)
// // if id does not exist, then it must be new, so let make it
func (cs *CockroachStorage) SaveFlowFull(flow *pb.Flow) (string, error) {
	if flow.Path == nil {
		return "", fmt.Errorf("Flow does not have a path associated with it: %v", flow)
	}

	// Process dependent Flows first. Make them if necessary
	lengthOfDependents := len(flow.DependentFlows)
	dependentFlowIDs := make([]string, lengthOfDependents)
	if flow.DependentFlows != nil && lengthOfDependents > 0 {
		for i := 0; i < lengthOfDependents; i++ {
			flowID, err := cs.SaveFlowFull(flow.DependentFlows[i])
			if err != nil {
				return "", fmt.Errorf("Error adding flow: %v", err)
			}
			dependentFlowIDs[i] = flowID
		}
	}

	flowID, err := cs.SaveFlowSingle(flow, dependentFlowIDs)
	if err != nil {
		return "", err
	}

	return flowID, nil
}

func (cs *CockroachStorage) SaveFlowSingle(flow *pb.Flow, dependentFlowIDs []string) (string, error) {
	// Check if all dependent flows ids exist (if not nil) .. if not error
	// Check if path exists, if not make it
	// Update/insert new properties
	// If dependent flow array is not nil, update/insert new dependent relationships
	// Return newly crete flow's id
	return "", fmt.Errorf("Function not done...")
}

// pathExist returns whether or not Path exists in database
func (cs *CockroachStorage) pathExist(path pb.Path) (bool, error) {
	return false, fmt.Errorf("Function not done...")

}

func initializeDatabase(db *sql.DB) error {
	query, err := ioutil.ReadFile(SCHEMA_FILE)
	if err != nil {
		return err
	}
	if _, err := db.Exec(string(query)); err != nil {
		return err
	}
	return nil
}
