// +build all integration

package main_test

import (
	. "github.com/edfungus/conduction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Conduction", func() {
	Describe("Storage", func() {
		Describe("Given creating a new Storage", func() {
			Context("When Cockroach is not available", func() {
				It("Then an error should occur", func() {
					_, err := NewCockroachStorage("postgresql://conductor@localhost:8888/badAddess")
					Expect(err).ToNot(BeNil())
				})
			})
			Context("When Cockroach is available", func() {
				It("Then Storage should be returned", func() {
					db, err := NewCockroachStorage(cockroachURL)
					Expect(err).To(BeNil())
					Expect(db).ToNot(BeNil())
				})
			})
		})
		Describe("Given adding a new Flow recursively", func() {
			Context("When the Flow id does not exist", func() {
				It("Then it should return an error", func() {

				})
			})
			Context("When the Flow incorrectly has no Path", func() {
				It("Then it should return an error", func() {

				})
			})
			Context("When the Flow is correct, has no dependents and has new Path", func() {
				It("Then it should make the Path and Flow, add without error and return Flow id", func() {

				})
			})
			Context("When the Flow is correct, has one dependent Flow that was already created", func() {
				It("Then it should make the new Flow, add without error and return Flow id", func() {

				})
			})
			Context("When the Flow is correct, has one dependent Flow that has not been already created", func() {
				It("Then it should make both new Flows, add without error and return Flow id", func() {

				})
			})
			Context("When the Flow is correct, has multi-level dependent Flows", func() {
				It("Then it should make Flows as needed, add without error and return Flow id", func() {

				})
			})
		})
		Describe("Given adding a Flow without traversing deeper", func() {
			Describe("When the Flow id does not exist", func() {
				It("Then an error should be thrown and nothing should be added to the database", func() {

				})
			})
			Describe("When the Flow is missing id", func() {
				It("Then a new Flow should be made", func() {

				})
			})
			Describe("When the Flow is correct and dependents arg is nil", func() {
				It("Then the Flow should be updated without touching dependents and without an error", func() {
				})
			})
			Describe("When the dependent Flow id does not exist", func() {
				It("Then an error should be thrown", func() {

				})
			})
			Describe("When the Flow is correct and dependents arg is not nil", func() {
				It("Then the Flow should be updated with dependents and without an error", func() {
				})
			})
		})
	})
})
