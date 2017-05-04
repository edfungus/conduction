package storage

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/sql" // Need for Cayley to connect o database
	"github.com/cayleygraph/cayley/quad"
	"github.com/cayleygraph/cayley/schema"
	"github.com/edfungus/conduction/pb"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
)

type Storage interface {
	AddFlow(flow *pb.Flow) (uuid.UUID, error)
	ReadFlow(uuid uuid.UUID) (*pb.Flow, error)
	SavePath(path *pb.Path) error
}

const (
	DATABASE_URL string = "postgresql://%s@%s:%d/%s?sslmode=disable"
)

// Logger logs but can be replaced
var Logger = logrus.New()

type GraphStorage struct {
	store *cayley.Handle
	qw    graph.BatchWriter
}

type GraphStorageConfig struct {
	Host         string
	Port         int
	User         string
	DatabaseName string
	DatabaseType string
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

// NewGraphStorage returns a new Storage that uses Cayley and CockroachDB
func NewGraphStorage(config *GraphStorageConfig) (*GraphStorage, error) {
	databasePath := fmt.Sprintf(DATABASE_URL, config.User, config.Host, config.Port, config.DatabaseName)
	opts := graph.Options{"flavor": config.DatabaseType}
	initDatabase(databasePath, config.DatabaseName, opts)
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
	pathUUID, err := gs.SavePath(flow.Path)
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
	_, err = schema.WriteAsQuads(gs.qw, flowDTO)
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

// SavePath adds path to graph if new, else it will return the id of the existing path. Path are unique based on route and type combined
func (gs *GraphStorage) SavePath(path *pb.Path) (string, error) {
	// Find if Path already exist. If so, return the Path's id
	p := cayley.StartPath(gs.store, quad.StringToValue(path.Type)).In(quad.StringToValue("<type>")).Has(quad.StringToValue("<route>"), quad.StringToValue(path.Route))
	pathList, err := p.Iterate(nil).Limit(1).AllValues(gs.store)
	if err != nil {
		return "", err
	}
	if len(pathList) == 1 {
		return quadToID(pathList[0]), nil
	}

	// Insert new Path
	pathID := fmt.Sprintf("path:%s", uuid.NewRandom().String())
	pathDTO := pathDTO{
		ID:    toQuadIRI(pathID),
		Route: path.Route,
		Type:  path.Type,
	}
	_, err = schema.WriteAsQuads(gs.qw, pathDTO)
	if err != nil {
		return "", err
	}
	return pathID, nil
}

// readPath returns Path based on uuid. Used only internally
func (gs *GraphStorage) readPath(uuid string) (*pb.Path, error) {
	var pathDTO pathDTO
	err := schema.LoadTo(nil, gs.store, &pathDTO, toQuadIRI(uuid))
	if err != nil {
		return nil, err
	}
	return &pb.Path{
		Route: pathDTO.Route,
		Type:  pathDTO.Type,
	}, nil
}

func toQuadIRI(uuid string) quad.IRI {
	return quad.IRI(uuid).Full().Short()
}

func quadToID(quadValue quad.Value) string {
	return strings.TrimFunc(quadValue.String(), func(r rune) bool {
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

func initDatabase(connectionPath string, databaseName string, opts graph.Options) error {
	initdb, err := sql.Open("postgres", connectionPath)
	if err != nil {
		return err
	}

	// Check if the database needs to be set up
	_, err = initdb.Exec(fmt.Sprintf("SET DATABASE = %s", databaseName))
	if err != nil {
		_, err = initdb.Exec(fmt.Sprintf("CREATE DATABASE %s", databaseName))
		if err != nil {
			return err
		}
	}
	_, err = initdb.Exec(fmt.Sprintf("SET DATABASE = %s; SELECT * FROM quads", databaseName))
	if err != nil {
		err := graph.InitQuadStore("sql", connectionPath, opts)
		if err != nil {
			return err
		}
	}

	return nil
}
