package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/edfungus/conduction/distributor"
	"github.com/edfungus/conduction/distributor/kafka"
	"github.com/edfungus/conduction/model"
)

const (
	topic         = "KafkaDistributorTest"
	consumerGroup = "KafkaDistributorTest"
	broker        = "localhost:9092"
)

func main() {
	fmt.Println("Hello Conduction!")
	k, err := kafka.NewKafkaSarama(broker, topic, consumerGroup, kafka.DefaultKafkaSaramaConfigs())
	if err != nil {
		log.Println(fmt.Sprintf("Could not connec to Kafka. Is Kafka running on %s? Error: %s", broker, err.Error()))
	}
	kd := distributor.NewKafkaDistributor(k)
	defer kd.Close()

	firstMessage := &model.Message{
		Endpoint: "hello",
		Payload:  []byte(strconv.Itoa(32)),
	}

	kd.Send(firstMessage)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		for {
			select {
			case msg := <-kd.Messages():
				kd.Acknowledge(msg)
				number, err := strconv.Atoi(string(msg.Message.GetPayload()))
				if err != nil {
					return
				}
				fmt.Printf("Got label: %s and number: %d\n", msg.Message.GetEndpoint(), number)
				if number-1 > 0 {
					newMessage := &model.Message{
						Endpoint: msg.Message.GetEndpoint(),
						Payload:  []byte(strconv.Itoa(number - 1)),
					}
					kd.Send(newMessage)
				}
			case err := <-kd.Errors():
				fmt.Println("Got an error: ", err.Error())
			case <-signals:
				return
			}
		}
	}()
}
