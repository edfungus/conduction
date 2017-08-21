package router

import (
	"errors"
	"fmt"

	"github.com/edfungus/conduction/messenger"
	"github.com/edfungus/conduction/storage"
	"github.com/sirupsen/logrus"
)

// Router handles all incoming messages and uses storage to reroute those messages
type Router struct {
	messenger  messenger.Messenger
	storage    storage.Storage
	topicNames TopicNames

	stop  chan bool
	start chan bool
}

type RouterConfig struct {
	TopicNames TopicNames
}

// TopicNames maps a routable type to a topic
type TopicNames map[string]string

// Logger logs but can be replaced
var Logger = logrus.New()

// NewRouter returns a new router that routes messages in/out of messenger based on storage
func NewRouter(messenger messenger.Messenger, storage storage.Storage, config RouterConfig) *Router {
	r := &Router{
		messenger:  messenger,
		storage:    storage,
		topicNames: config.TopicNames,
		stop:       make(chan bool),
		start:      make(chan bool),
	}
	go r.startRouting()
	return r
}

// Start begins the message routing
// Careful! Calling this multiple times will queue starts which can unexpectedly cancel out a stop calls
func (r *Router) Start() {
	go func() {
		r.start <- true
	}()
}

// Stop stops message routing
func (r *Router) Stop() {
	go func() {
		r.stop <- true
	}()
}

func (r *Router) startRouting() {
	<-r.start
	for {
		select {
		case <-r.stop:
			<-r.start
		case message := <-r.messenger.Receive():
			err := r.processMessage(message)
			Logger.Debugln(err)
		}
	}
}

func (r *Router) processMessage(message messenger.Message) (err error) {
	defer func() {
		switch {
		case err == nil:
			err = r.messenger.Acknowledge(message)
		default:
			r.messenger.Acknowledge(message)
		}
	}()

	nextFlows, err := r.getNextFlowsForMessage(message)
	if err != nil {
		return err
	}
	if len(nextFlows) == 0 {
		Logger.Debugln("No next Flow for path", message.Origin)
		return
	}
	err = r.forwardMessageToFlows(message, nextFlows)
	return err
}

func (r *Router) getNextFlowsForMessage(message messenger.Message) ([]storage.Flow, error) {
	if message.Origin == nil {
		return nil, errors.New("Message does not have an origin property")
	}
	pathKey, err := r.storage.GetKeyOfPath(*message.Origin)
	if err != nil {
		return nil, err
	}
	return r.storage.GetNextFlows(pathKey)
}

func (r *Router) forwardMessageToFlows(message messenger.Message, nextFlows []storage.Flow) error {
	for _, flow := range nextFlows {
		err := r.forwardMessageToPath(message, *flow.Path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Router) forwardMessageToPath(message messenger.Message, destinationPath messenger.Path) error {
	message = addPathAsDestination(message, destinationPath)
	topic, err := r.getTopicForPathType(destinationPath.Type)
	if err != nil {
		return err
	}
	err = r.messenger.Send(topic, message)
	return err
}

func addPathAsDestination(message messenger.Message, path messenger.Path) messenger.Message {
	message.Destination = &path
	return message
}

func (r *Router) getTopicForPathType(pathType string) (string, error) {
	topic, ok := r.topicNames[pathType]
	if !ok {
		return "", fmt.Errorf("Path type '%s' is an unknown type", pathType)
	}
	return topic, nil
}
