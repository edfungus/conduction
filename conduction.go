package main

import (
	"os"
	"os/signal"

	"github.com/edfungus/conduction/connector"
	"github.com/sirupsen/logrus"
)

type Conduction struct {
}

var Logger = logrus.New()

const (
	mqttBroker string = "tcp://localhost:1883/"
)

// NewConduction just making poc then expanding
func NewConduction() {
	mqttConfig := connector.MqttConnectorConfig{
		ClientID: "conduction",
	}
	mqttClient, err := connector.NewMqttConnector(mqttBroker, mqttConfig)
	if err != nil {
		Logger.Error(err)
	}

	mqttClient.Subscribe("inTopic")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	for {
		select {
		case msg := <-mqttClient.Requests():
			mqttClient.Respond("outTopic", msg.Payload)
		case <-signals:
			return
		}
	}
}

func setupLogger() {
	Logger.Level = logrus.WarnLevel
}
