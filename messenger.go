package main

import (
	"fmt"
	"time"

	"log"

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
	Receive() <-chan pb.Message
	Acknowledge(pb.Message) error
	Close() error
}

// KafkaMessenger implements Messenger using Kafka
type KafkaMessenger struct {
	producer sarama.SyncProducer
	consumer *cluster.Consumer

	messages chan *pb.Message
	stop     chan bool
}

type KafkaMessengerConfig struct {
	ConsumerGroup string
	InputTopic    string
}

const (
	MESSAGE_PARTITION string = "messagePartition"
	MESSAGE_OFFSET    string = "messageOffset"
)

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
	kafkaConsumerConfig.Consumer.Return.Errors = true
	kafkaConsumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaConsumer, err := cluster.NewConsumer([]string{broker}, config.ConsumerGroup, []string{config.InputTopic}, kafkaConsumerConfig)
	if err != nil {
		return nil, err
	}

	messages := make(chan *pb.Message)
	stop := make(chan bool)
	km := &KafkaMessenger{
		producer: kafkaProducer,
		consumer: kafkaConsumer,
		messages: messages,
		stop:     stop,
	}

	go toMessage(km.consumer, km.messages, km.stop)

	return km, nil
}

func (km *KafkaMessenger) Send(topic string, msg *pb.Message) error {
	return nil
}

func (km *KafkaMessenger) Receive() <-chan pb.Message {
	return nil
}

func (km *KafkaMessenger) Acknowledge(pb.Message) error {
	return nil
}

// Close stops the Kafka Messenger from sending and receiving messages
func (km *KafkaMessenger) Close() error {
	km.stop <- true
	time.Sleep(time.Second * 1)
	err := km.producer.Close()
	if err != nil {
		return fmt.Errorf("Error closing Kafka Producer. %v", err)
	}
	return nil
}

func toMessage(consumer *cluster.Consumer, messages chan *pb.Message, stop chan bool) {
	for {
		select {
		case msg := <-consumer.Messages():
			log.Println("MESSAGE!")
			log.Println(msg.Topic)
			msgObj := &pb.Message{}
			log.Println(msg.Value)
			err := proto.Unmarshal(msg.Value, msgObj)
			if err != nil {
				Logger.Error("Could not unmarshal message from Kafka. Skipping message. Error:", err)
				consumer.MarkOffset(msg, "")
			}
			messages <- msgObj
		case err := <-consumer.Errors():
			Logger.Error(err.Error())
		case <-stop:
			err := consumer.Close()
			if err != nil {
				Logger.Errorf("Error closing Kafka Consumer. %v", err)
			}
			return
		}
	}
}

func (km *KafkaMessenger) GetProducer() sarama.SyncProducer {
	return km.producer
}
