// +build all work

package router

import (
	"fmt"
	"time"

	"github.com/edfungus/conduction/messenger"
	"github.com/edfungus/conduction/storage"

	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"
)

var _ = Describe("Conduction", func() {
	Describe("Router", func() {
		var (
			mockStorage   *mockStorage
			mockMessenger *mockMessenger
		)

		topicNames := map[string]string{
			"REST": "restTopic",
			"MQTT": "mqttTopic",
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
			Context("Cleanup", func() {
				It("Cleanup mock functions", func() {
					mockAcknowledge = nil
					mockReceive = nil
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

var mockSend func(topic string, msg messenger.Message) error
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
