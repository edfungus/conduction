package main

import (
	"fmt"

	"github.com/edfungus/conduction/connector"
)

type Conduction struct {
	connections       map[string]connector.Connector
	incomeRequests    chan *connector.Request
	mqttRelationships map[string][]string // used in place of database to store routing relationships for now
}

const (
	mqttBroker string = "tcp://localhost:1883/"
)

// NewConduction returns conduction
func NewConduction() *Conduction {
	return &Conduction{
		connections:       make(map[string]connector.Connector),
		incomeRequests:    make(chan *connector.Request),
		mqttRelationships: make(map[string][]string),
	}
}

// Wire takes content from one connection endpoint and sends the contents to another connection endpoint
func (c *Conduction) Wire(inConn string, inPath connector.Path, outConn string, outPath connector.Path) error {
	// Check if connection exists
	if _, ok := c.connections[inConn]; !ok {
		return fmt.Errorf("Input connection %s does not exist", inConn)
	}
	if _, ok := c.connections[outConn]; !ok {
		return fmt.Errorf("Output connection %s does not exist", outConn)
	}

	// If not listening to inPath, then listen (right now we always adds)
	err := c.connections[inConn].AddPath(inPath)
	if err != nil {
		return err
	}

	// Stores this relationship somewhere...
	if val, ok := c.mqttRelationships[inPath.PathName]; ok {
		c.mqttRelationships[inPath.PathName] = append(val, outPath.PathName)
	} else {
		// Not an existing input, so create it
		c.mqttRelationships[inPath.PathName] = []string{outPath.PathName}
	}

	return nil
}

// Close stops conduction and cleans up connections
func (c *Conduction) Close() {
	for _, v := range c.connections {
		v.Close()
	}
}

// AddMqttConnection starts a new MQTT connection which can be later referenced by a name
func (c *Conduction) AddMqttConnection(name string) (string, error) {
	mqttConfig := connector.MqttConnectorConfig{
		ClientID: "conduction",
	}
	mqttClient, err := connector.NewMqttConnector(mqttBroker, mqttConfig)
	if err != nil {
		return "", err
	}

	go func() {
		msg := <-mqttClient.Requests()
		c.incomeRequests <- msg
	}()

	return
}
