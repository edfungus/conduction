// +build all integration work

package messenger

import (
	"time"

	"github.com/Shopify/sarama"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Conduction", func() {
	Describe("Messenger", func() {
		var (
			config = &KafkaMessengerConfig{
				ConsumerGroup:   kafkaConsumerGroup,
				TopicsToConsume: []string{kafkaInputTopic},
			}
		)
		Describe("Given creating a new Messenger", func() {
			Context("When Kafka is not available", func() {
				It("Then an error should occur", func() {
					_, err := NewKafkaMessenger("bad_broker_address", config)
					Expect(err).ToNot(BeNil())
				})
			})
			Context("When Kafka is available", func() {
				It("Then Messenger should be returned", func() {
					messenger, err := NewKafkaMessenger(kafkaBroker, config)
					Expect(err).To(BeNil())
					Expect(messenger).ToNot(BeNil())
					err = messenger.Close()
					Expect(err).To(BeNil())
				})
			})
		})
		Describe("Given Kafka is connected", func() {
			var (
				messenger *KafkaMessenger
				message   = &Message{
					Payload: []byte("payload"),
				}
			)
			BeforeEach(func() {
				var err error
				messenger, err = NewKafkaMessenger(kafkaBroker, config)
				Expect(err).To(BeNil())
				Expect(messenger).ToNot(BeNil())
			})
			AfterEach(func() {
				err := messenger.Close()
				Expect(err).To(BeNil())
			})
			Context("When Messenger sends a message", func() {
				It("Then the message should send without an error and received", func() {
					err := messenger.Send(kafkaInputTopic, message)
					Expect(err).To(BeNil())

					select {
					case msg := <-messenger.Receive():
						Expect(err).To(BeNil())
						Expect(msg.Payload).To(Equal(message.Payload))
						Expect(msg.Metadata).To(HaveKey(messagePartition))
						Expect(msg.Metadata).To(HaveKey(messageOffset))
						Expect(msg.Metadata).To(HaveKey(messageTopic))

						err = messenger.Acknowledge(msg)
						Expect(err).To(BeNil())
					case <-time.After(time.Second * 10):
						Fail("Message took too long to send and receive")
					}
				})
			})
			Context("When an invalid message was sent to Kafka", func() {
				It("Then the message should be skipped and nothing received", func() {
					producer := messenger.getProducer()
					producer.SendMessage(&sarama.ProducerMessage{
						Topic: kafkaInputTopic,
						Value: sarama.ByteEncoder("This message is not of type Message"),
					})
					select {
					case <-messenger.Receive():
						Fail("Should not have received malformed message")
					case <-time.After(time.Second * 2):
						return
					}
				})
			})
			Context("When a message is not acknowledge when received", func() {
				It("Then the message should be received again on reconnect", func() {
					err := messenger.Send(kafkaInputTopic, message)
					Expect(err).To(BeNil())

					select {
					case msg := <-messenger.Receive():
						Expect(err).To(BeNil())
						Expect(msg.Payload).To(Equal(message.Payload))
						Expect(msg.Metadata).To(HaveKey(messagePartition))
						Expect(msg.Metadata).To(HaveKey(messageOffset))
						Expect(msg.Metadata).To(HaveKey(messageTopic))
					case <-time.After(time.Second * 10):
						Fail("Message took too long to send and receive")
					}
					err = messenger.Close()
					Expect(err).To(BeNil())

					messenger, err = NewKafkaMessenger(kafkaBroker, config)
					Expect(err).To(BeNil())
					Expect(messenger).ToNot(BeNil())

					select {
					case msg := <-messenger.Receive():
						Expect(err).To(BeNil())
						Expect(msg.Payload).To(Equal(message.Payload))
						Expect(msg.Metadata).To(HaveKey(messagePartition))
						Expect(msg.Metadata).To(HaveKey(messageOffset))
						Expect(msg.Metadata).To(HaveKey(messageTopic))

						err = messenger.Acknowledge(msg)
						Expect(err).To(BeNil())
					case <-time.After(time.Second * 10):
						Fail("Message took too long to send and receive")
					}
				})
			})
		})
	})
})
