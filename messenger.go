package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

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
	MESSAGE_TOPIC     string = "messageTopic"
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

	go listen(km.consumer, km.messages, km.stop)

	return km, nil
}

// Send sends messages to Kafka
func (km *KafkaMessenger) Send(topic string, msg *pb.Message) error {
	msgB, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	msgS := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msgB),
	}
	_, _, err = km.producer.SendMessage(msgS)
	if err != nil {
		return err
	}
	return nil
}

// Receive returns messages from Kafka
func (km *KafkaMessenger) Receive() <-chan *pb.Message {
	return km.messages
}

// Acknowledge tells Kafka that the message has been received and processed
func (km *KafkaMessenger) Acknowledge(msg *pb.Message) error {
	topic, partition, offset, err := getMessageMetadata(msg)
	if err != nil {
		return err
	}
	km.consumer.MarkPartitionOffset(topic, partition, offset, "")
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

func listen(consumer *cluster.Consumer, messages chan *pb.Message, stop chan bool) {
	for {
		select {
		case msg := <-consumer.Messages():
			msgObj, err := convertMessage(msg)
			if err != nil {
				Logger.Error(err)
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

func convertMessage(msg *sarama.ConsumerMessage) (*pb.Message, error) {
	msgObj := &pb.Message{}
	err := proto.Unmarshal(msg.Value, msgObj)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal message from Kafka. Skipping message. %v", err)
	}
	setMessageMetadata(msg, msgObj)
	return msgObj, nil
}

// Puts the Kafka metadata into Message. This is primarily for acking
func setMessageMetadata(msg *sarama.ConsumerMessage, msgObj *pb.Message) {
	if msgObj.Metadata == nil {
		msgObj.Metadata = make(map[string][]byte)
	}

	partition := make([]byte, 8)
	binary.PutVarint(partition, int64(msg.Partition))
	msgObj.Metadata[MESSAGE_PARTITION] = partition

	offset := make([]byte, 8)
	binary.PutVarint(offset, msg.Offset)
	msgObj.Metadata[MESSAGE_OFFSET] = offset

	msgObj.Metadata[MESSAGE_TOPIC] = []byte(msg.Topic)
}

// Gets the Kafka metadata from Message
func getMessageMetadata(msg *pb.Message) (string, int32, int64, error) {
	var (
		topic     string
		partition int32
		offset    int64
	)
	if val, ok := msg.Metadata[MESSAGE_TOPIC]; ok {
		topic = string(val)
	} else {
		return "", 0, 0, errors.New("Could not find topic in Message metadata")
	}
	if val, ok := msg.Metadata[MESSAGE_PARTITION]; ok {
		partition64, err := binary.ReadVarint(bytes.NewReader(val))
		if err != nil {
			return "", 0, 0, errors.New("Could read partition as int64")
		}
		partition = int32(partition64)
	} else {
		return "", 0, 0, errors.New("Could not find topic in Message metadata")
	}
	if val, ok := msg.Metadata[MESSAGE_OFFSET]; ok {
		err := errors.New("")
		offset, err = binary.ReadVarint(bytes.NewReader(val))
		if err != nil {
			return "", 0, 0, errors.New("Could read offset as int64")
		}
	} else {
		return "", 0, 0, errors.New("Could not find topic in Message metadata")
	}
	return topic, partition, offset, nil
}

// GetProducer gets the sarama producer. Used fro testing only
func (km *KafkaMessenger) GetProducer() sarama.SyncProducer {
	return km.producer
}
