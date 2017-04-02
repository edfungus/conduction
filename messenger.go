package main

import (
	"fmt"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/edfungus/conduction/pb"
	"github.com/golang/protobuf/proto"
)

// const (
// 	kafkaBroker          string = "localhost:9092"
// 	kafkaConsumerGroup   string = "conduction"
// 	conductionInputTopic string = "conductionIn"
// )

// Messenger orchestrates communication between conduction modules
type Messenger interface {
	Send(topic string, msg pb.Message) error
	Receive() (<-chan pb.Message, error)
	Acknowledge(pb.Message) error
	Close() error
}

// KafkaMessenger implements Messenger using Kafka
type KafkaMessenger struct {
	producer sarama.SyncProducer
	consumer *cluster.Consumer

	messages chan *pb.Message
}

type KafkaMessengerConfig struct {
	ConsumerGroup string
	InputTopic    string
}

// NewKafkaMessenger returns a new KafkaMessenger
func NewKafkaMessenger(broker string, config *KafkaMessengerConfig) (*KafkaMessenger, error) {
	kafkaProducerConfig := sarama.NewConfig()
	kafkaProducerConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaProducerConfig.Producer.Retry.Max = 2
	kafkaProducerConfig.Producer.Return.Successes = true
	kafkaProducer, err := sarama.NewSyncProducer([]string{broker}, kafkaProducerConfig)
	if err != nil {
		return nil, err
	}

	kafkaConsumerConfig := cluster.NewConfig()
	kafkaConsumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaConsumer, err := cluster.NewConsumer([]string{broker}, config.ConsumerGroup, []string{config.InputTopic}, kafkaConsumerConfig)
	if err != nil {
		return nil, err
	}

	messages := make(chan *pb.Message)

	go toMessage(kafkaConsumer, messages)

	return &KafkaMessenger{
		producer: kafkaProducer,
		consumer: kafkaConsumer,
		messages: messages,
	}, nil
}

func (km *KafkaMessenger) Send(topic string, msg pb.Message) error {

}

func (km *KafkaMessenger) Receive() (<-chan pb.Message, error) {

}

func (km *KafkaMessenger) Acknowledge(pb.Message) error {

}

// Close stops the Kafka Messenger from sending and receiving messages
func (km *KafkaMessenger) Close() error {
	perr := km.producer.Close()
	cerr := km.consumer.Close()
	if perr != nil || cerr != nil {
		return fmt.Errorf("Error closing Kafka Messenger. Kafka Producer Error: %v Kafka Consumer Error: %v", perr, cerr)
	}
	return nil
}

func toMessage(consumer *cluster.Consumer, messages chan *pb.Message) {
	for {
		msg := <-consumer.Messages()
		msgObj := &pb.Message{}
		err := proto.Unmarshal(msg.Value, msgObj)
		if err != nil {
			Logger.Error("Could not unmarshal message from Kafka. Skipping message. Error:", err)
			consumer.MarkOffset(msg, "")
		}
		messages <- msgObj
	}
}
