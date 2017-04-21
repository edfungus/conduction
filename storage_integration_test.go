// +build all integration work

package main_test

import (
	. "github.com/edfungus/conduction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Conduction", func() {
	Describe("Storage", func() {
		var (
			graph  *GraphStorage
			config = &GraphStorageConfig{
				// Host:         databaseHost, // To be set in test
				Port:         databasePort,
				User:         "root",
				DatabaseName: databaseName,
				DatabaseType: "cockroach",
			}
		)
		BeforeEach(func() {
			config.Host = databaseHost
			var err error
			graph, err = NewGraphStorage(config)
			Expect(err).To(BeNil())
			Expect(graph).ToNot(BeNil())
		})
		Describe("Given creating a new Storage", func() {
			Context("When Cockroach path is wrong or not available", func() {
				It("Then should return an error", func() {
					config.Host = "badHost"
					_, err := NewGraphStorage(config)
					Expect(err).ToNot(BeNil())
				})
			})
			Context("When Cockroach path is correct and available", func() {
				It("Then GraphStorage should return without error", func() {
					config.Host = databaseHost
					graph, err := NewGraphStorage(config)
					Expect(err).To(BeNil())
					Expect(graph).ToNot(BeNil())
				})
			})
		})
	})
})
