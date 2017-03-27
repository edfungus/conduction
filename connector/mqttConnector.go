package connector

import (
	"log"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type MqttConnector struct {
	client           MQTT.Client
	requests         chan *Request
	subscribeHandler MQTT.MessageHandler
}

type MqttConnectorConfig struct {
	ClientID string
}

func NewMqttConnector(broker string, config MqttConnectorConfig) (*MqttConnector, error) {
	mqttConfig := MQTT.NewClientOptions()
	mqttConfig.AddBroker(broker)
	mqttConfig.SetClientID(config.ClientID)
	mqttConfig.SetCleanSession(true) //false is not working correctly???
	mqttConfig.SetOnConnectHandler(MQTT.OnConnectHandler(func(client MQTT.Client) {
		log.Printf("MQTT client %s is connected", config.ClientID)
	}))
	mqttConfig.SetConnectionLostHandler(MQTT.ConnectionLostHandler(func(client MQTT.Client, err error) {
		log.Printf("MQTT client disconnected from %s %v", broker, err)
	}))

	requests := make(chan *Request)
	subscribeHandler := MQTT.MessageHandler(func(client MQTT.Client, message MQTT.Message) {
		request := &Request{
			Path:    message.Topic(),
			Payload: message.Payload(),
		}
		requests <- request // Go func?
	})

	client := MQTT.NewClient(mqttConfig)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &MqttConnector{
		client:           client,
		requests:         requests,
		subscribeHandler: subscribeHandler,
	}, nil
}

// Requests returns a channel with all incoming messages from MQTT
func (mc *MqttConnector) Requests() <-chan *Request {
	return mc.requests
}

// Respond sends mqtt messages where path is the topic
func (mc *MqttConnector) Respond(path string, payload []byte) error {
	token := mc.client.Publish(path, byte(1), false, payload)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (mc *MqttConnector) Close() error {
	mc.client.Disconnect(1000)
	return nil
}

// Subscribe listens to a queue and pushes it to the requetss channel
func (mc *MqttConnector) Subscribe(topic string) error {
	if token := mc.client.Subscribe(topic, byte(1), mc.subscribeHandler); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Unsubscribe stops listening to a topic
func (mc *MqttConnector) Unsubscribe(topic string) error {
	if token := mc.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
