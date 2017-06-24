// +build all unit

package storage

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Conduction", func() {
	Describe("Key", func() {
		Describe("Given a string", func() {
			Context("When there are < and > around a string", func() {
				It("Then the < and > can be removed leaving only the string", func() {
					string1 := "<test>"
					string2 := "test>"
					string3 := "<test"
					string4 := "test"
					string5 := "te<>st"

					Expect(removeIDBrackets(string1)).To(Equal("test"))
					Expect(removeIDBrackets(string2)).To(Equal("test"))
					Expect(removeIDBrackets(string3)).To(Equal("test"))
					Expect(removeIDBrackets(string4)).To(Equal("test"))
					Expect(removeIDBrackets(string5)).To(Equal("te<>st"))
				})
			})
		})
		Describe("Given two Keys", func() {
			Context("When they are different", func() {
				It("Then they should not be equal", func() {
					key1 := NewRandomKey()
					key2 := NewRandomKey()
					Expect(key1.Equals(key2)).To(Equal(false))
					Expect(key2.Equals(key1)).To(Equal(false))
				})
			})
		})
		Describe("Given two Keys", func() {
			Context("When they are same", func() {
				It("Then they should be equal", func() {
					key1 := NewRandomKey()
					key2 := key1
					Expect(key1.Equals(key2)).To(Equal(true))
					Expect(key2.Equals(key1)).To(Equal(true))
				})
			})
		})
		Describe("Given a Key", func() {
			Context("When converting to Quad Value and back", func() {
				It("Then the new key should equal the old key", func() {
					key := NewRandomKey()
					key1, err := NewKeyFromQuadValue(key.QuadValue())
					Expect(err).To(BeNil())
					Expect(key1).To(Equal(key))
				})
			})
		})
		Describe("Given a Key", func() {
			Context("When converting to Quad Value and back", func() {
				It("Then the new key should equal the old key", func() {
					key := NewRandomKey()
					key1, err := NewKeyFromQuadIRI(key.QuadIRI())
					Expect(err).To(BeNil())
					Expect(key1).To(Equal(key))
				})
			})
		})
	})
})
