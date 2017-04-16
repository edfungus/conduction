package main_test

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	. "github.com/edfungus/conduction"
	"github.com/edfungus/conduction/pb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const (
	kafkaConsumerGroup string = "conduction-test"
	kafkaInputTopic    string = "conductionIn-test"
	databaseName       string = "conductionTest"
)

var (
	kafkaBroker  string = "localhost:9092"  // Override with KAFKA_URL if necessary
	cockroachURL string = "localhost:26257" // Override with DATABASE_URL if necessary
)

func TestConduction(t *testing.T) {
	if os.Getenv("KAFKA_URL") != "" {
		kafkaBroker = os.Getenv("KAFKA_URL")
	}
	if os.Getenv("DATABASE_URL") != "" {
		cockroachURL = os.Getenv("DATABASE_URL")
	}

	ClearKafkaTopic(kafkaBroker, kafkaInputTopic, kafkaConsumerGroup)

	err := dropDatabase(cockroachURL, databaseName)
	if err != nil {
		panic(err)
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Conduction Suite")
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

func dropDatabase(cockroachURL string, databaseName string) error {
	db, err := sql.Open("postgres", fmt.Sprintf(DATABASE_URL, "root", cockroachURL, databaseName))
	if err != nil {
		return err
	}

	db.Exec(fmt.Sprintf("DROP DATABASE %s", databaseName))
	return nil
}
