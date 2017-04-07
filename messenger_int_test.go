// +build all integration

package main_test

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	. "github.com/edfungus/conduction"

	"github.com/edfungus/conduction/pb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	kafkaBroker        string = "localhost:9092"
	kafkaConsumerGroup string = "conduction-test"
	kafkaInputTopic    string = "conductionIn-test"
)

var _ = Describe("Conduction", func() {
	Describe("Messenger", func() {
		var (
			config = &KafkaMessengerConfig{
				ConsumerGroup: kafkaConsumerGroup,
				InputTopic:    kafkaInputTopic,
			}
		)
		BeforeSuite(func() {
			ClearKafkaTopic(kafkaBroker, kafkaInputTopic, kafkaConsumerGroup)
		})
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
				message   = &pb.Message{
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
						Expect(msg.Metadata).To(HaveKey(MESSAGE_PARTITION))
						Expect(msg.Metadata).To(HaveKey(MESSAGE_OFFSET))
						Expect(msg.Metadata).To(HaveKey(MESSAGE_TOPIC))

						err = messenger.Acknowledge(msg)
						Expect(err).To(BeNil())
					case <-time.After(time.Second * 10):
						Fail("Message took too long to send and receive")
					}
				})
			})
			Context("When an invalid message was sent to Kafka", func() {
				It("Then the message should be skipped and nothing received", func() {
					producer := messenger.GetProducer()
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
						Expect(msg.Metadata).To(HaveKey(MESSAGE_PARTITION))
						Expect(msg.Metadata).To(HaveKey(MESSAGE_OFFSET))
						Expect(msg.Metadata).To(HaveKey(MESSAGE_TOPIC))
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
						Expect(msg.Metadata).To(HaveKey(MESSAGE_PARTITION))
						Expect(msg.Metadata).To(HaveKey(MESSAGE_OFFSET))
						Expect(msg.Metadata).To(HaveKey(MESSAGE_TOPIC))

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

func ClearKafkaTopic(broker string, topic string, consumerGroup string) {
	messenger, err := NewKafkaMessenger(broker, &KafkaMessengerConfig{
		ConsumerGroup: consumerGroup,
		InputTopic:    topic,
	})
	if err != nil {
		Fail(fmt.Sprintf("Could not connec to Kafka. Is Kafka running on %s? Error: %s", broker, err.Error()))
	}
	defer messenger.Close()

	// Send dummy message
	err = messenger.Send(topic, &pb.Message{
		Payload: []byte("payload"),
	})
	Expect(err).To(BeNil())

	stop := make(chan bool, 1)
	firstMessage := true
	count := 0
	for {
		select {
		case msg := <-messenger.Receive():
			if firstMessage {
				firstMessage = false
				go func() {
					time.Sleep(time.Second * 2) // Should be long enough to clear any old messages
					stop <- true
				}()
			}
			messenger.Acknowledge(msg)
			count++
		case <-stop:
			fmt.Printf("Cleared %d message(s)\n", count)
			return
		case <-time.After(time.Second * 30):
			Fail("Could not even receive dummy message")
			return
		}
	}
}
