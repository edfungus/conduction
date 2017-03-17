package events

import (
	"log"

	"github.com/golang/protobuf/proto"
)

// KafkaDistributor is the Kafka version of the distributor
type KafkaDistributor struct {
	kafka    Kafka
	messages chan *DistributorMessage
	errors   chan error
}

// NewKafkaDistributor creates a new Kafka instance
func NewKafkaDistributor(kafka Kafka) (*KafkaDistributor, error) {
	kd := KafkaDistributor{
		kafka:    kafka,
		messages: make(chan *DistributorMessage),
		errors:   make(chan error),
	}

	go consumeMessages(kd.messages, kd.kafka.Messages())
	go consumeErrors(kd.errors, kd.kafka.Errors())

	return &kd, nil
}

// Send sends a message to the designated topic
func (kd *KafkaDistributor) Send(msg *Message) error {
	out, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	err = kd.kafka.Send(out)
	if err != nil {
		return err
	}
	return nil
}

// Messages returns a channel to receive messages
func (kd *KafkaDistributor) Messages() <-chan *DistributorMessage {
	return kd.messages
}

// Errors returns a channel to receive messages
func (kd *KafkaDistributor) Errors() <-chan error {
	return kd.errors
}

// Acknowledge lets Kafka know that the message has been processed
func (kd *KafkaDistributor) Acknowledge(msg *DistributorMessage) {
	kd.kafka.MarkOffset(msg.topic, msg.partition, msg.offset)
}

// Close cleans up both producer and consumer
func (kd *KafkaDistributor) Close() {
	kd.kafka.Close()
	close(kd.errors)
	close(kd.messages)
}

func consumeMessages(messagesOut chan *DistributorMessage, messagesIn <-chan *KafkaMessage) {
	for {
		km := <-messagesIn
		message := &Message{}
		if err := proto.Unmarshal(km.Data, message); err != nil {
			log.Fatalln("Could not parse message to transport.Message")
			continue
		}
		dm := &DistributorMessage{
			Message:   message,
			topic:     km.Topic,
			partition: km.Partition,
			offset:    km.Offset,
		}
		messagesOut <- dm
	}
}
func consumeErrors(errorsOut chan error, errorsIn <-chan error) {
	for {
		error := <-errorsIn
		errorsOut <- error
	}
}
