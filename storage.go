package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"

	_ "github.com/lib/pq"

	"github.com/edfungus/conduction/pb"
)

type Storage interface {
	PathListen(path pb.Path) (bool, error)                             // Check if path is being listened to
	FindFlow(path pb.Path) (map[string]pb.Flow, error)                 // Gets flows from a path. Could be multiple therefore the uuid is important to identify
	GetFlow(uuid string) (pb.Flow, error)                              // Gets a flow based on the id
	AddFlow(flow pb.Flow) (string, error)                              // Adds a flow. Will traverse through dependent flows and add if not yet made
	UpdateFlowFull(uuid string) error                                  // Updates everythig including all subsequent dependent flows
	UpdateFlowProperties(uuid string, flow pb.Flow) error              // Updates everything except dependent flows (doesn't even look at it)
	UpdateFlowDepedent(uuid string, dependentFlowUuids []string) error // Updates only the dependent flow order/relationship (doesn't look inside dependent flows)
	DeleteFlow(uuid string) error                                      // Deletes the flow and removes from all relationships. Path may stay only if another flow uses it
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
