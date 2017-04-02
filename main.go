package main

import (
	"github.com/sirupsen/logrus"
)

// Logger controls logging and levels
var Logger = logrus.New()

func main() {
	Logger.Info("Hello Conduction!")

}

func setupLogger() {
	Logger.Level = logrus.WarnLevel
}
