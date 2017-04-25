package main

import (
	"github.com/edfungus/conduction/messenger"
	"github.com/sirupsen/logrus"
)

// Logger controls logging and levels
var Logger = logrus.New()

func main() {
	Logger.Info("Hello Conduction!")
	messenger.Logger = Logger
}

func setupLogger() {
	Logger.Level = logrus.WarnLevel
}
