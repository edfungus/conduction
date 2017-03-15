package main

import (
	"os"
	"os/signal"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/edfungus/conduction/events"

	"fmt"
	"log"

	"time"

	"github.com/golang/protobuf/proto"
)

const (
	topic         = "testing"
	consumerGroup = "consumers"
)

var (
	broker = []string{"localhost:9092"}
)

func main() {
	fmt.Println("Hello Conduction!")
	listen()

	kafka := events.NewKafka()
}

func listen() {
	// Create Kafka producer
	configP := sarama.NewConfig()
	configP.Producer.RequiredAcks = sarama.WaitForAll
	configP.Producer.Retry.Max = 2
	configP.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(broker, configP)
	if err != nil {
		log.Fatalln("Error making Kafka producer: ", err)
	}
	defer producer.Close()

	// Create Kafka consumer
	configC := cluster.NewConfig()
	configC.Consumer.Return.Errors = true
	configC.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumer, err := cluster.NewConsumer(broker, consumerGroup, []string{topic}, configC)
	if err != nil {
		log.Fatalln("Error making Kafka consumer: ", err)
	}
	defer consumer.Close()

	test := &events.Kafka.Configs{

	}

	go func() {
		time.Sleep(time.Second * 3)
		message := &event.Message{
			Label:  "Hello!!",
			Number: 30,
		}
		send(message, producer)
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	done := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-consumer.Messages():
				consumer.MarkOffset(msg, "")
				message := &event.Message{}
				if err := proto.Unmarshal(msg.Value, message); err != nil {
					log.Fatalln("Could not parse message to transport.Message")
				}
				fmt.Printf("Got label: %s and number: %d\n", message.GetLabel(), message.GetNumber())
				message.Number = message.Number - 1
				if message.Number > 0 {
					send(message, producer)
				}
			case err := <-consumer.Errors():
				fmt.Println("Got an error: ", err.Error())
			case notif := <-consumer.Notifications():
				fmt.Println("Got a notification", notif.Current)
			case <-signals:
				done <- struct{}{}
			}
		}
	}()

	<-done

}

func send(message *event.Message, producer sarama.SyncProducer) {
	out, err := proto.Marshal(message)
	if err != nil {
		log.Fatalln("Could not marshal transport.Message")
	}
	producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(out),
	})
}
