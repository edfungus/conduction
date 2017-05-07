// +build all integration work

package storage

import (
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/quad"
	"github.com/edfungus/conduction/pb"
	uuid "github.com/satori/go.uuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Conduction", func() {
	Describe("Storage", func() {
		var (
			graph  *GraphStorage
			config = &GraphStorageConfig{
				Host:         databaseHost,
				Port:         databasePort,
				User:         "root",
				DatabaseName: databaseName,
				DatabaseType: "cockroach",
			}
		)
		BeforeEach(func() {
			var err error
			graph, err = NewGraphStorage(config)
			Expect(err).To(BeNil())
			Expect(graph).ToNot(BeNil())
		})
		Describe("Given creating a new Storage", func() {
			Context("When Cockroach path is wrong or not available", func() {
				It("Then should return an error", func() {
					badConfig := &GraphStorageConfig{
						Host:         config.Host,
						Port:         8888,
						User:         config.User,
						DatabaseName: config.DatabaseName,
						DatabaseType: config.DatabaseType,
					}
					_, err := NewGraphStorage(badConfig)
					Expect(err).ToNot(BeNil())
				})
			})
			Context("When Cockroach path is correct and available", func() {
				It("Then GraphStorage should return without error", func() {
					graph, err := NewGraphStorage(config)
					Expect(err).To(BeNil())
					Expect(graph).ToNot(BeNil())
				})
			})
		})
		Describe("Given saving a Flow to graph", func() {
			Context("When the Path does not exist", func() {
				It("Then the Path and Flow should also be inserted connected to Path", func() {
					flow := &pb.Flow{
						Name:        "Flow Name",
						Description: "Flow Description",
						Path: &pb.Path{
							Route: "/test2",
							Type:  "mqtt-duplicate",
						},
					}
					id, err := graph.AddFlow(flow)
					Expect(err).To(BeNil())
					Expect(id).ToNot(Equal(uuid.Nil))

					newFlow, err := graph.ReadFlow(id)
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
					path := &pb.Path{
						Route: pathRoute,
						Type:  pathType,
					}
					pathID, err := graph.AddPath(path)
					Expect(err).To(BeNil())
					Expect(pathID).ToNot(Equal(uuid.Nil))

					// Save flow with same path route and type
					flow := &pb.Flow{
						Name:        "Flow Name",
						Description: "Flow Description",
						Path: &pb.Path{
							Route: path.Route,
							Type:  path.Type,
						},
					}
					flowID, err := graph.AddFlow(flow)
					Expect(err).To(BeNil())
					Expect(flowID).ToNot(Equal(uuid.Nil))

					// Make sure Path was not recreated
					p := cayley.StartPath(graph.store, quad.StringToValue(pathType)).In(quad.StringToValue("<type>")).Has(quad.StringToValue("<route>"), quad.StringToValue(pathRoute))
					pathList, err := p.Iterate(nil).AllValues(graph.store)
					Expect(err).To(BeNil())
					Expect(len(pathList)).To(Equal(1))

					// Check that Flow was correctly saved
					newFlow, err := graph.ReadFlow(flowID)
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
					path := &pb.Path{
						Route: pathRoute,
						Type:  pathType,
					}

					// Insert Path twice
					pathID1, err := graph.AddPath(path)
					pathID2, err := graph.AddPath(path)
					Expect(pathID1).To(Equal(pathID2))

					// Check Path was added only once
					p := cayley.StartPath(graph.store, quad.StringToValue(pathType)).In(quad.StringToValue("<type>")).Has(quad.StringToValue("<route>"), quad.StringToValue(pathRoute))
					pathList, err := p.Iterate(nil).AllValues(graph.store)
					Expect(err).To(BeNil())
					Expect(len(pathList)).To(Equal(1))

					// Check Path content
					readPath, err := graph.ReadPath(pathID1)
					Expect(err).To(BeNil())
					Expect(readPath.Route).To(Equal(path.Route))
					Expect(readPath.Type).To(Equal(path.Type))
				})
			})
			Context("When the Path does not already exist", func() {
				It("Then a new Path should be inserted and id returned", func() {
					// Path to save
					pathRoute := "/unique/path"
					pathType := "mqtt-unique"
					path := &pb.Path{
						Route: pathRoute,
						Type:  pathType,
					}

					// Ensure path doesn't exist
					p := cayley.StartPath(graph.store, quad.StringToValue(pathType)).In(quad.StringToValue("<type>")).Has(quad.StringToValue("<route>"), quad.StringToValue(pathRoute))
					pathList, err := p.Iterate(nil).AllValues(graph.store)
					Expect(err).To(BeNil())
					Expect(len(pathList)).To(Equal(0))

					// Insert Path
					pathID, err := graph.AddPath(path)

					// Check Path was added
					p = cayley.StartPath(graph.store, quad.StringToValue(pathType)).In(quad.StringToValue("<type>")).Has(quad.StringToValue("<route>"), quad.StringToValue(pathRoute))
					pathList, err = p.Iterate(nil).AllValues(graph.store)
					Expect(err).To(BeNil())
					Expect(len(pathList)).To(Equal(1))

					// Check Path content
					readPath, err := graph.ReadPath(pathID)
					Expect(err).To(BeNil())
					Expect(readPath.Route).To(Equal(path.Route))
					Expect(readPath.Type).To(Equal(path.Type))
				})
			})
		})
	})
})
