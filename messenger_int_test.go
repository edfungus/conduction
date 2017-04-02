// +build all integration

package main_test

import (
	"time"

	. "github.com/edfungus/conduction"

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
					err := messenger.Close()
					Expect(err).ToNot(BeNil())
				})
			})
		})
		Describe("Given Kafka is connected", func() {

			Context("When Messenger sends a message", func() {
				It("Then the message should send without an error", func() {
					// remmeber to check partition and offset!
				})
			})
			Context("When a valid message was sent to Kafka", func() {
				It("Then the message should be received", func() {

				})
			})
			Context("When an invalid message was sent to Kafka", func() {
				It("Then the message should be skipped and nothing received", func() {

				})
			})
			Context("When a message is not acknowledge when received", func() {
				It("Then the message should be received again on reconnect", func() {

				})
			})
		})
	})
})

func ClearKafkaTopic(broker string, topic string, consumerGroup string) {
	kd, _ := MakeKafkaSaramaDistributor(broker, topic, consumerGroup, kafka.DefaultKafkaSaramaConfigs())
	defer kd.Close()

	kd.Send(&model.Message{
		Endpoint: "Clearing messages",
	})

	stop := make(chan bool, 1)
	firstMessage := true
	for {
		select {
		case msg := <-kd.Messages():
			if firstMessage {
				firstMessage = false
				go func() {
					time.Sleep(time.Second * 2) // Should be long enough to clear any old messages
					stop <- true
				}()
			}
			kd.Acknowledge(msg)
		case <-stop:
			return
		case <-time.After(time.Second * 30):
			return
		}
	}
}
