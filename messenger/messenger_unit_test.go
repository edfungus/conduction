// +build all unit

package messenger

import (
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Conduction", func() {
	Describe("Messenger", func() {
		var (
			message *Message = &Message{
				Payload: []byte("payload"),
			}
			topic string = "topic"
		)

		const (
			partition int32 = 1
			offset    int64 = 500
		)

		Describe("Given a sarama message to set kafka metadata to a Message", func() {
			Context("When setting the metadata", func() {
				It("Then the Message should have kafka metadata in the metadata map", func() {
					serialized, err := proto.Marshal(message)
					Expect(err).To(BeNil())

					sm := &sarama.ConsumerMessage{
						Value:     serialized,
						Topic:     topic,
						Partition: partition,
						Offset:    offset,
					}

					msg := &Message{}
					msg.SetMetadataFromConsumerMessage(sm)

					newTopic, newPartition, newOffset, err := msg.getTopicPartitionOffsetFromMessageMetadata()
					Expect(err).To(BeNil())
					Expect(newTopic).To(Equal(topic))
					Expect(newPartition).To(Equal(partition))
					Expect(newOffset).To(Equal(offset))
				})
			})
		})
		Describe("Given a sarama message to convert to a Message", func() {
			Context("When the sarama message has valid value", func() {
				It("Then it should return a Message with values and kafka metadata", func() {
					serialized, err := proto.Marshal(message)
					Expect(err).To(BeNil())

					sm := &sarama.ConsumerMessage{
						Value:     serialized,
						Topic:     topic,
						Partition: partition,
						Offset:    offset,
					}
					msg, err := NewMessageFromSaramaConsumerMessage(sm)
					Expect(err).To(BeNil())
					Expect(msg.Payload).To(Equal(message.Payload))

					newTopic, newPartition, newOffset, err := msg.getTopicPartitionOffsetFromMessageMetadata()
					Expect(err).To(BeNil())
					Expect(newTopic).To(Equal(topic))
					Expect(newPartition).To(Equal(partition))
					Expect(newOffset).To(Equal(offset))
				})
			})
			Context("When the sarama message has invalid value", func() {
				It("Then an error should be thrown for unmarshaling errors", func() {
					badMessage := []byte("some message not Message serialized")
					sm := &sarama.ConsumerMessage{
						Value: badMessage,
					}

					_, err := NewMessageFromSaramaConsumerMessage(sm)
					Expect(err).ToNot(BeNil())
				})
			})
		})
		Describe("Given a Message to get kafka metadata from", func() {
			var messageToRead *Message
			BeforeEach(func() {
				serialized, err := proto.Marshal(message)
				Expect(err).To(BeNil())

				sm := &sarama.ConsumerMessage{
					Value:     serialized,
					Topic:     topic,
					Partition: partition,
					Offset:    offset,
				}
				messageToRead, err = NewMessageFromSaramaConsumerMessage(sm)
				Expect(err).To(BeNil())
				Expect(messageToRead).ToNot(BeNil())
			})
			Context("When all fields are valid and present", func() {
				It("Then the topic, partition and offset should return correctly", func() {
					newTopic, newPartition, newOffset, err := messageToRead.getTopicPartitionOffsetFromMessageMetadata()
					Expect(err).To(BeNil())
					Expect(newTopic).To(Equal(topic))
					Expect(newPartition).To(Equal(partition))
					Expect(newOffset).To(Equal(offset))
				})
			})
			Context("When topic is not present", func() {
				It("Then an error should be returned regarding topic", func() {
					delete(messageToRead.Metadata, messageTopic)
					_, _, _, err := messageToRead.getTopicPartitionOffsetFromMessageMetadata()
					Expect(err).ToNot(BeNil())
				})
			})
			Context("When partition is not present", func() {
				It("Then an error should be returned regarding partition", func() {
					delete(messageToRead.Metadata, messagePartition)
					_, _, _, err := messageToRead.getTopicPartitionOffsetFromMessageMetadata()
					Expect(err).ToNot(BeNil())
				})
			})
			Context("When offset is not present", func() {
				It("Then an error should be returned regarding offset", func() {
					delete(messageToRead.Metadata, messageOffset)
					_, _, _, err := messageToRead.getTopicPartitionOffsetFromMessageMetadata()
					Expect(err).ToNot(BeNil())
				})
			})
			Context("When the metadata map is nil", func() {
				It("Then an error should be returned", func() {
					messageToRead.Metadata = nil
					_, _, _, err := messageToRead.getTopicPartitionOffsetFromMessageMetadata()
					Expect(err).ToNot(BeNil())
				})
			})
		})
	})
})
