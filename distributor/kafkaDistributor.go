package distributor

import (
	"errors"
	"time"

	"github.com/golang/protobuf/proto"
)

// KafkaDistributor is the Kafka version of the distributor
type KafkaDistributor struct {
	kafka    Kafka
	messages chan *DistributorMessage
	errors   chan error

	stopMessages chan bool
	stopErrors   chan bool
}

// NewKafkaDistributor creates a new Kafka instance
func NewKafkaDistributor(kafka Kafka) *KafkaDistributor {
	kd := KafkaDistributor{
		kafka:        kafka,
		messages:     make(chan *DistributorMessage),
		errors:       make(chan error),
		stopMessages: make(chan bool, 1),
		stopErrors:   make(chan bool, 1),
	}

	go kd.consumeMessages()
	go consumeErrors(kd.errors, kd.kafka.Errors(), kd.stopErrors)

	return &kd
}

// Send sends a message to the designated topic
func (kd *KafkaDistributor) Send(msg *Message) error {
	out, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	kd.kafka.Send(out)
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
	kd.stopMessages <- true
	time.Sleep(time.Second * 1)
}

func (kd *KafkaDistributor) consumeMessages() {
	for {
		select {
		case km := <-kd.kafka.Messages():
			message := &Message{}
			if err := proto.Unmarshal(km.Data, message); err != nil {
				// log.Println("Could not parse message to transport.Message, skipping...")
				kd.errors <- errors.New("could not parse message to transport.Message, skipping message" + string(km.Data))
				kd.kafka.MarkOffset(km.Topic, km.Partition, km.Offset)
				continue
			}
			dm := &DistributorMessage{
				Message:   message,
				topic:     km.Topic,
				partition: km.Partition,
				offset:    km.Offset,
			}
			kd.messages <- dm
		case <-kd.stopMessages:
			close(kd.messages)
			kd.stopErrors <- true
			return
		}
	}
}
func consumeErrors(errorsOut chan error, errorsIn <-chan error, stopErrors chan bool) {
	for {
		select {
		case error := <-errorsIn:
			errorsOut <- error
		case <-stopErrors:
			close(errorsOut)
			return
		}

	}
}
