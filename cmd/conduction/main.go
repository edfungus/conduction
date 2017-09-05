package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/edfungus/conduction/admin"
	"github.com/edfungus/conduction/messenger"
	"github.com/edfungus/conduction/router"
	"github.com/edfungus/conduction/storage"
	"github.com/sirupsen/logrus"
)

// Logger controls logging and levels
var Logger = logrus.New()

func main() {
	Logger.Info("Hello Conduction! :)")

	messengerConfig := &messenger.KafkaMessengerConfig{
		ConsumerGroup:   "conduction",
		TopicsToConsume: []string{"KAFKA-topic"},
	}
	messenger, err := messenger.NewKafkaMessenger("localhost:9092", messengerConfig)
	if err != nil {
		Logger.Fatal("Could not make messenger")
	}

	storage, err := storage.NewGraphStorageBolt("./database.bolt")
	if err != nil {
		Logger.Fatal("Could not make storage")
	}

	topicNames := map[string]string{"REST": "REST-topic", "MQTT": "MQTT-topic"}
	routerConfig := router.RouterConfig{
		TopicNames: topicNames,
	}
	router := router.NewRouter(messenger, storage, routerConfig)
	go router.Start()

	admin := admin.NewAdmin(storage)
	go log.Fatal(http.ListenAndServe(":8080", admin.Router))

	signalChan := make(chan os.Signal, 1)
	exitReady := make(chan bool, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		router.Stop()
		storage.Close()
		messenger.Close()
		exitReady <- true
	}()
	<-exitReady
}
