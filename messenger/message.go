package messenger

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/gogo/protobuf/proto"
)

const (
	messagePartition string = "messagePartition"
	messageOffset    string = "messageOffset"
	messageTopic     string = "messageTopic"
)

// NewMessageFromSaramaConsumerMessage returns new Message
func NewMessageFromSaramaConsumerMessage(consumerMessage *sarama.ConsumerMessage) (*Message, error) {
	message := Message{}
	err := proto.Unmarshal(consumerMessage.Value, &message)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal message from Kafka. Skipping message. %v", err)
	}
	message.SetMetadataFromConsumerMessage(consumerMessage)
	return &message, nil
}

// ConvertToSaramaProducerMessage returns a Kafka Producer Message from Message
func (m *Message) ConvertToSaramaProducerMessage(topic string) (*sarama.ProducerMessage, error) {
	producerMessage, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(producerMessage),
	}, nil
}

// SetMetadataFromConsumerMessage sets the Kafka metadata like partiion, offset and topic into Message
func (m *Message) SetMetadataFromConsumerMessage(consumerMessage *sarama.ConsumerMessage) {
	if m.Metadata == nil {
		m.Metadata = make(map[string][]byte)
	}

	m.Metadata[messagePartition] = int64ToByteArray(int64(consumerMessage.Partition))
	m.Metadata[messageOffset] = int64ToByteArray(consumerMessage.Offset)
	m.Metadata[messageTopic] = []byte(consumerMessage.Topic)
}

func int64ToByteArray(x int64) []byte {
	buf := make([]byte, 8)
	binary.PutVarint(buf, x)
	return buf
}

func (m *Message) getTopicPartitionOffsetFromMessageMetadata() (string, int32, int64, error) {
	topic, err := m.getTopic()
	if err != nil {
		return "", 0, 0, err
	}
	partition, err := m.getPartition()
	if err != nil {
		return "", 0, 0, err
	}
	offset, err := m.getOffset()
	if err != nil {
		return "", 0, 0, err
	}
	return topic, partition, offset, nil
}

func (m *Message) getTopic() (string, error) {
	if val, ok := m.Metadata[messageTopic]; ok {
		return string(val), nil
	}
	return "", errors.New("Could not find topic in Message metadata")
}

func (m *Message) getPartition() (int32, error) {
	if val, ok := m.Metadata[messagePartition]; ok {
		partition64, err := binary.ReadVarint(bytes.NewReader(val))
		if err != nil {
			return 0, errors.New("Could read partition as int64")
		}
		return int32(partition64), nil
	}
	return 0, errors.New("Could not find partition in Message metadata")
}

func (m *Message) getOffset() (int64, error) {
	if val, ok := m.Metadata[messageOffset]; ok {
		return binary.ReadVarint(bytes.NewReader(val))
	}
	return 0, errors.New("Could not find offset in Message metadata")
}
