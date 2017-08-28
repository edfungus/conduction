package storage

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const (
	databaseName string = "conductionTest"
)

var (
	databaseHost = "localhost" // Override with DATABASE_HOST if necessary
	databasePort = 26257       // Override with DATABASE_PORT if necessary
	tempFilePath = "./test_db.tmp"
)

func TestStorage(t *testing.T) {
	if os.Getenv("DATABASE_HOST") != "" {
		databaseHost = os.Getenv("DATABASE_HOST")
	}
	if os.Getenv("DATABASE_PORT") != "" {
		var err error
		databasePort, err = strconv.Atoi(os.Getenv("DATABASE_PORT"))
		if err != nil {
			panic(err)
		}
	}
	err := dropDatabase(databaseHost, databasePort, databaseName)
	if err != nil {
		panic(err)
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Suite")
}

func dropDatabase(databaseHost string, databasePort int, databaseName string) error {
	databasePath := fmt.Sprintf(databaseURL, "root", databaseHost, databasePort, databaseName)
	db, err := sql.Open("postgres", databasePath)
	if err != nil {
		return err
	}

	db.Exec(fmt.Sprintf("DROP DATABASE %s", databaseName))
	return nil
}
