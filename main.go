package main

import (
	"fmt"
	"log"

	"github.com/edfungus/conduction/distributor"
)

const (
	topic         = "KafkaDistributorTest"
	consumerGroup = "KafkaDistributorTest"
	broker        = "localhost:9092"
)

func main() {
	fmt.Println("Hello Conduction!")
	k, err := distributor.NewKafkaSarama(broker, topic, consumerGroup, distributor.DefaultKafkaSaramaConfigs())
	if err != nil {
		log.Println(fmt.Sprintf("Could not connec to Kafka. Is Kafka running on %s? Error: %s", broker, err.Error()))
	}
	kd := distributor.NewKafkaDistributor(k)

	firstMessage := &distributor.Message{
		Label:  "Hello!",
		Number: 32,
	}

	kd.Send(firstMessage)

	done := make(chan bool)
	go func() {
		for {
			select {
			case msg := <-kd.Messages():
				kd.Acknowledge(msg)
				fmt.Printf("Got label: %s and number: %d\n", msg.Message.GetLabel(), msg.Message.GetNumber())
				newMessage := &distributor.Message{
					Label:  msg.Message.GetLabel(),
					Number: msg.Message.GetNumber() - 1,
				}
				if newMessage.GetNumber() > 0 {
					kd.Send(newMessage)
				}
			case err := <-kd.Errors():
				fmt.Println("Got an error: ", err.Error())
			}
		}
	}()

	<-done
}
