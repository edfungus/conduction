package distributor

import (
	"time"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

// Kafka is an interface to interact with Kafka
type Kafka interface {
	Send([]byte)
	Messages() <-chan *KafkaMessage
	Errors() <-chan error
	MarkOffset(topic string, partition int32, offset int64)
	Close()
}

// KafkaMessage is a message from Kafka
type KafkaMessage struct {
	Data      []byte
	Topic     string
	Partition int32
	Offset    int64
}

// KafkaSarama implements Kafka with the sarama library
type KafkaSarama struct {
	configs                      *KafkaSaramaConfigs
	broker, topic, consumerGroup string

	producer sarama.SyncProducer
	consumer *cluster.Consumer

	messages     chan *KafkaMessage
	stopMessages chan bool
}

// KafkaSaramaConfigs holds the producer and consumer configs
type KafkaSaramaConfigs struct {
	Producer *sarama.Config
	Consumer *cluster.Config
}

// NewKafkaSarama creates a new KafkaSarama
func NewKafkaSarama(broker string, topic string, consumerGroup string, configs *KafkaSaramaConfigs) (*KafkaSarama, error) {
	ks := &KafkaSarama{
		configs:       configs,
		broker:        broker,
		topic:         topic,
		consumerGroup: consumerGroup,
		messages:      make(chan *KafkaMessage),
		stopMessages:  make(chan bool, 1),
	}

	consumer, err := createConsumer(ks)
	if err != nil {
		return nil, err
	}
	producer, err := createProducer(ks)
	if err != nil {
		return nil, err
	}

	ks.consumer = consumer
	ks.producer = producer

	go func() {
		for {
			select {
			case sm := <-ks.consumer.Messages():
				ks.messages <- saramaMessage2KafkaMessage(sm)
			case <-ks.stopMessages:
				ks.consumer.Close()
				return
			}
		}
	}()

	return ks, nil
}

// DefaultKafkaSaramaConfigs returns the default Sarama configuration
func DefaultKafkaSaramaConfigs() *KafkaSaramaConfigs {
	kc := &KafkaSaramaConfigs{
		Producer: sarama.NewConfig(),
		Consumer: cluster.NewConfig(),
	}
	kc.Consumer.Consumer.Return.Errors = true
	kc.Consumer.Consumer.Offsets.Initial = sarama.OffsetOldest
	kc.Producer.Producer.RequiredAcks = sarama.WaitForAll
	kc.Producer.Producer.Retry.Max = 2
	kc.Producer.Producer.Return.Successes = true
	return kc
}

// Send sends message to Kafka
func (ks *KafkaSarama) Send(msg []byte) {
	// ks.producer.Input() <- &sarama.ProducerMessage{
	// 	Topic: ks.topic,
	// 	Value: sarama.ByteEncoder(msg),
	// }
	ks.producer.SendMessage(&sarama.ProducerMessage{
		Topic: ks.topic,
		Value: sarama.ByteEncoder(msg),
	})
}

// Messages gets messages from Kafka
func (ks *KafkaSarama) Messages() <-chan *KafkaMessage {
	return ks.messages
}

// Errors gets errors from Kafka
func (ks *KafkaSarama) Errors() <-chan error {
	return ks.consumer.Errors()
}

// MarkOffset ackowledges a message
func (ks *KafkaSarama) MarkOffset(topic string, partition int32, offset int64) {
	ks.consumer.MarkPartitionOffset(topic, partition, offset, "")
}

// Close ends a session with Kafka
func (ks *KafkaSarama) Close() {
	ks.stopMessages <- true
	time.Sleep(time.Second * 1)
	ks.producer.Close()
}

func saramaMessage2KafkaMessage(sm *sarama.ConsumerMessage) *KafkaMessage {
	km := KafkaMessage{
		Data:      sm.Value,
		Topic:     sm.Topic,
		Partition: sm.Partition,
		Offset:    sm.Offset,
	}
	return &km
}

func createProducer(ks *KafkaSarama) (sarama.SyncProducer, error) {
	p, err := sarama.NewSyncProducer([]string{ks.broker}, ks.configs.Producer)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func createConsumer(ks *KafkaSarama) (*cluster.Consumer, error) {
	c, err := cluster.NewConsumer([]string{ks.broker}, ks.consumerGroup, []string{ks.topic}, ks.configs.Consumer)
	if err != nil {
		return nil, err
	}
	return c, nil
}
