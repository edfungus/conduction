package main

import (
	"testing"

	"github.com/Shopify/sarama/mocks"
)

// Only test so far
func TestMain(t *testing.T) {
	dataCollectorMock := mocks.NewSyncProducer(t, nil)
	dataCollectorMock.ExpectSendMessageAndSucceed()

}
