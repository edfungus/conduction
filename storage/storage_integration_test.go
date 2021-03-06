// +build all integration

package storage

import (
	"os"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/quad"
	"github.com/edfungus/conduction/messenger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
)

var _ = Describe("Conduction", func() {
	Describe("Storage", func() {
		var (
			graph *GraphStorage
			// config = GraphStorageConfig{
			// 	Host:         databaseHost,
			// 	Port:         databasePort,
			// 	User:         "root",
			// 	DatabaseName: databaseName,
			// 	DatabaseType: "cockroach",
			// }
		)
		BeforeEach(func() {
			var err error
			os.Create(tempFilePath)
			graph, err = NewGraphStorageBolt(tempFilePath)
			Expect(err).To(BeNil())
			Expect(graph).ToNot(BeNil())
		})
		AfterEach(func() {
			graph.Close()
			os.Remove(tempFilePath)
		})
		Describe("Given creating a new Storage", func() {
			Context("When Cockroach path is wrong or not available", func() {
				It("Then should return an error", func() {
					// TODO: When cayley sql is fixed, turn me back on!
					// badConfig := &GraphStorageConfig{
					// 	Host:         config.Host,
					// 	Port:         8888,
					// 	User:         config.User,
					// 	DatabaseName: config.DatabaseName,
					// 	DatabaseType: config.DatabaseType,
					// }
					// _, err := NewGraphStorage(badConfig)
					// Expect(err).ToNot(BeNil())
				})
			})
			Context("When Cockroach path is correct and available", func() {
				It("Then GraphStorage should return without error", func() {
					tempFilePathForTest := "test_db_test.tmp"
					os.Create(tempFilePathForTest)
					graph, err := NewGraphStorageBolt(tempFilePathForTest)
					Expect(err).To(BeNil())
					Expect(graph).ToNot(BeNil())
					os.Remove(tempFilePathForTest)
				})
			})
		})
		Describe("Given saving a Flow to graph", func() {
			Context("When the Path does not exist", func() {
				It("Then the Path and Flow should also be inserted connected to Path", func() {
					flow := Flow{
						Name:        "Flow Name",
						Description: "Flow Description",
						Path: &messenger.Path{
							Route: "/test2",
							Type:  "mqtt-duplicate",
						},
					}
					key, err := graph.SaveFlow(flow)
					Expect(err).To(BeNil())
					Expect(key).ToNot(Equal(Key{}))

					newFlow, err := graph.GetFlowByKey(key)
					Expect(err).To(BeNil())
					Expect(newFlow.Name).To(Equal(flow.Name))
					Expect(newFlow.Description).To(Equal(flow.Description))
					Expect(newFlow.Path.Route).To(Equal(flow.Path.Route))
					Expect(newFlow.Path.Type).To(Equal(flow.Path.Type))
				})
			})
			Context("When the Path exists", func() {
				It("Then the Flow should be inserted connected to existing Path", func() {
					// Save Path
					pathRoute := "/test"
					pathType := "mqtt-duplicate"
					path := messenger.Path{
						Route: pathRoute,
						Type:  pathType,
					}
					pathKey, err := graph.SavePath(path)
					Expect(err).To(BeNil())
					Expect(pathKey).ToNot(Equal(Key{}))

					// Save flow with same path route and type
					flow := Flow{
						Name:        "Flow Name",
						Description: "Flow Description",
						Path: &messenger.Path{
							Route: path.Route,
							Type:  path.Type,
						},
					}
					flowKey, err := graph.SaveFlow(flow)
					Expect(err).To(BeNil())
					Expect(flowKey).ToNot(Equal(Key{}))

					// Make sure Path was not recreated
					p := cayley.StartPath(graph.store, quad.StringToValue(pathType)).In(quad.StringToValue("<type>")).Has(quad.StringToValue("<route>"), quad.StringToValue(pathRoute))
					pathList, err := p.Iterate(nil).AllValues(graph.store)
					Expect(err).To(BeNil())
					Expect(len(pathList)).To(Equal(1))

					// Check that Flow was correctly saved
					newFlow, err := graph.GetFlowByKey(flowKey)
					Expect(err).To(BeNil())
					Expect(newFlow.Name).To(Equal(flow.Name))
					Expect(newFlow.Description).To(Equal(flow.Description))
					Expect(newFlow.Path.Route).To(Equal(flow.Path.Route))
					Expect(newFlow.Path.Type).To(Equal(flow.Path.Type))
				})
			})
		})
		Describe("Given saving a Path to graph", func() {
			Context("When the Path does already exist", func() {
				It("Then the Path id should be returned", func() {
					// Path to save
					pathRoute := "/unique/path"
					pathType := "mqtt-duplicate"
					path := messenger.Path{
						Route: pathRoute,
						Type:  pathType,
					}

					// Insert Path twice
					pathKey1, err := graph.SavePath(path)
					Expect(err).To(BeNil())
					pathKey2, err := graph.SavePath(path)
					Expect(err).To(BeNil())
					Expect(pathKey1).To(Equal(pathKey2))

					// Check Path was added only once
					p := cayley.StartPath(graph.store, quad.StringToValue(pathType)).In(quad.StringToValue("<type>")).Has(quad.StringToValue("<route>"), quad.StringToValue(pathRoute))
					pathList, err := p.Iterate(nil).AllValues(graph.store)
					Expect(err).To(BeNil())
					Expect(len(pathList)).To(Equal(1))

					// Check Path content
					GetPathByKey, err := graph.GetPathByKey(pathKey1)
					Expect(err).To(BeNil())
					Expect(GetPathByKey.Route).To(Equal(path.Route))
					Expect(GetPathByKey.Type).To(Equal(path.Type))
				})
			})
			Context("When the Path does not already exist", func() {
				It("Then a new Path should be inserted and id returned", func() {
					// Path to save
					pathRoute := "/unique/path"
					pathType := "mqtt-unique"
					path := messenger.Path{
						Route: pathRoute,
						Type:  pathType,
					}

					// Ensure path doesn't exist
					p := cayley.StartPath(graph.store, quad.StringToValue(pathType)).In(quad.StringToValue("<type>")).Has(quad.StringToValue("<route>"), quad.StringToValue(pathRoute))
					pathList, err := p.Iterate(nil).AllValues(graph.store)
					Expect(err).To(BeNil())
					Expect(len(pathList)).To(Equal(0))

					// Insert Path
					pathKey, err := graph.SavePath(path)

					// Check Path was added
					p = cayley.StartPath(graph.store, quad.StringToValue(pathType)).In(quad.StringToValue("<type>")).Has(quad.StringToValue("<route>"), quad.StringToValue(pathRoute))
					pathList, err = p.Iterate(nil).AllValues(graph.store)
					Expect(err).To(BeNil())
					Expect(len(pathList)).To(Equal(1))

					// Check Path content
					GetPathByKey, err := graph.GetPathByKey(pathKey)
					Expect(err).To(BeNil())
					Expect(GetPathByKey.Route).To(Equal(path.Route))
					Expect(GetPathByKey.Type).To(Equal(path.Type))
				})
			})
		})
		Describe("Given a Flow id and Path id (which will trigger the Flow)", func() {
			Context("When the ids are given are valid", func() {
				It("Then the Path will be connected to the Flow", func() {
					// Save Path
					pathTriggerRoute := "/test"
					pathTriggerType := "path-trigger"
					pathTrigger := messenger.Path{
						Route: pathTriggerRoute,
						Type:  pathTriggerType,
					}
					pathTriggerKey, err := graph.SavePath(pathTrigger)
					Expect(err).To(BeNil())
					Expect(pathTriggerKey).ToNot(Equal(Key{}))

					// Save Flow
					flow := Flow{
						Name:        "Flow Name",
						Description: "Flow Description",
						Path: &messenger.Path{
							Route: "/some-route",
							Type:  "mqtt",
						},
					}
					flowKey, err := graph.SaveFlow(flow)
					Expect(err).To(BeNil())
					Expect(flowKey).ToNot(Equal(Key{}))

					// Connect Flow to Path
					err = graph.ChainNextFlowToPath(flowKey, pathTriggerKey)
					Expect(err).To(BeNil())

					// Check that Path connects to the Flow with vertex "triggers"
					p := cayley.StartPath(graph.store, pathTriggerKey.QuadValue()).Out(quad.IRI("triggers"))
					triggersList, err := p.Iterate(nil).AllValues(graph.store)
					Expect(err).To(BeNil())
					Expect(len(triggersList)).To(Equal(1))
					Expect(NewKeyFromQuadValue(triggersList[0])).To(Equal(flowKey))
				})
			})
			Context("When the Path already has a Flow that it triggers", func() {
				It("Then should add the new Flow so the Path triggers two Flows", func() {
					// Save Path
					pathTriggerRoute := "/test"
					pathTriggerType := "path-trigger"
					pathTrigger := messenger.Path{
						Route: pathTriggerRoute,
						Type:  pathTriggerType,
					}
					pathTriggerKey, err := graph.SavePath(pathTrigger)
					Expect(err).To(BeNil())
					Expect(pathTriggerKey).ToNot(Equal(uuid.Nil))

					// Save both Flows
					flow := Flow{
						Name:        "Flow Name",
						Description: "Flow Description",
						Path: &messenger.Path{
							Route: "/some-route",
							Type:  "mqtt",
						},
					}
					flowKey1, err := graph.SaveFlow(flow)
					Expect(err).To(BeNil())
					Expect(flowKey1).ToNot(Equal(Key{}))
					flowKey2, err := graph.SaveFlow(flow)
					Expect(err).To(BeNil())
					Expect(flowKey2).ToNot(Equal(Key{}))

					// Connect Flow to Path
					err = graph.ChainNextFlowToPath(flowKey1, pathTriggerKey)
					Expect(err).To(BeNil())
					err = graph.ChainNextFlowToPath(flowKey2, pathTriggerKey)
					Expect(err).To(BeNil())

					// Check that Path connects to both Flow 1 and Flow 2
					p := cayley.StartPath(graph.store, pathTriggerKey.QuadValue()).Out(quad.IRI("triggers"))
					triggersList, err := p.Iterate(nil).AllValues(graph.store)
					Expect(err).To(BeNil())
					Expect(len(triggersList)).To(Equal(2))
					for _, v := range triggersList {
						key, err := NewKeyFromQuadValue(v)
						Expect(err).To(BeNil())
						switch {
						case key.Equals(flowKey1):
						case key.Equals(flowKey2):
						default:
							Fail("Unknown Flow uuid connected to Path")
						}
					}
				})
			})
			Context("When either id does not exists", func() {
				It("Then an error will be thrown", func() {
					// Save Path
					pathTriggerRoute := "/test"
					pathTriggerType := "path-trigger"
					pathTrigger := messenger.Path{
						Route: pathTriggerRoute,
						Type:  pathTriggerType,
					}
					pathTriggerKey, err := graph.SavePath(pathTrigger)
					Expect(err).To(BeNil())
					Expect(pathTriggerKey).ToNot(Equal(Key{}))

					// Save Flow
					flow := Flow{
						Name:        "Flow Name",
						Description: "Flow Description",
						Path: &messenger.Path{
							Route: "/some-route",
							Type:  "mqtt",
						},
					}
					flowKey, err := graph.SaveFlow(flow)
					Expect(err).To(BeNil())
					Expect(flowKey).ToNot(Equal(Key{}))

					// Connect bad Flow id or Path id
					err = graph.ChainNextFlowToPath(NewRandomKey(), pathTriggerKey)
					Expect(err).ToNot(BeNil())
					err = graph.ChainNextFlowToPath(flowKey, NewRandomKey())
					Expect(err).ToNot(BeNil())
				})
			})
		})
		Describe("Given a Path triggers Flows", func() {
			Context("When a Path UUID is given", func() {
				It("Then a list of Flows should be returned", func() {
					// Save Path
					pathTriggerRoute := "/test"
					pathTriggerType := "path-trigger"
					pathTrigger := messenger.Path{
						Route: pathTriggerRoute,
						Type:  pathTriggerType,
					}
					pathTriggerKey, err := graph.SavePath(pathTrigger)
					Expect(err).To(BeNil())
					Expect(pathTriggerKey).ToNot(Equal(Key{}))

					// Save both Flows
					flow := Flow{
						Name:        "Flow Name",
						Description: "Flow Description",
						Path: &messenger.Path{
							Route: "/some-route",
							Type:  "mqtt",
						},
					}
					flowKey1, err := graph.SaveFlow(flow)
					Expect(err).To(BeNil())
					Expect(flowKey1).ToNot(Equal(Key{}))
					flowKey2, err := graph.SaveFlow(flow)
					Expect(err).To(BeNil())
					Expect(flowKey2).ToNot(Equal(Key{}))

					// Connect Flow to Path
					err = graph.ChainNextFlowToPath(flowKey1, pathTriggerKey)
					Expect(err).To(BeNil())
					err = graph.ChainNextFlowToPath(flowKey2, pathTriggerKey)
					Expect(err).To(BeNil())

					// Get Flows
					flows, keys, err := graph.GetNextFlows(pathTriggerKey)
					Expect(err).To(BeNil())
					Expect(len(flows)).To(Equal(2))
					Expect(flows[0]).To(Equal(flow))
					Expect(flows[1]).To(Equal(flow))
					Expect(keys).To(ContainElement(flowKey1))
					Expect(keys).To(ContainElement(flowKey2))
				})
			})
		})
	})
})
