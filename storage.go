package main

import (
	"database/sql"
	"fmt"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/sql"
	"github.com/cayleygraph/cayley/quad"
	"github.com/cayleygraph/cayley/schema"
	"github.com/edfungus/conduction/pb"
	"github.com/pborman/uuid"
)

type Storage interface {
	AddFlow(flow *pb.Flow) (uuid.UUID, error)
	ReadFlow(uuid uuid.UUID) (*pb.Flow, error)
	SavePath(path *pb.Path) error
}

type GraphStorage struct {
	store *cayley.Handle
	qw    graph.BatchWriter
}

type flowDTO struct {
	ID          quad.IRI `quad:"@id"`
	Name        string   `quad:"name"`
	Description string   `quad:"description"`
	Path        *pathDTO `quad:"path"`
}

type pathDTO struct {
	ID    quad.IRI `quad:"@id"`
	Route string   `quad:"route"`
	Type  string   `quad:"type"`
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
	graph.InitQuadStore("sql", databasePath, opts)
	store, err := cayley.NewGraph("sql", databasePath, opts)
	if err != nil {
		return nil, err
	}
	return &GraphStorage{
		store: store,
		qw:    graph.NewWriter(store),
	}, nil
}

// AddFlow adds a new Flow to the graph. If the Path does not exist, it will be added, else it will be made
func (gs *GraphStorage) AddFlow(flow *pb.Flow) (string, error) {
	pathUUID := fmt.Sprintf("path:%s", uuid.NewRandom().String())
	flowUUID := fmt.Sprintf("flow:%s", uuid.NewRandom().String())
	flowDTO := flowDTO{
		ID:          toQuadIRI(flowUUID),
		Name:        flow.Name,
		Description: flow.Description,
		Path: &pathDTO{
			ID:    toQuadIRI(pathUUID),
			Route: flow.Path.Route,
			Type:  flow.Path.Type,
		},
	}
	_, err := schema.WriteAsQuads(gs.qw, flowDTO)
	if err != nil {
		return "", err
	}
	return flowUUID, nil
}

// ReadFlow returns a Flow of the sepcified uuid from the graph
func (gs *GraphStorage) ReadFlow(uuid string) (*pb.Flow, error) {
	var flowDTO flowDTO
	err := schema.LoadTo(nil, gs.store, &flowDTO, toQuadIRI(uuid))
	if err != nil {
		return nil, err
	}
	return &pb.Flow{
		Name:        flowDTO.Name,
		Description: flowDTO.Description,
		Path: &pb.Path{
			Route: flowDTO.Path.Route,
			Type:  flowDTO.Path.Type,
		},
	}, nil
}

// func (gs *GraphStorage) SavePath(path pb.Path) error {
// }

func toQuadIRI(uuid string) quad.IRI {
	return quad.IRI(uuid).Full().Short()
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
