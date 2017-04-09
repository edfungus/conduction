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
	})
})
