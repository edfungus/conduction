package events

import (
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

type KafkaInteractions interface {
	Send([]byte) error
	Messages() <-chan *sarama.ConsumerMessage
	Errors() <-chan error
	MarkOffset(*sarama.ConsumerMessage)
	Close()
}

type KafkaInteract struct {
	configs                      KafkaConfigs
	broker, topic, consumerGroup string

	producer sarama.SyncProducer
	consumer *cluster.Consumer
}

func NewKafkaInteract(broker string, topic string, consumerGroup string, configs KafkaConfigs) (*KafkaInteract, error) {
	ki := KafkaInteract{
		configs:       configs,
		broker:        broker,
		topic:         topic,
		consumerGroup: consumerGroup,
	}

	producer, err := createProducer(&ki)
	if err != nil {
		return nil, err
	}
	consumer, err := createConsumer(&ki)
	if err != nil {
		return nil, err
	}

	ki.producer = producer
	ki.consumer = consumer

	return &ki, nil
}

// Send sends message to Kafka
func (ki *KafkaInteract) Send(msg []byte) error {
	_, _, err := ki.producer.SendMessage(&sarama.ProducerMessage{
		Topic: ki.topic,
		Value: sarama.ByteEncoder(msg),
	})
	return err
}

func (ki *KafkaInteract) Messages() <-chan *sarama.ConsumerMessage {
	return ki.consumer.Messages()
}

func (ki *KafkaInteract) Errors() <-chan error {
	return ki.consumer.Errors()
}

func (ki *KafkaInteract) MarkOffset(msg *sarama.ConsumerMessage) {
	ki.consumer.MarkOffset(msg, "")
}

func (ki *KafkaInteract) Close() {
	ki.consumer.Close()
	ki.producer.Close()
}

func createProducer(ki *KafkaInteract) (sarama.SyncProducer, error) {
	p, err := sarama.NewSyncProducer([]string{ki.broker}, &ki.configs.Producer)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func createConsumer(ki *KafkaInteract) (*cluster.Consumer, error) {
	c, err := cluster.NewConsumer([]string{ki.broker}, ki.consumerGroup, []string{ki.topic}, &ki.configs.Consumer)
	if err != nil {
		return nil, err
	}
	return c, nil
}
