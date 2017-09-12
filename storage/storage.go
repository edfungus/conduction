package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/bolt"
	_ "github.com/cayleygraph/cayley/graph/sql"
	"github.com/cayleygraph/cayley/quad" // Need for Cayley to connect to database
	"github.com/cayleygraph/cayley/schema"
	"github.com/edfungus/conduction/messenger"
	"github.com/sirupsen/logrus"
)

// Logger logs but can be replaced
var Logger = logrus.New()

type Storage interface {
	SaveFlow(flow Flow) (Key, error)
	GetFlowByKey(key Key) (Flow, error)

	SavePath(path messenger.Path) (Key, error)
	GetPathByKey(key Key) (messenger.Path, error)
	GetKeyOfPath(path messenger.Path) (Key, error)

	ChainNextFlowToPath(flowKey Key, pathKey Key) error
	GetNextFlows(key Key) ([]Flow, []Key, error)
}

const (
	databaseURL string = "postgresql://%s@%s:%d/%s?sslmode=disable"
)

var (
	ErrFlowCannotBeRetrieved error = fmt.Errorf("Could not retrieve Flow from storage")
	ErrPathCannotBeRetrieved error = fmt.Errorf("Could not retrieve Path from storage")
	ErrResolvingKey          error = fmt.Errorf("Error resolving key in database")
)

type flowDTO struct {
	ID          quad.IRI `quad:"@id"`
	Name        string   `quad:"name"`
	Description string   `quad:"description"`
	Path        quad.IRI `quad:"path"`
}

// NewFlowDTO returns a new flowDTO
func NewFlowDTO(id quad.IRI, name string, description string, path quad.IRI) flowDTO {
	return flowDTO{
		ID:          id,
		Name:        name,
		Description: description,
		Path:        path,
	}
}

type pathDTO struct {
	ID    quad.IRI   `quad:"@id"`
	Route string     `quad:"route"`
	Type  string     `quad:"type"`
	Flows []quad.IRI `quad:"triggers,optional"`
}

func NewPathDTO(id quad.IRI, route string, pathType string, flows []quad.IRI) pathDTO {
	return pathDTO{
		ID:    id,
		Route: route,
		Type:  pathType,
		Flows: flows,
	}
}

type GraphStorageConfig struct {
	Host         string
	Port         int
	User         string
	DatabaseName string
	DatabaseType string
}

// GetDatabaseConnectionPath returns a constructed URL from the database details in config
func (gsc GraphStorageConfig) GetDatabaseConnectionPath() string {
	return fmt.Sprintf(databaseURL, gsc.User, gsc.Host, gsc.Port, gsc.DatabaseName)
}

type GraphStorage struct {
	store *cayley.Handle
}

// NewGraphStorage returns a new Storage that uses Cayley and CockroachDB
func NewGraphStorage(config GraphStorageConfig) (*GraphStorage, error) {
	err := setupDatabaseForGraphStore(config)
	if err != nil {
		return nil, err
	}
	store, err := cayley.NewGraph("sql", config.GetDatabaseConnectionPath(), graph.Options{"flavor": config.DatabaseType})
	if err != nil {
		return nil, err
	}
	return &GraphStorage{
		store: store,
	}, nil
}

// NewGraphStorageBolt allows graph to be stored in a file instead sql above. This is to dodge the issue with inserting a struct of optional or empty array in to the graph (postgres issue: https://github.com/cayleygraph/cayley/issues/563)
func NewGraphStorageBolt(filepath string) (*GraphStorage, error) {
	tmpfile, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	graph.InitQuadStore("bolt", tmpfile.Name(), nil)
	store, err := cayley.NewGraph("bolt", tmpfile.Name(), nil)
	if err != nil {
		return nil, err
	}
	return &GraphStorage{
		store: store,
	}, nil
}

