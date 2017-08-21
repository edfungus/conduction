// +build all work

package router

import (
	"fmt"
	"time"

	"github.com/edfungus/conduction/messenger"
	"github.com/edfungus/conduction/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Conduction", func() {
	Describe("Router", func() {
		var (
			mockStorage   *mockStorage
			mockMessenger *mockMessenger
		)
		typeKeyREST := "REST"
		topicNames := map[string]string{
			typeKeyREST: "restTopic",
			"MQTT":      "mqttTopic",
		}
		config := RouterConfig{
			TopicNames: topicNames,
		}

		Describe("Given creating a new Router", func() {
			var router *Router
			ackCalled := make(chan bool)
			mockAcknowledge = func(messenger.Message) error {
				ackCalled <- true
				return nil
			}
			messageInChannel := make(chan messenger.Message)
			mockReceive = func() <-chan messenger.Message {
				return messageInChannel
			}

			Context("When a new Router is returned", func() {
				It("Then the Router should be returned not started", func() {
					go func() {
						message := messenger.Message{}
						messageInChannel <- message
					}()
					router = NewRouter(mockMessenger, mockStorage, config)
					select {
					case <-ackCalled:
						Fail("Message should not have been received")
					case <-time.After(time.Second * 1):
						return
					}
				})
			})
			Context("When Router is started", func() {
				It("Then the Router should start receiving messages", func() {
					router.Start()
					select {
					case <-ackCalled:
						return
					case <-time.After(time.Second * 1):
						Fail("Message should have been received")
					}
				})
			})
			Context("When Router is stopped", func() {
				It("Then the Router should stop receiving messages", func() {
					router.Stop()
					select {
					case <-ackCalled:
						Fail("Message should not have been received")
					case <-time.After(time.Second * 1):
						return
					}
				})
			})
			Context("Cleanup", func() {
				It("Cleanup mock functions", func() {
					mockAcknowledge = nil
					mockReceive = nil
				})
			})
		})
		Describe("Given Router has a map of type to topic name", func() {
			router := NewRouter(mockMessenger, mockStorage, config)

			Context("When finding topic name for existing type", func() {
				It("Then the corresponding topic name should be returned", func() {
					topic, err := router.getTopicForPathType(typeKeyREST)
					Expect(topic).To(Equal(topicNames[typeKeyREST]))
					Expect(err).To(BeNil())
				})
			})
			Context("When finding topic name for non-existing type", func() {
				It("Then an error should be thrown", func() {
					_, err := router.getTopicForPathType("typeThatDoesNotExist")
					Expect(err).ToNot(BeNil())
				})
			})
		})
		Describe("Given a message to be forwarded to a Path", func() {
			router := NewRouter(mockMessenger, mockStorage, config)
			messageToBeForwarded := messenger.Message{
				Origin: &messenger.Path{
					Route: "/test",
					Type:  "Origin type doesn't matter",
				},
				Payload: []byte("payload"),
			}

			Context("When the Path has a valid type", func() {
				It("Then a message should be sent to the right topic with the Path in destination property", func() {
					destinationToBeForwardedTo := messenger.Path{
						Route: "/pass",
						Type:  typeKeyREST,
					}
					mockSend = func(topic string, message messenger.Message) error {
						Expect(topic).To(Equal(topicNames[destinationToBeForwardedTo.Type]))
						Expect(*message.Destination).To(Equal(destinationToBeForwardedTo))
						return nil
					}

					err := router.forwardMessageToPath(messageToBeForwarded, destinationToBeForwardedTo)
					Expect(err).To(BeNil())
				})
			})
			Context("When the Path has a invalid type", func() {
				It("Then a message should be not be sent and an error returned", func() {
					badDestinationToBeForwardedTo := messenger.Path{
						Route: "/pass",
						Type:  "badType",
					}
					mockSend = func(topic string, message messenger.Message) error {
						Fail("Send should not have been called")
						return nil
					}

					err := router.forwardMessageToPath(messageToBeForwarded, badDestinationToBeForwardedTo)
					Expect(err).ToNot(BeNil())
				})
			})
			Context("When the Path is incomplete", func() {
				It("Then a message should be not be sent and an error returned", func() {
					incompleteDestinationToBeForwardedTo := messenger.Path{}
					mockSend = func(topic string, message messenger.Message) error {
						Fail("Send should not have been called")
						return nil
					}

					err := router.forwardMessageToPath(messageToBeForwarded, incompleteDestinationToBeForwardedTo)
					Expect(err).ToNot(BeNil())
				})
			})
			Context("Cleanup", func() {
				It("Cleanup mock functions", func() {
					mockSend = nil
				})
			})
		})
		Describe("Given a message is to be processed", func() {
			router := NewRouter(mockMessenger, mockStorage, config)
			messageToBeForwarded := messenger.Message{
				Origin: &messenger.Path{
					Route: "/test",
					Type:  "Origin type doesn't matter",
				},
				Payload: []byte("payload"),
			}
			Context("When the message has next Flows", func() {
				It("Then the message should be forwarded to the next Flows and original message acknowledged", func() {
					mockGetKeyOfPath = func(path messenger.Path) (storage.Key, error) {
						return storage.Key{}, nil
					}
					flow1 := storage.Flow{
						Path: &messenger.Path{
							Route: "/test1",
							Type:  typeKeyREST,
						},
					}
					flow2 := storage.Flow{
						Path: &messenger.Path{
							Route: "/test2",
							Type:  typeKeyREST,
						},
					}
					nextFlowArray := []storage.Flow{flow1, flow2}
					mockGetNextFlows = func(key storage.Key) ([]storage.Flow, error) {
						return nextFlowArray, nil
					}
					acknowledgeCalled := 0
					mockAcknowledge = func(message messenger.Message) error {
						acknowledgeCalled++
						return nil
					}
					mockSendCalled := 0
					mockSend = func(topic string, message messenger.Message) error {
						mockSendCalled++
						return nil
					}

					err := router.processMessage(messageToBeForwarded)
					Expect(err).To(BeNil())
					Expect(acknowledgeCalled).To(Equal(1))
					Expect(mockSendCalled).To(Equal(2))
				})
			})
			Context("When the message has no next Flows", func() {
				It("Then no messages should be forwarded and original message acknowledged", func() {
					mockGetKeyOfPath = func(path messenger.Path) (storage.Key, error) {
						return storage.Key{}, nil
					}
					nextFlowArray := []storage.Flow{}
					mockGetNextFlows = func(key storage.Key) ([]storage.Flow, error) {
						return nextFlowArray, nil
					}
					acknowledgeCalled := 0
					mockAcknowledge = func(message messenger.Message) error {
						acknowledgeCalled++
						return nil
					}
					mockSendCalled := 0
					mockSend = func(topic string, message messenger.Message) error {
						mockSendCalled++
						return nil
					}

					err := router.processMessage(messageToBeForwarded)
					Expect(err).To(BeNil())
					Expect(acknowledgeCalled).To(Equal(1))
					Expect(mockSendCalled).To(Equal(0))
				})
			})
			Context("Cleanup", func() {
				It("Cleanup mock functions", func() {
					mockSend = nil
				})
			})
		})
	})
})

