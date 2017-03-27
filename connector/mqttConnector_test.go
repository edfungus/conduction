package connector_test

import (
	"fmt"
	"time"

	"errors"

	. "github.com/edfungus/conduction/connector"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MqttConnector", func() {
	var (
		mc      *MqttConnector
		message []byte = []byte("test message")
	)
	const (
		topic    string = "testTopic"
		clientID string = "MqttTestClient2"
	)
	BeforeSuite(func() {
		ClearMqttTopic(clientID, topic)
	})
	BeforeEach(func() {
		mcConfig := MqttConnectorConfig{ClientID: clientID}
		err := errors.New("")
		mc, err = NewMqttConnector("tcp://localhost:1883/", mcConfig)
		Expect(err).To(BeNil())
	})
	AfterEach(func() {
		mc.Close()
	})
	Describe("Given a message is to be sent", func() {
		Context("When the connector send a valid message", func() {
			It("Should publish without error", func() {
				err := mc.Respond(topic, []byte("first message"))
				Expect(err).To(BeNil())
			})
		})
	})
	Describe("Given a message comes to a topic", func() {
		Context("When the topic is being subscribed to", func() {
			It("Should push a Request object to its given Requests() channel", func() {
				err := mc.Subscribe(topic)
				Expect(err).To(BeNil())
				err = mc.Respond(topic, message)
				Expect(err).To(BeNil())
				err = mc.Respond(topic, message) // This should really be the message from first test... but cleanSesssion seems broken right now
				Expect(err).To(BeNil())

				count := 2
				for {
					select {
					case req := <-mc.Requests():
						Expect(req).To(Equal(&Request{Path: topic, Payload: message}))
						count--
					case <-time.After(time.Second * 10):
						Fail(fmt.Sprintf("Mqtt message took too long to arrive. Missing %v message(s)", count))
					}
					if count <= 0 {
						break
					}
				}
			})
		})
		Context("When the topic has been unsubscribed to", func() {
			It("Should not receive a message", func() {
				err := mc.Unsubscribe(topic)
				Expect(err).To(BeNil())
				err = mc.Respond(topic, message)
				Expect(err).To(BeNil())

				select {
				case <-mc.Requests():
					Fail("Should have not received MQTT message")
				case <-time.After(time.Second * 5):
					break
				}
			})
		})
	})
})

func ClearMqttTopic(clientID string, topic string) {
	mcConfig := MqttConnectorConfig{ClientID: clientID}
	mc, err := NewMqttConnector("tcp://localhost:1883/", mcConfig)
	Expect(err).To(BeNil())
	defer mc.Close()

	mc.Subscribe(topic)
	mc.Respond(topic, []byte("some message"))
	defer mc.Unsubscribe(topic)

	stop := make(chan bool, 1)
	firstMessage := true
	for {
		select {
		case <-mc.Requests():
			if firstMessage {
				firstMessage = false
				go func() {
					time.Sleep(time.Second * 2) // Should be long enough to clear any old messages
					stop <- true
				}()
			}
		case <-stop:
			return
		case <-time.After(time.Second * 5):
			return
		}
	}
}
