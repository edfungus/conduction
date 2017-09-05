package admin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	tempFilePath = "./test_db.tmp"
)

func TestAdmin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Admin Suite")
}
