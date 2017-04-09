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
)

var (
	kafkaBroker  string = "localhost:9092"  // Override with KAFKA_URL
	cockroachURL string = "localhost:26257" // Override with COCKROACH_URL
)

func TestConduction(t *testing.T) {
	if os.Getenv("KAFKA_URL") != "" {
		kafkaBroker = os.Getenv("KAFKA_URL")
	}
	if os.Getenv("COCKROACH_URL") != "" {
		cockroachURL = os.Getenv("COCKROACH_URL")
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Conduction Suite")
}
