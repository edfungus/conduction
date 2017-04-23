// +build all integration work

package main_test

import (
	. "github.com/edfungus/conduction"
	"github.com/edfungus/conduction/pb"

	"fmt"

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
			Context("When the Flow is simplified to name and description", func() {
				It("Then the Flow should be inserted and an UUID is returned", func() {
					flow := &pb.Flow{
						Name:        "Flow Name",
						Description: "Flow Description",
						Path: &pb.Path{
							Route: "/test",
							Type:  "mqtt",
						},
					}
					id, err := graph.AddFlow(flow)
					Expect(err).To(BeNil())
					Expect(id).ToNot(BeNil())

					fmt.Println(id)

					newFlow, err := graph.ReadFlow(id)
					Expect(err).To(BeNil())
					Expect(newFlow.Name).To(Equal(flow.Name))
					Expect(newFlow.Description).To(Equal(flow.Description))
					Expect(newFlow.Path.Route).To(Equal(flow.Path.Route))
					Expect(newFlow.Path.Type).To(Equal(flow.Path.Type))
				})
			})
			// Context("When the Path exists", func() {
			// 	It("Then the Flow should be inserted connected to existing Path", func() {

			// 	})
			// })
			// Context("When the Path does not exist", func() {
			// 	It("Then the Path and Flow should also be inserted connected to Path", func() {

			// 	})
			// })
		})
	})
})
