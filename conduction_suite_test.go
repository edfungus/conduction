package main_test

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
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
	kafkaBroker  string = "localhost:9092" // Override with KAFKA_URL if necessary
	databaseHost string = "localhost"      // Override with DATABASE_HOST if necessary
	databasePort int    = 26257            // Override with DATABASE_PORT if necessary
)

func TestConduction(t *testing.T) {
	if os.Getenv("KAFKA_URL") != "" {
		kafkaBroker = os.Getenv("KAFKA_URL")
	}
	if os.Getenv("DATABASE_HOST") != "" {
		databaseHost = os.Getenv("DATABASE_HOST")
	}
	if os.Getenv("DATABASE_PORT") != "" {
		var err error
		databasePort, err = strconv.Atoi(os.Getenv("DATABASE_PORT"))
		if err != nil {
			panic(err)
		}
	}

	ClearKafkaTopic(kafkaBroker, kafkaInputTopic, kafkaConsumerGroup)

	err := dropDatabase(databaseHost, databasePort, databaseName)
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

func dropDatabase(databaseHost string, databasePort int, databaseName string) error {
	databasePath := fmt.Sprintf(DATABASE_URL, "root", databaseHost, databasePort, databaseName)
	db, err := sql.Open("postgres", databasePath)
	if err != nil {
		return err
	}

	db.Exec(fmt.Sprintf("DROP DATABASE %s", databaseName))
	return nil
}
