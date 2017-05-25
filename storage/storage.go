package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"os"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/bolt"
	_ "github.com/cayleygraph/cayley/graph/sql" // Need for Cayley to connect to database
	"github.com/cayleygraph/cayley/quad"
	"github.com/cayleygraph/cayley/schema"
	"github.com/edfungus/conduction/pb"
	"github.com/sirupsen/logrus"
)

type Storage interface {
	AddFlow(flow *pb.Flow) (Key, error)
	ReadFlow(key Key) (*pb.Flow, error)

	AddPath(path *pb.Path) (Key, error)
	ReadPath(key Key) (*pb.Path, error)

	AddFlowToPath(pathKey Key, flowKey Key) error
	GetFlowsForPath(pathKey Key) ([]pb.Flow, error)
}

const (
	DATABASE_URL string = "postgresql://%s@%s:%d/%s?sslmode=disable"
)

// Logger logs but can be replaced
var Logger = logrus.New()

type GraphStorage struct {
	store   *cayley.Handle
	tmpFile *os.File
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
	Path        quad.IRI `quad:"path"`
}

type pathDTO struct {
	ID    quad.IRI   `quad:"@id"`
	Route string     `quad:"route"`
	Type  string     `quad:"type"`
	Flow  []quad.IRI `quad:"triggers,optional"`
}

// NewGraphStorage returns a new Storage that uses Cayley and CockroachDB
func NewGraphStorage(config *GraphStorageConfig) (*GraphStorage, error) {
	return NewGraphStorageBolt() // overriding sql for now
	databasePath := fmt.Sprintf(DATABASE_URL, config.User, config.Host, config.Port, config.DatabaseName)
	opts := graph.Options{"flavor": config.DatabaseType}
	initDatabase(databasePath, config.DatabaseName, opts)
	store, err := cayley.NewGraph("sql", databasePath, opts)
	if err != nil {
		return nil, err
	}
	return &GraphStorage{
		store: store,
	}, nil
}

// NewGraphStorageBolt allows graph to be stored in a file instead sql above. This is to dodge the issue with inserting a struct of optional or empty array in to the graph (postgres issue: https://github.com/cayleygraph/cayley/issues/563)
func NewGraphStorageBolt() (*GraphStorage, error) {
	tmpfile, err := ioutil.TempFile("./", "example")
	if err != nil {
		log.Fatal(err)
	}
	graph.InitQuadStore("bolt", tmpfile.Name(), nil)
	store, err := cayley.NewGraph("bolt", tmpfile.Name(), nil)
	if err != nil {
		return nil, err
	}
	return &GraphStorage{
		store:   store,
		tmpFile: tmpfile,
	}, nil
}

// Close ends graph database session. Currently, it will delete temporary database file if used
func (gs *GraphStorage) Close() {
	if gs.tmpFile != nil {
		os.Remove(gs.tmpFile.Name())
	}
}

// AddFlow adds a new Flow to the graph. If the Path does not exist, it will be added, else it will be made
func (gs *GraphStorage) AddFlow(flow *pb.Flow) (Key, error) {
	pathKey, err := gs.AddPath(flow.Path)
	if err != nil {
		return Key{}, err
	}
	flowKey := NewRandomKey()
	flowDTO := flowDTO{
		ID:          flowKey.QuadIRI(),
		Name:        flow.Name,
		Description: flow.Description,
		Path:        pathKey.QuadIRI(),
	}
	err = gs.writeToGraph(flowDTO)
	if err != nil {
		return Key{}, err
	}
	return flowKey, nil
}

