// +build all unit work

package storage

import (
	"github.com/cayleygraph/cayley/quad"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("Conduction", func() {
	Describe("Storage", func() {
		Describe("Given a string", func() {
			Context("When an id needs to be generated for the graph", func() {
				It("Then a quad.IRI object should be created with the ID", func() {
					uuid := uuid.NewV4()
					quad := quad.IRI(uuid.String())
					generatedQuad := uuidToQuadIRI(uuid)
					Expect(generatedQuad).To(Equal(quad))
				})
			})
		})
		Describe("Given a quad.Value", func() {
			Context("When an id is needed from a quad.Value", func() {
				It("Then the id should be retrieved without < or > around the id", func() {
					uuid := uuid.NewV4()
					id1 := quad.StringToValue("<" + uuid.String() + ">")
					id3 := quad.StringToValue(uuid.String())

					uuid1, err := quadValueToUUID(id1)
					Expect(err).To(BeNil())
					Expect(uuid1).To(Equal(uuid))

					_, err = quadValueToUUID(id3)
					Expect(err).ToNot(BeNil()) // cayley adds "" around the id which is invalid
				})
			})
		})
	})
})
