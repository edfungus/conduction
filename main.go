package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func main() {
	fmt.Println("Hello Conduction!")
	setupLogger()
	c := NewConduction()
	defer c.Close()
}

func setupLogger() {
	Logger.Level = logrus.WarnLevel
}