type mockStorage struct{}

var mockSaveFlow func(flow storage.Flow) (storage.Key, error)
var mockGetFlowByKey func(key storage.Key) (storage.Flow, error)
var mockSavePath func(path messenger.Path) (storage.Key, error)
var mockGetPathByKey func(key storage.Key) (messenger.Path, error)
var mockGetKeyOfPath func(path messenger.Path) (storage.Key, error)
var mockChainNextFlowToPath func(flowKey storage.Key, pathKey storage.Key) error
var mockGetNextFlows func(key storage.Key) ([]storage.Flow, error)

func (ms *mockStorage) SaveFlow(flow storage.Flow) (storage.Key, error) {
	if mockSaveFlow == nil {
		fmt.Println("SaveFlow not implemented")
		return storage.Key{}, nil
	}
	return mockSaveFlow(flow)
}

func (ms *mockStorage) GetFlowByKey(key storage.Key) (storage.Flow, error) {
	if mockGetFlowByKey == nil {
		fmt.Println("GetFlowByKey not implemented")
		return storage.Flow{}, nil
	}
	return mockGetFlowByKey(key)
}

func (ms *mockStorage) SavePath(path messenger.Path) (storage.Key, error) {
	if mockSavePath == nil {
		fmt.Println("SavePath not implemented")
		return storage.Key{}, nil
	}
	return mockSavePath(path)
}

func (ms *mockStorage) GetPathByKey(key storage.Key) (messenger.Path, error) {
	if mockGetPathByKey == nil {
		fmt.Println("GetPathByKey not implemented")
		return messenger.Path{}, nil
	}
	return mockGetPathByKey(key)
}

func (ms *mockStorage) GetKeyOfPath(path messenger.Path) (storage.Key, error) {
	if mockGetKeyOfPath == nil {
		fmt.Println("GetKeyOfPath not implemented")
		return storage.Key{}, nil
	}
	return mockGetKeyOfPath(path)
}

func (ms *mockStorage) ChainNextFlowToPath(flowKey storage.Key, pathKey storage.Key) error {
	if mockChainNextFlowToPath == nil {
		fmt.Println("ChainNextFlowToPath not implemented")
		return nil
	}
	return mockChainNextFlowToPath(flowKey, pathKey)
}

func (ms *mockStorage) GetNextFlows(key storage.Key) ([]storage.Flow, error) {
	if mockGetNextFlows == nil {
		fmt.Println("GetNextFlows not implemented")
		return nil, nil
	}
	return mockGetNextFlows(key)
}

type mockMessenger struct{}

var mockSend func(topic string, message messenger.Message) error
var mockReceive func() <-chan messenger.Message
var mockAcknowledge func(messenger.Message) error
var mockClose func() error

func (mm *mockMessenger) Send(topic string, message messenger.Message) error {
	if mockSend == nil {
		fmt.Println("Send not implemented")
		return nil
	}
	return mockSend(topic, message)
}

func (mm *mockMessenger) Receive() <-chan messenger.Message {
	if mockReceive == nil {
		fmt.Println("Receive not implemented")
		return nil
	}
	return mockReceive()
}

func (mm *mockMessenger) Acknowledge(message messenger.Message) error {
	if mockAcknowledge == nil {
		fmt.Println("Acknowledge not implemented")
		return nil
	}
	return mockAcknowledge(message)
}

func (mm *mockMessenger) Close() error {
	if mockClose == nil {
		fmt.Println("Close not implemented")
		return nil
	}
	return mockClose()
}
