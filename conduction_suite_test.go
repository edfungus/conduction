package main_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const (
	kafkaConsumerGroup string = "conduction-test"
	kafkaInputTopic    string = "conductionIn-test"
	databaseName       string = "conductionTest"
)

var (
	kafkaBroker  string = "localhost:9092"  // Override with KAFKA_URL if necessary
	cockroachURL string = "localhost:26257" // Override with DATABASE_URL if necessary
)

func TestConduction(t *testing.T) {
	if os.Getenv("KAFKA_URL") != "" {
		kafkaBroker = os.Getenv("KAFKA_URL")
	}
	if os.Getenv("DATABASE_URL") != "" {
		cockroachURL = os.Getenv("DATABASE_URL")
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Conduction Suite")
}
