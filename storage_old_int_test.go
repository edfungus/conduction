// +build old

package main_test

import (
	. "github.com/edfungus/conduction"
	"github.com/edfungus/conduction/pb"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Conduction", func() {
	Describe("Storage", func() {
		var cs *CockroachStorage
		BeforeEach(func() {
			var err error
			cs, err = NewCockroachStorage(cockroachURL, databaseName)
			Expect(err).To(BeNil())
			Expect(cs).ToNot(BeNil())
		})
		AfterEach(func() {
			cs.Close()
		})
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
					db.Close()
				})
			})
		})
		Describe("Given saving a Path", func() {
			Context("When Path does not exists in the database", func() {
				It("Then a new Path is insert and id is returned", func() {
					// Insert Path
					path := &pb.Path{
						Route: "/testRoute/pathTest/1",
						Type:  0,
					}
					id, err := cs.SavePath(path)
					Expect(err).To(BeNil())
					Expect(id).ToNot(BeNil())

					// Make sure it was actually inserted
					var (
						readId    int64
						readRoute string
						readType  pb.Path_ConnectorType
					)
					err = cs.DB.QueryRow("SELECT id, route, type FROM paths WHERE route=$1 AND type=$2", path.Route, path.Type).Scan(&readId, &readRoute, &readType)
					Expect(err).To(BeNil())
					Expect(readId).To(Equal(id))
					Expect(readRoute).To(Equal(path.Route))
					Expect(readType).To(Equal(path.Type))
				})
			})
			Context("When Path does exists in the database", func() {
				It("Then only the id is returned", func() {
					// Insert Path so it exists
					path := &pb.Path{
						Route: "/testRoute/pathTest/2",
						Type:  0,
					}
					id, err := cs.SavePath(path)
					Expect(err).To(BeNil())
					Expect(id).ToNot(BeNil())

					// Insert same Path again
					id, err = cs.SavePath(path)
					Expect(err).To(BeNil())
					Expect(id).ToNot(BeNil())

					// Make sure it was not insert more than once
					var (
						readId    int64
						readRoute string
						readType  pb.Path_ConnectorType
					)
					rows, err := cs.DB.Query("SELECT id, route, type FROM paths WHERE route=$1 AND type=$2", path.Route, path.Type)
					defer rows.Close()
					Expect(err).To(BeNil())
					Expect(rows.Next()).To(Equal(true))

					err = rows.Scan(&readId, &readRoute, &readType)
					Expect(err).To(BeNil())
					Expect(rows.Next()).To(Equal(false))
					Expect(readId).To(Equal(id))
					Expect(readRoute).To(Equal(path.Route))
					Expect(readType).To(Equal(path.Type))
				})
			})
		})
		Describe("Given checking if Flow id exists", func() {
			var (
				done   bool  = false
				flowId int64 = 0
			)
			BeforeEach(func() {
				if !done {
					// Inserting new Flow. In future, replace with some insert flow function!
					pathId := 0
					err := cs.DB.QueryRow("INSERT INTO paths(route, type) VALUES($1, $2) RETURNING id", "/testRoute/FlowIDExist", 0).Scan(&pathId)
					Expect(err).To(BeNil())
					Expect(pathId).ToNot(Equal(0))

					err = cs.DB.QueryRow("INSERT INTO flows(\"path\", name, wait, listen) VALUES($1, $2, $3, $4) RETURNING id", pathId, "Test Flow", false, false).Scan(&flowId)
					Expect(err).To(BeNil())
					Expect(flowId).ToNot(Equal(0))

					done = true
				}
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
		Describe("Given reading a single Flow", func() {
			var (
				done    bool = false
				flow1ID int64
				flow2ID int64
				flow3ID int64
			)

			const (
				flow1Name string = "Test Flow1"
				flow2Name string = "Test Flow2"
				route     string = "/testRoute/GetFlowSingle"
			)
			BeforeEach(func() {
				if !done {
					// Insert 3 flows where 1 Flow has 2 dependents (REPLACE when insert flow function is made)
					pathId := 0
					err := cs.DB.QueryRow("INSERT INTO paths(route, type) VALUES($1, $2) RETURNING id", route, 0).Scan(&pathId)
					Expect(err).To(BeNil())
					Expect(pathId).ToNot(Equal(0))

					err = cs.DB.QueryRow("INSERT INTO flows(\"path\", name, wait, listen) VALUES($1, $2, $3, $4) RETURNING id", pathId, flow1Name, false, false).Scan(&flow1ID)
					Expect(err).To(BeNil())
					Expect(flow1ID).ToNot(Equal(0))

					err = cs.DB.QueryRow("INSERT INTO flows(\"path\", name, wait, listen) VALUES($1, $2, $3, $4) RETURNING id", pathId, flow2Name, false, false).Scan(&flow2ID)
					Expect(err).To(BeNil())
					Expect(flow2ID).ToNot(Equal(0))

					err = cs.DB.QueryRow("INSERT INTO flows(\"path\", name, wait, listen) VALUES($1, $2, $3, $4) RETURNING id", pathId, "Test Flow3", false, false).Scan(&flow3ID)
					Expect(err).To(BeNil())
					Expect(flow3ID).ToNot(Equal(0))

					_, err = cs.DB.Exec("INSERT INTO flow_dependency(parent_path, parent_flow, dependent_path, dependent_flow, position) VALUES((SELECT \"path\" FROM flows WHERE id=$1), $1, (SELECT \"path\" FROM flows WHERE id=$2), $2, $3)", flow1ID, flow2ID, 0)
					Expect(err).To(BeNil())

					_, err = cs.DB.Exec("INSERT INTO flow_dependency(parent_path, parent_flow, dependent_path, dependent_flow, position) VALUES((SELECT \"path\" FROM flows WHERE id=$1), $1, (SELECT \"path\" FROM flows WHERE id=$2), $2, $3)", flow1ID, flow3ID, 1)
					Expect(err).To(BeNil())

					done = true
				}
			})
			Context("When the Flow does not exist", func() {
				It("Then an error should be thrown", func() {
					var expectedDenpendents []int64
					flow, dependents, err := cs.GetFlowSingle(flow1ID + 1)
					Expect(err).ToNot(BeNil())
					Expect(flow).To(BeNil())
					Expect(dependents).To(Equal(expectedDenpendents))
				})
			})
			Context("When the Flow has no dependents", func() {
				It("Then return the Flow with an empty dependent array", func() {
					var expectedDenpendents []int64
					flow, dependents, err := cs.GetFlowSingle(flow2ID)
					Expect(err).To(BeNil())
					Expect(flow).ToNot(BeNil())
					Expect(flow.Name).To(Equal(flow2Name))
					Expect(flow.Path.Route).To(Equal(route))
					Expect(dependents).To(Equal(expectedDenpendents))
				})
			})
			Context("When the Flow has multiple dependents", func() {
				It("Then return the Flow with a populated dependent array", func() {
					expectedDenpendents := []int64{flow2ID, flow3ID}
					flow, dependents, err := cs.GetFlowSingle(flow1ID)
					Expect(err).To(BeNil())
					Expect(flow).ToNot(BeNil())
					Expect(flow.Name).To(Equal(flow1Name))
					Expect(flow.Path.Route).To(Equal(route))
					Expect(dependents).To(Equal(expectedDenpendents))
				})
			})
		})
		Describe("Given adding a Flow without traversing deeper", func() {
			Context("When the Flow id does not exist in database", func() {
				It("Then an error should be thrown and nothing should be added to the database", func() {

				})
			})
			Context("When the Flow is missing id", func() {
				It("Then a new Flow should be made", func() {

				})
			})
			Context("When the Flow is correct and dependents arg is nil", func() {
				It("Then the Flow should be updated without touching dependents and without an error", func() {
				})
			})
			Context("When the dependent Flow id does not exist", func() {
				It("Then an error should be thrown", func() {

				})
			})
			Context("When the Flow is correct and dependents arg is not nil", func() {
				It("Then the Flow should be updated with dependents and without an error", func() {
				})
			})
			Context("When the Flow Path is change the old Path is no longer used", func() {
				It("Then the Flow should be updated and old Path removed", func() {
				})
			})
			Context("When the Flow Path is change the old Path is still used", func() {
				It("Then the Flow should be updated and old Path should not be removed", func() {
				})
			})
			Context("When the Flow dependent is removed and the dependent is not used and listen is false", func() {
				It("Then the Flow should be updated and old Flow removed", func() {
				})
			})
			Context("When the Flow dependent is removed and the dependent is used or listen is true", func() {
				It("Then the Flow should be updated and old Flow should not be removed", func() {
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
