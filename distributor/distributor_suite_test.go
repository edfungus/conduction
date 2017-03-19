package distributor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDistributor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Distributor Suite")
}
