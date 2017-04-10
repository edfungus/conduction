// +build all unit

package main_test

import (
	// . "github.com/edfungus/conduction"

	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"
)

var _ = Describe("Conduction", func() {
	Describe("Storage", func() {
		Describe("Given checking if Path exists", func() {
			Context("When Path does exists in the database", func() {
				It("Then `true` should be returned", func() {
				})
			})
			Context("When Path does not exists in the database", func() {
				It("Then `false` should be returned", func() {
				})
			})
		})
	})
})
