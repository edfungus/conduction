// +build all integration

package main_test

import (
	"database/sql"
	"fmt"

	. "github.com/edfungus/conduction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Conduction", func() {
	Describe("Storage", func() {
		Describe("Given creating a new Storage", func() {
			Context("When Cockroach is not available", func() {
				It("Then an error should occur", func() {
					_, err := NewCockroachStorage("postgresql://conductor@localhost:8888/badAddress", databaseName)
					Expect(err).ToNot(BeNil())
				})
			})
			Context("When Cockroach is available", func() {
				It("Then Storage should be returned", func() {
					db, err := NewCockroachStorage(cockroachURL, databaseName)
					Expect(err).To(BeNil())
					Expect(db).ToNot(BeNil())
				})
			})
		})
		Describe("Given checking if Path exists", func() {
			Context("When Path does exists in the database", func() {
				It("Then `true` should be returned", func() {
				})
			})
			Context("When Path does not exists in the database", func() {
				It("Then `false` should be returned", func() {
				})
			})
		})
		Describe("Given checking if Flow id exists", func() {
			var (
				cs     *CockroachStorage
				flowId int64 = 0
			)
			BeforeEach(func() {
				err := dropDatabase(cockroachURL, databaseName)
				Expect(err).To(BeNil())

				cs, err = NewCockroachStorage(cockroachURL, databaseName)
				Expect(err).To(BeNil())
				Expect(cs).ToNot(BeNil())

				// In future, replace with some insert flow function!
				pathId := 0
				err = cs.DB.QueryRow("INSERT INTO paths(route, type) VALUES($1, $2) RETURNING id", "/testRoute", 0).Scan(&pathId)
				Expect(err).To(BeNil())
				Expect(pathId).ToNot(Equal(0))

				err = cs.DB.QueryRow("INSERT INTO flows(\"path\", name, wait, listen) VALUES($1, $2, $3, $4) RETURNING id", pathId, "Test Flow", false, false).Scan(&flowId)
				Expect(err).To(BeNil())
				Expect(flowId).ToNot(Equal(0))
			})
			AfterEach(func() {
				cs.Close()
			})
			Context("When Flow id does exists in the database", func() {
				It("Then `true` should be returned", func() {
					ok, err := cs.FlowIDExist(flowId)
					Expect(err).To(BeNil())
					Expect(ok).To(Equal(true))
				})
			})
			Context("When Flow id does not exists in the database", func() {
				It("Then `false` should be returned", func() {
					ok, err := cs.FlowIDExist(flowId + 1)
					Expect(err).To(BeNil())
					Expect(ok).To(Equal(false))
				})
			})
		})
		Describe("Given adding a Flow without traversing deeper", func() {
			Describe("When the Flow id does not exist in database", func() {
				It("Then an error should be thrown and nothing should be added to the database", func() {

				})
			})
			Describe("When the Flow is missing id", func() {
				It("Then a new Flow should be made", func() {

				})
			})
			Describe("When the Flow is correct and dependents arg is nil", func() {
				It("Then the Flow should be updated without touching dependents and without an error", func() {
				})
			})
			Describe("When the dependent Flow id does not exist", func() {
				It("Then an error should be thrown", func() {

				})
			})
			Describe("When the Flow is correct and dependents arg is not nil", func() {
				It("Then the Flow should be updated with dependents and without an error", func() {
				})
			})
		})
		Describe("Given adding a new Flow recursively", func() {
			Context("When the Flow id does not exist", func() {
				It("Then it should return an error", func() {

				})
			})
			Context("When the Flow incorrectly has no Path", func() {
				It("Then it should return an error", func() {

				})
			})
			Context("When the Flow is correct, has no dependents and has new Path", func() {
				It("Then it should make the Path and Flow, add without error and return Flow id", func() {

				})
			})
			Context("When the Flow is correct, has one dependent Flow that was already created", func() {
				It("Then it should make the new Flow, add without error and return Flow id", func() {

				})
			})
			Context("When the Flow is correct, has one dependent Flow that has not been already created", func() {
				It("Then it should make both new Flows, add without error and return Flow id", func() {

				})
			})
			Context("When the Flow is correct, has multi-level dependent Flows", func() {
				It("Then it should make Flows as needed, add without error and return Flow id", func() {

				})
			})
		})
	})
})

func dropDatabase(cockroachURL string, databaseName string) error {
	db, err := sql.Open("postgres", fmt.Sprintf(DATABASE_URL, "root", cockroachURL, databaseName))
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", databaseName))
	if err != nil {
		return err
	}
	return nil
}