// ReadFlow returns a Flow of the sepcified uuid from the graph
func (gs *GraphStorage) ReadFlow(key Key) (*pb.Flow, error) {
	var flowDTO flowDTO
	err := schema.LoadTo(nil, gs.store, &flowDTO, key.QuadValue())
	if err != nil {
		return nil, err
	}
	pathKey, err := NewKeyFromQuadIRI(flowDTO.Path)
	if err != nil {
		return nil, err
	}
	path, err := gs.ReadPath(pathKey)
	if err != nil {
		return nil, err
	}
	return &pb.Flow{
		Name:        flowDTO.Name,
		Description: flowDTO.Description,
		Path: &pb.Path{
			Route: path.Route,
			Type:  path.Type,
		},
	}, nil
}

// AddPath adds path to graph if new, else it will return the id of the existing path. Path are unique based on route and type combined
func (gs *GraphStorage) AddPath(path *pb.Path) (Key, error) {
	// Find if Path already exist. If so, return the Path's id
	p := cayley.StartPath(gs.store, quad.StringToValue(path.Type)).In(quad.IRI("type")).Has(quad.IRI("route"), quad.StringToValue(path.Route))
	pathList, err := p.Iterate(nil).Limit(1).AllValues(gs.store)
	if err != nil {
		return Key{}, err
	}
	if len(pathList) == 1 {
		pathKey, err := NewKeyFromQuadValue(pathList[0])
		if err != nil {
			return Key{}, err
		}
		return pathKey, nil
	}
	// Insert new Path
	pathKey := NewRandomKey()
	pathDTO := pathDTO{
		ID:    pathKey.QuadIRI(),
		Route: path.Route,
		Type:  path.Type,
	}
	err = gs.writeToGraph(pathDTO)
	if err != nil {
		return Key{}, err
	}
	return pathKey, nil
}

// ReadPath returns Path based on uuid.
func (gs *GraphStorage) ReadPath(key Key) (*pb.Path, error) {
	var pathDTO pathDTO
	err := schema.LoadTo(nil, gs.store, &pathDTO, key.QuadValue())
	if err != nil {
		return nil, err
	}
	return &pb.Path{
		Route: pathDTO.Route,
		Type:  pathDTO.Type,
	}, nil
}

// AddFlowToPath connects Flows to be triggered by a Path
func (gs *GraphStorage) AddFlowToPath(pathKey Key, flowKey Key) error {
	pathCheck, err := gs.uuidExists(pathKey)
	switch {
	case err != nil:
		return err
	case pathCheck == false:
		return errors.New("Path Key does no exist in graph")
	}
	flowCheck, err := gs.uuidExists(flowKey)
	switch {
	case err != nil:
		return err
	case flowCheck == false:
		return errors.New("Flow Key does no exist in graph")
	}
	err = gs.store.AddQuad(quad.Make(pathKey.QuadValue(), quad.IRI("triggers"), flowKey.QuadValue(), nil))
	if err != nil {
		return err
	}
	return nil
}

// GetFlowsForPath returns a list of Flows that are triggers by the Flow
func (gs *GraphStorage) GetFlowsForPath(pathKey Key) ([]pb.Flow, error) {
	p := cayley.StartPath(gs.store, pathKey.QuadValue()).Out(quad.IRI("triggers"))
	flowQValues, err := p.Iterate(nil).AllValues(gs.store)
	if err != nil {
		return nil, err
	}
	flowList := []pb.Flow{}
	for _, v := range flowQValues {
		key, err := NewKeyFromQuadValue(v)
		if err != nil {
			return nil, err
		}
		flow, err := gs.ReadFlow(key)
		if err != nil {
			return nil, err
		}
		flowList = append(flowList, *flow)
	}
	return flowList, nil
}

func (gs *GraphStorage) uuidExists(key Key) (bool, error) {
	v, err := cayley.StartPath(gs.store, key.QuadValue()).Iterate(nil).AllValues(gs.store)
	switch {
	case err != nil:
		return false, err
	case len(v) >= 1:
		return true, nil
	default:
		return false, nil
	}
}

func (gs *GraphStorage) writeToGraph(dto interface{}) error {
	writer := graph.NewWriter(gs.store)
	_, err := schema.WriteAsQuads(writer, dto)
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}
	return nil
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
