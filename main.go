package main

import (
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/sirupsen/logrus"
)

// Logger controls logging and levels
var Logger = logrus.New()

const (
	kafkaBroker          string = "localhost:9092"
	kafkaConsumerGroup   string = "conduction"
	conductionInputTopic string = "conductionIn"
)

func main() {
	Logger.Info("Hello Conduction!")

	kafkaProducerConfig := sarama.NewConfig()
	kafkaProducerConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaProducerConfig.Producer.Retry.Max = 2
	kafkaProducerConfig.Producer.Return.Successes = true
	kafkaProducer, err := sarama.NewSyncProducer([]string{kafkaBroker}, kafkaProducerConfig)
	if err != nil {
		Logger.Error(err)
	}

	kafkaConsumerConfig := cluster.NewConfig()
	kafkaConsumerConfig.Consumer.Return.Errors = true
	kafkaConsumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaConsumer, err := cluster.NewConsumer([]string{kafkaBroker}, kafkaConsumerGroup, []string{conductionInputTopic}, kafkaConsumerConfig)
	if err != nil {
		Logger.Error(err)
	}

}

func setupLogger() {
	Logger.Level = logrus.WarnLevel
}
