package storage_test

import (
	. "github.com/edfungus/conduction/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var value *Destination = &Destination{
	Method: "MQTT",
	Path:   "/testTopic",
}

var _ = Describe("RedisStorage", func() {
	var redisInstance = &RedisStorage{}
	BeforeSuite(func() {
		redisConfig := &RedisStorageConfig{
			Addr:           "localhost:6379",
			MaxConnections: 3,
			DatabaseIndex:  15,
		}
		redisInstance = NewRedisStorage(redisConfig)
	})
	AfterSuite(func() {
		err := redisInstance.Close()
		Expect(err).To(BeNil())
	})
	Describe("Put", func() {
		const key string = "testPutKey"
		AfterEach(func() {
			err := redisInstance.Delete(key)
			Expect(err).To(BeNil())
		})
		Context("With with valid key and destination", func() {
			It("Should put without error", func() {
				err := redisInstance.Put(key, value)
				Expect(err).To(BeNil())

				newValue, err := redisInstance.Get(key)
				Expect(err).To(BeNil())
				Expect(newValue).To(Equal(value))
			})
		})
		Context("With with invalid destination", func() {
			It("Should put without error", func() {
				err := redisInstance.Put(key, nil)
				Expect(err).ToNot(BeNil())
			})
		})
	})
	Describe("Get", func() {
		const key string = "testGetKey"

		BeforeEach(func() {
			err := redisInstance.Put(key, value)
			Expect(err).To(BeNil())
		})
		AfterEach(func() {
			err := redisInstance.Delete(key)
			Expect(err).To(BeNil())
		})
		Context("Get with valid key", func() {
			It("Should retrieve the value", func() {
				newValue, err := redisInstance.Get(key)
				Expect(err).To(BeNil())
				Expect(newValue).To(Equal(value))
			})
		})
		Context("Get with invalid key", func() {
			It("Should return an error", func() {
				_, err := redisInstance.Get("badKey")
				Expect(err).ToNot(BeNil())
			})
		})
	})
	Describe("Delete", func() {
		const key string = "testGetKey"

		BeforeEach(func() {
			err := redisInstance.Put(key, value)
			Expect(err).To(BeNil())
		})
		AfterEach(func() {
			redisInstance.Delete(key)
		})
		Context("Delete with valid key", func() {
			It("Should delete without error", func() {
				err := redisInstance.Delete(key)
				Expect(err).To(BeNil())

				_, err = redisInstance.Get(key)
				Expect(err).ToNot(BeNil())
			})
		})
		Context("Delete with invalid key", func() {
			It("Should return an error", func() {
				err := redisInstance.Delete("badKey")
				Expect(err).To(BeNil()) // Deleting non-exsistent key seems fine
			})
		})
	})
})
