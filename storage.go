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
	GetFlow(id string) (*pb.Flow, error)                                     // Gets a flow based on the id
	SaveFlowFull(flow *pb.Flow) (string, error)                              // Adds/update a flow. Will traverse through dependent flows and add if not yet made and replace existing contents will given
	SaveFlowSingle(flow *pb.Flow, dependentFlowIDs []string) (string, error) // Adds/updates only the flow passed in. Will not traverse. Dependent flows provide by IDs not Flow
	// DeleteFlow(id string) error                                              // Deletes the flow and removes from all relationships. Path may stay only if another flow uses it
}

type CockroachStorage struct {
	url string
	DB  *sql.DB
}

const (
	SCHEMA_FILE   string = "./database/schema.sql"
	DATABASE_USER string = "conductor"
	// DATABASE_NAME string = "conduction"
	DATABASE_URL string = "postgresql://%s@%s/%s?sslmode=disable"
)

// NewCockroachStorage returns a new cockroach storage object
func NewCockroachStorage(url string, databaseName string) (*CockroachStorage, error) {
	initdb, err := sql.Open("postgres", fmt.Sprintf(DATABASE_URL, "root", url, databaseName))
	if err != nil {
		return nil, err
	}
	_, err = initdb.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", databaseName))
	if err != nil {
		return nil, err
	}
	if err := initializeDatabase(initdb); err != nil {
		return nil, err
	}
	_, err = initdb.Exec(fmt.Sprintf("GRANT ALL ON %s.* TO %s", databaseName, DATABASE_USER))
	if err != nil {
		return nil, err
	}
	initdb.Close()

	db, err := sql.Open("postgres", fmt.Sprintf(DATABASE_URL, DATABASE_USER, url, databaseName))
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(fmt.Sprintf("SET DATABASE = %s", databaseName))
	if err != nil {
		return nil, err
	}
	cs := &CockroachStorage{
		url: url,
		DB:  db,
	}
	return cs, nil
}

// Close ends cockroadh db connection
func (cs *CockroachStorage) Close() error {
	return cs.DB.Close()
}

// for this flow, check all dependents, call AddFlow (aka recursive) to get the flow ids to add
// if this flow, see if id exists and check if it exist
// // if it exists call UpdateFlowFull (then return flow id)
// // if id does not exist, then it must be new, so let make it
func (cs *CockroachStorage) SaveFlowFull(flow *pb.Flow) (int64, error) {
	if flow.Path == nil {
		return 0, fmt.Errorf("Flow does not have a path associated with it: %v", flow)
	}

	// Process dependent Flows first. Make them if necessary
	lengthOfDependents := len(flow.DependentFlows)
	dependentFlowIDs := make([]int64, lengthOfDependents)
	if flow.DependentFlows != nil && lengthOfDependents > 0 {
		for i := 0; i < lengthOfDependents; i++ {
			flowID, err := cs.SaveFlowFull(flow.DependentFlows[i])
			if err != nil {
				return 0, fmt.Errorf("Error adding flow: %v", err)
			}
			dependentFlowIDs[i] = flowID
		}
	}

	flowID, err := cs.SaveFlowSingle(flow, dependentFlowIDs)
	if err != nil {
		return 0, err
	}

	return flowID, nil
}

// SaveFlowSingle updates or inserts a Flow without traversing into dependents. A list of Flow ids `dependentFlowIDs` is used instead of the `Flow.dependentFlows` so that Flow objects do not need ot be made.
func (cs *CockroachStorage) SaveFlowSingle(flow *pb.Flow, dependentFlowIDs []int64) (int64, error) {
	// If Flow has an existing id, get old Flow. A flow with no id (default zero value) then the flow needs to be created
	var oldFlow *pb.Flow
	if flow.Id != 0 {
		ok, err := cs.FlowIDExist(flow.Id)
		switch {
		case err != nil:
			return 0, err
		case !ok:
			return 0, fmt.Errorf("Flow id %d was not found in database", flow.Id)
		}
		oldFlow, err = cs.GetFlow(flow.Id)
	}

	// Ensure that all Flow dependents exists
	for i := 0; i < len(dependentFlowIDs); i++ {
		ok, err := cs.FlowIDExist(dependentFlowIDs[i])
		switch {
		case err != nil:
			return 0, err
		case !ok:
			return 0, fmt.Errorf("Flow dependent id %d was not found in database", dependentFlowIDs[i])
		}
	}

	// Ensure new Path is added to database
	_, err := cs.SavePath(flow.Path)
	if err != nil {
		return 0, err
	}

	// Update/insert??

	// Clean up old stuff if possible
	if oldFlow != nil && oldFlow.Path != flow.Path {
		// check to see if we need to delete old Path... DeletePath() should only delete if nothing relies on it... so maybe call it anyways?
	}

	// *** need cases where if this is an update... check if old and new path changed.. also do this for dependents!

	// Check if all dependent flows ids exist (if not nil) .. if not error
	// Check if path exists, if not make it
	// Update/insert new properties
	// If dependent flow array is not nil, update/insert new dependent relationships
	// Return newly crete flow's id
	return 0, fmt.Errorf("Function not done...")
}

// GetFlow gets a Flow by id. If flow does not exist, a nil value will be returned
func (cs *CockroachStorage) GetFlow(id int64) (*pb.Flow, error) {
	return nil, fmt.Errorf("Function not done...")
}

// SavePath saves a Path to the database and return the id. If the path is already in the database, it will just return the id
func (cs *CockroachStorage) SavePath(path *pb.Path) (int64, error) {
	var pathID int64
	// See if path already exist
	err := cs.DB.QueryRow("SELECT id FROM paths WHERE route=$1 AND type=$2", path.Route, path.Type).Scan(&pathID)
	if err == sql.ErrNoRows {
		// Insert new Path
		err := cs.DB.QueryRow("INSERT INTO paths(route, type) VALUES($1, $2) RETURNING id", path.Route, path.Type).Scan(&pathID)
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	return pathID, nil
}

// FlowIDExist returns whether or not Flow id exists in database
func (cs *CockroachStorage) FlowIDExist(id int64) (bool, error) {
	var flowID int64
	err := cs.DB.QueryRow("SELECT id FROM flows WHERE id = $1", id).Scan(&flowID)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
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
