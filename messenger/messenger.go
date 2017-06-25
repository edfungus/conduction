package messenger

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/sirupsen/logrus"
)

// Messenger orchestrates communication between conduction modules
type Messenger interface {
	Start()
	Send(topic string, msg Message) error
	Receive() <-chan Message
	Acknowledge(Message) error
	Close() error
}

// Logger logs but can be replaced
var Logger = logrus.New()

// KafkaMessenger implements Messenger using Kafka
type KafkaMessenger struct {
	producer sarama.SyncProducer
	consumer *cluster.Consumer

	messages chan *Message
	stop     chan bool
}

type KafkaMessengerConfig struct {
	ConsumerGroup string
	InputTopic    string
}

// NewKafkaMessenger returns a new KafkaMessenger
func NewKafkaMessenger(broker string, config *KafkaMessengerConfig) (*KafkaMessenger, error) {
	kafkaConsumerConfig := newKafkaConsumerConfig()
	kafkaConsumer, err := cluster.NewConsumer([]string{broker}, config.ConsumerGroup, []string{config.InputTopic}, kafkaConsumerConfig)
	if err != nil {
		return nil, err
	}

	kafkaProducerConfig := newKafkaProducerConfig()
	kafkaProducer, err := sarama.NewSyncProducer([]string{broker}, kafkaProducerConfig)
	if err != nil {
		return nil, err
	}

	km := &KafkaMessenger{
		producer: kafkaProducer,
		consumer: kafkaConsumer,
		messages: make(chan *Message),
		stop:     make(chan bool, 1),
	}
	km.startConsuming()

	return km, nil
}

// Start begins listening to the messages coming into the topics
func (km *KafkaMessenger) startConsuming() {
	go listen(km.consumer, km.messages, km.stop)
}

// Send sends messages to Kafka
func (km *KafkaMessenger) Send(topic string, message *Message) error {
	saramaMessage, err := message.ConvertToSaramaProducerMessage(topic)
	if err != nil {
		return err
	}
	_, _, err = km.producer.SendMessage(saramaMessage)
	if err != nil {
		return err
	}
	return nil
}

// Receive returns messages from Kafka
func (km *KafkaMessenger) Receive() <-chan *Message {
	return km.messages
}

// Acknowledge tells Kafka that the message has been received and processed
func (km *KafkaMessenger) Acknowledge(message *Message) error {
	topic, partition, offset, err := message.getTopicPartitionOffsetFromMessageMetadata()
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

func newKafkaProducerConfig() *sarama.Config {
	kafkaProducerConfig := sarama.NewConfig()
	kafkaProducerConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaProducerConfig.Producer.Retry.Max = 2
	kafkaProducerConfig.Producer.Return.Successes = true
	return kafkaProducerConfig
}

func newKafkaConsumerConfig() *cluster.Config {
	kafkaConsumerConfig := cluster.NewConfig()
	kafkaConsumerConfig.Consumer.Return.Errors = true
	kafkaConsumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	return kafkaConsumerConfig
}

func listen(consumer *cluster.Consumer, messages chan *Message, stop chan bool) {
	for {
		select {
		case msg := <-consumer.Messages():

			message, err := NewMessageFromSaramaConsumerMessage(msg)
			if err != nil {
				Logger.Debugln(err)
				consumer.MarkOffset(msg, "")
				continue
			}
			messages <- message
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

// GetProducer gets the sarama producer. Used for testing only
func (km *KafkaMessenger) getProducer() sarama.SyncProducer {
	return km.producer
}
