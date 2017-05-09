package storage

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"os"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/bolt"
	_ "github.com/cayleygraph/cayley/graph/sql" // Need for Cayley to connect to database
	"github.com/cayleygraph/cayley/quad"
	"github.com/cayleygraph/cayley/schema"
	"github.com/edfungus/conduction/pb"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

type Storage interface {
	AddFlow(flow *pb.Flow) (uuid.UUID, error)
	ReadFlow(uuid uuid.UUID) (*pb.Flow, error)
	AddPath(path *pb.Path) (uuid.UUID, error)
	ReadPath(uuid uuid.UUID) (*pb.Path, error)
	AddFlowToPath(pathUUID uuid.UUID, flowUUID uuid.UUID) error
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
	tmpfile, err := ioutil.TempFile("", "example")
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
func (gs *GraphStorage) AddFlow(flow *pb.Flow) (uuid.UUID, error) {
	pathUUID, err := gs.AddPath(flow.Path)
	flowUUID := uuid.NewV4()
	flowDTO := flowDTO{
		ID:          uuidToQuadIRI(flowUUID),
		Name:        flow.Name,
		Description: flow.Description,
		Path:        uuidToQuadIRI(pathUUID),
	}
	err = gs.writeToGraph(flowDTO)
	if err != nil {
		return uuid.Nil, err
	}
	return flowUUID, nil
}

// ReadFlow returns a Flow of the sepcified uuid from the graph
func (gs *GraphStorage) ReadFlow(uuid uuid.UUID) (*pb.Flow, error) {
	var flowDTO flowDTO
	err := schema.LoadTo(nil, gs.store, &flowDTO, uuidToQuadIRI(uuid))
	if err != nil {
		return nil, err
	}
	pathUUID, err := quadIRIToUUID(flowDTO.Path)
	if err != nil {
		return nil, err
	}
	path, err := gs.ReadPath(pathUUID)
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
func (gs *GraphStorage) AddPath(path *pb.Path) (uuid.UUID, error) {
	// Find if Path already exist. If so, return the Path's id
	p := cayley.StartPath(gs.store, quad.StringToValue(path.Type)).In(quad.StringToValue("<type>")).Has(quad.StringToValue("<route>"), quad.StringToValue(path.Route))
	pathList, err := p.Iterate(nil).Limit(1).AllValues(gs.store)
	if err != nil {
		return uuid.Nil, err
	}
	if len(pathList) == 1 {
		pathUUID, err := quadValueToUUID(pathList[0])
		if err != nil {
			return uuid.Nil, err
		}
		return pathUUID, nil
	}
	// Insert new Path
	pathID := uuid.NewV4()
	pathDTO := pathDTO{
		ID:    uuidToQuadIRI(pathID),
		Route: path.Route,
		Type:  path.Type,
	}
	err = gs.writeToGraph(pathDTO)
	if err != nil {
		return uuid.Nil, err
	}
	return pathID, nil
}

// ReadPath returns Path based on uuid.
func (gs *GraphStorage) ReadPath(uuid uuid.UUID) (*pb.Path, error) {
	var pathDTO pathDTO
	err := schema.LoadTo(nil, gs.store, &pathDTO, uuidToQuadIRI(uuid))
	if err != nil {
		return nil, err
	}
	return &pb.Path{
		Route: pathDTO.Route,
		Type:  pathDTO.Type,
	}, nil
}

func uuidToQuadIRI(uuid uuid.UUID) quad.IRI {
	return quad.IRI(uuid.String()).Short()
}

func quadValueToUUID(quadValue quad.Value) (uuid.UUID, error) {
	return uuid.FromString(removeIDBrackets(quadValue.String()))
}

func quadIRIToUUID(quadIRI quad.IRI) (uuid.UUID, error) {
	return uuid.FromString(removeIDBrackets(quadIRI.String()))
}

func removeIDBrackets(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
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

func (gs *GraphStorage) writeToGraph(dto interface{}) error {
	writer := graph.NewWriter(gs.store)
	_, err := schema.WriteAsQuads(writer, dto)
	if err != nil {
		return err
	}
	writer.Close()
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
