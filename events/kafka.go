package events

import (
	"log"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/golang/protobuf/proto"
)

type Kafka struct {
	configs                      Configs
	broker, topic, consumerGroup string

	producer sarama.SyncProducer
	consumer *cluster.Consumer

	messages chan *Message
	errors   chan error
}

type Configs struct {
	Producer sarama.Config
	Consumer cluster.Config
}

// NewKafka creates a new Kafka instance
func NewKafka(broker string, topic string, consumerGroup string, configs Configs) Kafka {

	k := Kafka{
		configs:       configs,
		broker:        broker,
		topic:         topic,
		consumerGroup: consumerGroup,
		messages:      make(chan *Message),
		errors:        make(chan error),
	}

	return k
}

// Send sends a message to the designated topic
func (k *Kafka) Send(msg *Message) error {
	out, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	if k.producer == nil {
		err = createProducer(k, &k.producer)
		if err != nil {
			return err
		}
	}

	k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: k.topic,
		Value: sarama.ByteEncoder(out),
	})
	return nil
}

// Messages returns a channel to receive messages
func (k *Kafka) Messages() (<-chan *Message, error) {
	if k.consumer == nil {
		err := createConsumer(k, k.consumer)
		if err != nil {
			return nil, err
		}

		// Pass message as a Message object to k.messages channel
		go func() {
			for {
				msg := <-k.consumer.Messages()
				k.consumer.MarkOffset(msg, "")
				message := &Message{}
				if err := proto.Unmarshal(msg.Value, message); err != nil {
					log.Fatalln("Could not parse message to transport.Message")
					continue
				}
				k.messages <- message
			}
		}()
	}
	return k.messages, nil
}

// Errors returns a channel to receive messages
func (k *Kafka) Errors() (<-chan error, error) {
	if k.consumer == nil {
		err := createConsumer(k, k.consumer)
		if err != nil {
			return nil, err
		}

		// Pass errors to error channel
		go func() {
			for {
				error := <-k.consumer.Errors()
				k.errors <- error
			}
		}()
	}
	return k.errors, nil
}

// Close cleans up both producer and consumer
func (k *Kafka) Close() {
	if k.producer != nil {
		k.producer.Close()
		close(k.messages)
	}
	if k.consumer != nil {
		k.consumer.Close()
		close(k.errors)
	}
}

func createProducer(k *Kafka, producer *sarama.SyncProducer) error {
	p, err := sarama.NewSyncProducer([]string{k.broker}, &k.configs.Producer)
	if err != nil {
		return err
	}
	producer = &p
	return nil
}

func createConsumer(k *Kafka, consumer *cluster.Consumer) error {
	c, err := cluster.NewConsumer([]string{k.broker}, k.consumerGroup, []string{k.topic}, &k.configs.Consumer)
	if err != nil {
		return err
	}
	consumer = c
	return nil
}
