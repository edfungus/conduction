package distributor_test

import (
	"fmt"
	"time"

	. "github.com/edfungus/conduction/distributor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	broker        string = "localhost:9092"
	topic         string = "KafkaDistributorTest"
	consumerGroup string = "KafkaDistributorTest"
)

var _ = Describe("KafkaDistributor", func() {
	BeforeSuite(func() {
		ClearTopic(broker, topic, consumerGroup)
	})
	Describe("When KafkaSarama is connected succesfully", func() {
		var kd = &KafkaDistributor{}
		var k = &KafkaSarama{}

		BeforeEach(func() {
			kd, k = MakeKafkaDistributor(broker, topic, consumerGroup, DefaultKafkaSaramaConfigs())
			Expect(kd).ToNot(BeNil())
		})
		AfterEach(func() {
			kd.Close()
		})
		Describe("Send & Receive", func() {
			Context("With valid Message", func() {
				It("Should send the message and receive it without error", func() {
					msg := &Message{
						Label:  "Test Send & Receive",
						Number: 42,
					}
					err := kd.Send(msg)
					Expect(err).To(BeNil())

					select {
					case newMsg := <-kd.Messages():
						kd.Acknowledge(newMsg)
						Expect(newMsg.Message).Should(Equal(msg))
					case err := <-kd.Errors():
						Fail(err.Error())
					case <-time.After(time.Second * 10):
						Fail("Test took too long")
					}
				})
			})
		})
		Describe("Receive", func() {
			Context("With invalid Message", func() {
				It("Should get an error and no messages", func() {
					k.Send([]byte("invalid message for transport"))
					select {
					case <-kd.Messages():
						Fail("Message received should not be a valid Message")
					case err := <-kd.Errors():
						Expect(err.Error()).To(ContainSubstring("could not parse message to transport.Message"))
					case <-time.After(time.Second * 10):
						Fail("Test took too long")
					}
				})
			})
			Context("Without acknowledging", func() {
				It("Should get the message again on reconnct", func() {
					msg := &Message{
						Label:  "Test Without acknowledging",
						Number: 42,
					}
					err := kd.Send(msg)
					Expect(err).To(BeNil())

					select {
					case newMsg := <-kd.Messages():
						Expect(newMsg.Message).Should(Equal(msg))
					case err := <-kd.Errors():
						Fail(err.Error())
					case <-time.After(time.Second * 30):
						Fail("Test took too long")
					}

					newkd, _ := MakeKafkaDistributor(broker, topic, consumerGroup, DefaultKafkaSaramaConfigs())

					select {
					case newMsg := <-newkd.Messages():
						kd.Acknowledge(newMsg)
						Expect(newMsg.Message).Should(Equal(msg))
					case err := <-newkd.Errors():
						Fail(err.Error())
					case <-time.After(time.Second * 10):
						Fail("Test took too long")
					}
				})
			})
		})
	})
})

func MakeKafkaDistributor(broker string, topic string, consumerGroup string, config *KafkaSaramaConfigs) (*KafkaDistributor, *KafkaSarama) {
	k, err := NewKafkaSarama(broker, topic, consumerGroup, config)
	if err != nil {
		Fail(fmt.Sprintf("Could not connec to Kafka. Is Kafka running on %s? Error: %s", broker, err.Error()))
	}
	kd := NewKafkaDistributor(k)
	return kd, k
}

func ClearTopic(broker string, topic string, consumerGroup string) {
	kd, _ := MakeKafkaDistributor(broker, topic, consumerGroup, DefaultKafkaSaramaConfigs())
	defer kd.Close()

	kd.Send(&Message{
		Label:  "Clearing messages",
		Number: 2,
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
