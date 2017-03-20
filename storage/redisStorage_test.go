package storage_test

import (
	. "github.com/edfungus/conduction/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	key string = "redisStorageTestKey"
)

var _ = Describe("RedisStorage", func() {
	BeforeSuite(func() {
		// connect to redis
	})
	Describe("Put & Get", func() {
		BeforeEach(func() {
			//store the value
		})
		AfterEach(func() {
			//cleanup
		})
		Context("With valid key", func() {
			It("Should retrieve the value", func() {

			})
		})
		Context("With invalid key", func() {
			It("Should return an error", func() {

			})
		})
	})
})