// SaveFlow adds a new Flow to the graph. If the Path does not exist, it will be added, else it will be made
func (gs *GraphStorage) SaveFlow(flow Flow) (Key, error) {
	pathKey, err := gs.SavePath(*flow.Path)
	if err != nil {
		return Key{}, err
	}
	flowKey := NewRandomKey()
	flowDTO := NewFlowDTO(flowKey.QuadIRI(), flow.Name, flow.Description, pathKey.QuadIRI())
	err = gs.writeToGraph(flowDTO)
	if err != nil {
		return Key{}, err
	}
	return flowKey, nil
}

// GetFlowByKey returns a Flow of the sepcified uuid from the graph
func (gs *GraphStorage) GetFlowByKey(key Key) (Flow, error) {
	var flowDTO flowDTO
	err := schema.LoadTo(nil, gs.store, &flowDTO, key.QuadValue())
	if err != nil {
		return Flow{}, ErrFlowCannotBeRetrieved
	}
	pathKey, err := NewKeyFromQuadIRI(flowDTO.Path)
	if err != nil {
		return Flow{}, err
	}
	path, err := gs.GetPathByKey(pathKey)
	if err != nil {
		return Flow{}, err
	}
	return Flow{
		Name:        flowDTO.Name,
		Description: flowDTO.Description,
		Path: &messenger.Path{
			Route: path.Route,
			Type:  path.Type,
		},
	}, nil
}

// SavePath adds path to graph if new, else it will return the id of the existing path. Path are unique based on route and type combined
func (gs *GraphStorage) SavePath(path messenger.Path) (Key, error) {
	pathExists, err := gs.doesPathExists(path)
	if err != nil {
		return Key{}, err
	}
	if pathExists {
		return gs.GetKeyOfPath(path)
	}
	pathKey := NewRandomKey()
	pathDTO := NewPathDTO(pathKey.QuadIRI(), path.Route, path.Type, nil)
	err = gs.writeToGraph(pathDTO)
	if err != nil {
		return Key{}, err
	}
	return pathKey, nil
}

// GetKeyOfPath returns the Key of a given Path if it exists
func (gs *GraphStorage) GetKeyOfPath(path messenger.Path) (Key, error) {
	pathDTOList, err := gs.getPathDTOsByRouteAndType(path.Route, path.Type)
	if err != nil {
		return Key{}, err
	}
	switch length := len(pathDTOList); length {
	case 0:
		return Key{}, errors.New("Path was not found in graph store")
	case 1:
		pathKey, err := NewKeyFromQuadIRI(pathDTOList[0].ID)
		if err != nil {
			return Key{}, err
		}
		return pathKey, nil
	default:
		return Key{}, errors.New("There are multiple matching Paths. They are expected to be unique")
	}
}

// GetPathByKey returns Path based on uuid.
func (gs *GraphStorage) GetPathByKey(key Key) (messenger.Path, error) {
	var pathDTO pathDTO
	err := schema.LoadTo(nil, gs.store, &pathDTO, key.QuadValue())
	if err != nil {
		return messenger.Path{}, ErrPathCannotBeRetrieved
	}
	return messenger.Path{
		Route: pathDTO.Route,
		Type:  pathDTO.Type,
	}, nil
}

// ChainNextFlowToPath connects Flows to be triggered by a Path
func (gs *GraphStorage) ChainNextFlowToPath(flowKey Key, pathKey Key) error {
	_, err := gs.GetFlowByKey(flowKey)
	if err != nil {
		return err
	}
	_, err = gs.GetPathByKey(pathKey)
	if err != nil {
		return err
	}
	err = gs.linkKeyToTriggerKey(pathKey, flowKey)
	if err != nil {
		return err
	}
	return nil
}

// GetNextFlows returns a list of Flows that are triggers by the Flow
func (gs *GraphStorage) GetNextFlows(key Key) ([]Flow, []Key, error) {
	flowKeyList, err := gs.getKeysTriggeredByKey(key)
	if err != nil {
		return nil, nil, err
	}
	var flowList []Flow
	for _, v := range flowKeyList {
		flow, err := gs.GetFlowByKey(v)
		if err != nil {
			return nil, nil, err
		}
		flowList = append(flowList, flow)
	}
	return flowList, flowKeyList, nil
}

