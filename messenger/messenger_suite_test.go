package messenger

import (
	"fmt"
	"os"
	"time"

	"github.com/edfungus/conduction/pb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const (
	kafkaConsumerGroup string = "conduction-test"
	kafkaInputTopic    string = "conductionIn-test"
)

var (
	kafkaBroker string = "localhost:9092" // Override with KAFKA_URL if necessary
)

func TestMessenger(t *testing.T) {
	if os.Getenv("KAFKA_URL") != "" {
		kafkaBroker = os.Getenv("KAFKA_URL")
	}
	ClearKafkaTopic(kafkaBroker, kafkaInputTopic, kafkaConsumerGroup)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Messenger Suite")
}

func ClearKafkaTopic(broker string, topic string, consumerGroup string) {
	messenger, err := NewKafkaMessenger(broker, &KafkaMessengerConfig{
		ConsumerGroup: consumerGroup,
		InputTopic:    topic,
	})
	if err != nil {
		panic(fmt.Sprintf("Could not connec to Kafka. Is Kafka running on %s? Error: %s", broker, err.Error()))
	}
	defer messenger.Close()

	// Send dummy message
	err = messenger.Send(topic, &pb.Message{
		Payload: []byte("payload"),
	})
	if err != nil {
		panic(err)
	}

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
