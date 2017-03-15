package events

import (
	"log"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/golang/protobuf/proto"
)

type Kafka struct {
	interact KafkaInteractions
	messages chan *Message
	errors   chan error
}

type KafkaConfigs struct {
	Producer sarama.Config
	Consumer cluster.Config
}

// NewKafka creates a new Kafka instance
func NewKafka(broker string, topic string, consumerGroup string, configs KafkaConfigs) (*Kafka, error) {
	ki, err := NewKafkaInteract(broker, topic, consumerGroup, configs)
	if err != nil {
		return nil, err
	}

	messages := make(chan *Message)
	errors := make(chan error)

	k := Kafka{
		interact: ki,
		messages: messages,
		errors:   errors,
	}

	go consumeMessages(messages, ki.Messages())
	go consumeErrors(errors, ki.Errors())

	return &k, nil
}

// Send sends a message to the designated topic
func (k *Kafka) Send(msg *Message) error {
	out, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	err = k.interact.Send(out)
	if err != nil {
		return err
	}
	return nil
}

// Messages returns a channel to receive messages
func (k *Kafka) Messages() <-chan *Message {
	return k.messages
}

// Errors returns a channel to receive messages
func (k *Kafka) Errors() <-chan error {
	return k.errors
}

// Close cleans up both producer and consumer
func (k *Kafka) Close() {
	k.interact.Close()
	close(k.errors)
	close(k.messages)
}

func consumeMessages(messagesOut chan *Message, messagesIn <-chan *sarama.ConsumerMessage) {
	for {
		msg := <-messagesIn
		message := &Message{}
		if err := proto.Unmarshal(msg.Value, message); err != nil {
			log.Fatalln("Could not parse message to transport.Message")
			continue
		}
		messagesOut <- message
		// ki.MarkOffset(msg) // This should happen later... somehow we need to trigger acks from somewhere else
	}
}
func consumeErrors(errorsOut chan error, errorsIn <-chan error) {
	for {
		error := <-errorsIn
		errorsOut <- error
	}
}
