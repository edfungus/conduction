// +build all unit work

package storage

import (
	"fmt"

	"github.com/cayleygraph/cayley/quad"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
)

var _ = Describe("Conduction", func() {
	Describe("Storage", func() {
		Describe("Given a string", func() {
			Context("When an id needs to be generated for the graph", func() {
				It("Then a quad.IRI object should be created with the ID", func() {
					id := fmt.Sprintf("id:%s", uuid.NewRandom().String())
					quad := quad.IRI(id)
					generatedQuad := stringToQuadIRI(id)
					Expect(generatedQuad).To(Equal(quad))
				})
			})
		})
		Describe("Given a quad.Value", func() {
			Context("When an id is needed from a quad.Value", func() {
				It("Then the id should be retrieved without < or > around the id", func() {
					id1 := quad.StringToValue("<test>")
					id2 := quad.StringToValue("<te<>st>")
					id3 := quad.StringToValue("test")

					Expect(quadToString(id1)).To(Equal("test"))
					Expect(quadToString(id2)).To(Equal("te<>st"))
					Expect(quadToString(id3)).To(Equal(`"test"`))
				})
			})
		})
	})
})