func (gs *GraphStorage) doesPathExists(path messenger.Path) (bool, error) {
	pathDTOList, err := gs.getPathDTOsByRouteAndType(path.Route, path.Type)
	if err != nil {
		return false, nil
	}
	if len(pathDTOList) > 0 {
		return true, nil
	}
	return false, nil
}

func (gs *GraphStorage) getPathDTOsByRouteAndType(pathRoute string, pathType string) ([]pathDTO, error) {
	p := cayley.StartPath(gs.store, quad.StringToValue(pathType)).In(quad.IRI("type")).Has(quad.IRI("route"), quad.StringToValue(pathRoute))
	var pathDTOs []pathDTO
	err := schema.LoadIteratorTo(nil, gs.store, reflect.ValueOf(&pathDTOs), p.BuildIterator())
	return pathDTOs, err
}

func (gs *GraphStorage) linkKeyToTriggerKey(originKey Key, destinationKey Key) error {
	return gs.store.AddQuad(quad.Make(originKey.QuadValue(), quad.IRI("triggers"), destinationKey.QuadValue(), nil))
}

func (gs *GraphStorage) getKeysTriggeredByKey(key Key) ([]Key, error) {
	p := cayley.StartPath(gs.store, key.QuadValue()).Out(quad.IRI("triggers"))
	keyQuadList, err := p.Iterate(nil).AllValues(gs.store)
	if err != nil {
		return nil, ErrResolvingKey
	}
	keys, err := convertQuadValueListToKeyList(keyQuadList)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func convertQuadValueListToKeyList(quads []quad.Value) ([]Key, error) {
	var keys []Key
	for _, v := range quads {
		key, err := NewKeyFromQuadValue(v)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
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

// Close ends graph database session. Currently, it will delete temporary database file if used
func (gs *GraphStorage) Close() {
	gs.store.Close()
}

func setupDatabaseForGraphStore(config GraphStorageConfig) error {
	err := createSQLDatabaseIfNotExist(config)
	if err != nil {
		return err
	}
	err = initGraphStoreIfNotInitialized(config)
	if err != nil {
		return err
	}
	return nil
}

func createSQLDatabaseIfNotExist(config GraphStorageConfig) error {
	databaseExist, err := doesDatabaseExist(config)
	if err != nil {
		return err
	}
	if !databaseExist {
		err := createDatabase(config)
		if err != nil {
			return err
		}
	}
	return nil
}

func initGraphStoreIfNotInitialized(config GraphStorageConfig) error {
	databaseInitialized, err := isDatabaseInitializedForGraphStore(config)
	if err != nil {
		return err
	}
	if !databaseInitialized {
		err := graph.InitQuadStore("sql", config.GetDatabaseConnectionPath(), graph.Options{"flavor": config.DatabaseType})
		if err != nil {
			return err
		}
	}
	return nil
}

func doesDatabaseExist(config GraphStorageConfig) (bool, error) {
	db, err := connectToSQLDatabase(config)
	if err != nil {
		return false, err
	}
	defer db.Close()
	_, err = db.Exec(fmt.Sprintf("SET DATABASE = %s", config.DatabaseName))
	if err != nil {
		return false, nil
	}
	return true, nil
}

func createDatabase(config GraphStorageConfig) error {
	db, err := connectToSQLDatabase(config)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", config.DatabaseName))
	return err
}

func isDatabaseInitializedForGraphStore(config GraphStorageConfig) (bool, error) {
	db, err := connectToSQLDatabase(config)
	if err != nil {
		return false, err
	}
	defer db.Close()
	_, err = db.Exec(fmt.Sprintf("SET DATABASE = %s; SELECT * FROM quads", config.DatabaseName))
	if err != nil {
		return false, nil
	}
	return true, nil
}

func connectToSQLDatabase(config GraphStorageConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.GetDatabaseConnectionPath())
	if err != nil {
		return nil, err
	}
	return db, nil
}
