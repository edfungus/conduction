package main

import (
	"database/sql"
	"fmt"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/sql"
	"github.com/edfungus/conduction/pb"
)

type Storage interface {
	AddFlow(flow pb.Flow) (int64, error)
	SavePath(path pb.Path) error
}

type GraphStorage struct {
	store *cayley.Handle
}

const (
	DATABASE_URL string = "postgresql://%s@%s:%d/%s?sslmode=disable"
)

type GraphStorageConfig struct {
	Host         string
	Port         int
	User         string
	DatabaseName string
	DatabaseType string
}

// NewGraphStorage returns a new Storage that uses Cayley and CockroachDB
func NewGraphStorage(config *GraphStorageConfig) (*GraphStorage, error) {
	databasePath := fmt.Sprintf(DATABASE_URL, config.User, config.Host, config.Port, config.DatabaseName)
	initDatabase(databasePath, config.DatabaseName)

	opts := graph.Options{"flavor": config.DatabaseType}
	err := graph.InitQuadStore("sql", databasePath, opts)
	store, err := cayley.NewGraph("sql", databasePath, opts)
	if err != nil {
		return nil, err
	}
	return &GraphStorage{
		store: store,
	}, nil
}

func initDatabase(connectionPath string, databaseName string) error {
	initdb, err := sql.Open("postgres", connectionPath)
	if err != nil {
		return err
	}
	_, err = initdb.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", databaseName))
	if err != nil {
		return err
	}
	return nil
}
