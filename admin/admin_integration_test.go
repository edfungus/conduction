// +build all integration

package admin_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	. "github.com/edfungus/conduction/admin"
	"github.com/edfungus/conduction/messenger"
	"github.com/edfungus/conduction/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Admin", func() {
	Describe("Admin", func() {
		var (
			graph   *storage.GraphStorage
			manager *Admin
		)
		BeforeEach(func() {
			var err error
			os.Create(tempFilePath)
			graph, err = storage.NewGraphStorageBolt(tempFilePath)
			Expect(err).To(BeNil())
			Expect(graph).ToNot(BeNil())

			manager = NewAdmin(graph)
		})
		AfterEach(func() {
			graph.Close()
			os.Remove(tempFilePath)
		})
		Describe("Given inserting a new Flow", func() {
			Context("When the Flow is complete and correct", func() {
				It("Then the Flow will be inserted", func() {
					body :=
						`{
							"name": "Test Flow",
							"description": "Some description",
							"path": {
								"route": "/test",
								"type": "REST"
							}
						}`
					req, _ := http.NewRequest("POST", "/flows", bytes.NewBufferString(body))
					w := httptest.NewRecorder()
					manager.Router.ServeHTTP(w, req)

					Expect(w.Code).To(Equal(http.StatusCreated))
					Expect(w.Body.String()).To(ContainSubstring("uuid"))
				})
			})
			Context("When the Flow is missing Path", func() {
				It("Then an error will be returned", func() {
					body :=
						`{
							"name": "Test Flow",
							"description": "Some description"
						}`
					req, _ := http.NewRequest("POST", "/flows", bytes.NewBufferString(body))
					w := httptest.NewRecorder()
					manager.Router.ServeHTTP(w, req)

					Expect(w.Code).To(Equal(http.StatusBadRequest))
					Expect(w.Body.String()).To(ContainSubstring("Flow is missing field: path"))
				})
			})
			Context("When the Flow is missing description", func() {
				It("Then an error will be returned", func() {
					body :=
						`{
							"name": "Test Flow",
							"path": {
								"route": "/test",
								"type": "REST"
							}
						}`
					req, _ := http.NewRequest("POST", "/flows", bytes.NewBufferString(body))
					w := httptest.NewRecorder()
					manager.Router.ServeHTTP(w, req)

					Expect(w.Code).To(Equal(http.StatusBadRequest))
					Expect(w.Body.String()).To(ContainSubstring("Flow is missing field: description"))
				})
			})
		})
		Describe("Given retieving a new Flow", func() {
			Context("When the Flow uuid exist", func() {
				It("Then the Flow will be returned", func() {
					flow := storage.Flow{
						Name:        "Test flow",
						Description: "Test description",
						Path: &messenger.Path{
							Route: "Test route",
							Type:  "Test type",
						},
					}
					flowID, err := manager.Storage.SaveFlow(flow)
					Expect(err).To(BeNil())

					req, _ := http.NewRequest("GET", fmt.Sprintf("/flows/%s", flowID.String()), nil)
					w := httptest.NewRecorder()
					manager.Router.ServeHTTP(w, req)
					Expect(w.Code).To(Equal(http.StatusOK))

					var flowResponse storage.Flow
					err = json.Unmarshal(w.Body.Bytes(), &flowResponse)
					Expect(err).To(BeNil())
					Expect(flowResponse.Name).To(Equal(flow.Name))
					Expect(flowResponse.Description).To(Equal(flow.Description))
					Expect(flowResponse.Path.Route).To(Equal(flow.Path.Route))
					Expect(flowResponse.Path.Type).To(Equal(flow.Path.Type))
				})
			})
			Context("When the Flow uuid does not exist", func() {
				It("Then an error will be returned", func() {
					key := storage.NewRandomKey()
					req, _ := http.NewRequest("GET", fmt.Sprintf("/flows/%s", key.String()), nil)
					w := httptest.NewRecorder()
					manager.Router.ServeHTTP(w, req)

					Expect(w.Code).To(Equal(http.StatusNotFound))
					Expect(w.Body.String()).To(ContainSubstring("Could not retrieve Flow from storage"))
				})
			})
			Context("When the Flow uuid is not a valid uuid", func() {
				It("Then an error will be returned", func() {
					key := "xxx-xxx-xxx-xxx-xxx"
					req, _ := http.NewRequest("GET", fmt.Sprintf("/flows/%s", key), nil)
					w := httptest.NewRecorder()
					manager.Router.ServeHTTP(w, req)

					Expect(w.Code).To(Equal(http.StatusBadRequest))
					Expect(w.Body.String()).To(ContainSubstring("\"error\":\"uuid"))
				})
			})
		})
		Describe("Given inserting a new Path", func() {
			Context("When the Path is complete and correct", func() {
				It("Then the Path will be inserted", func() {
					body :=
						`{
							"route": "/test",
							"type": "REST"
						}`
					req, _ := http.NewRequest("POST", "/paths", bytes.NewBufferString(body))
					w := httptest.NewRecorder()
					manager.Router.ServeHTTP(w, req)

					Expect(w.Code).To(Equal(http.StatusCreated))
					Expect(w.Body.String()).To(ContainSubstring("uuid"))
				})
			})
			Context("When the Path is missing route", func() {
				It("Then an error wil be retutned", func() {
					body :=
						`{
							"type": "REST"
						}`
					req, _ := http.NewRequest("POST", "/paths", bytes.NewBufferString(body))
					w := httptest.NewRecorder()
					manager.Router.ServeHTTP(w, req)

					Expect(w.Code).To(Equal(http.StatusBadRequest))
					Expect(w.Body.String()).To(ContainSubstring("error"))
				})
			})
		})
		Describe("Given retieving a Path", func() {
			Context("When the Path uuid exist", func() {
				It("Then the Path will be returned", func() {
					path := messenger.Path{
						Route: "Test route",
						Type:  "Test type",
					}
					pathID, err := manager.Storage.SavePath(path)
					Expect(err).To(BeNil())

					req, _ := http.NewRequest("GET", fmt.Sprintf("/paths/%s", pathID.String()), nil)
					w := httptest.NewRecorder()
					manager.Router.ServeHTTP(w, req)
					Expect(w.Code).To(Equal(http.StatusOK))

					var pathResponse messenger.Path
					err = json.Unmarshal(w.Body.Bytes(), &pathResponse)
					Expect(err).To(BeNil())
					Expect(pathResponse.Route).To(Equal(path.Route))
					Expect(pathResponse.Type).To(Equal(path.Type))
				})
			})
			Context("When the Path uuid does not exist", func() {
				It("Then an error will be returned", func() {
					key := storage.NewRandomKey()
					req, _ := http.NewRequest("GET", fmt.Sprintf("/paths/%s", key.String()), nil)
					w := httptest.NewRecorder()
					manager.Router.ServeHTTP(w, req)

					Expect(w.Code).To(Equal(http.StatusNotFound))
					Expect(w.Body.String()).To(ContainSubstring("Could not retrieve Path from storage"))
				})
			})
			Context("When the Flow uuid is not a valid uuid", func() {
				It("Then an error will be returned", func() {
					key := "xxx-xxx-xxx-xxx-xxx"
					req, _ := http.NewRequest("GET", fmt.Sprintf("/paths/%s", key), nil)
					w := httptest.NewRecorder()
					manager.Router.ServeHTTP(w, req)

					Expect(w.Code).To(Equal(http.StatusBadRequest))
					Expect(w.Body.String()).To(ContainSubstring("\"error\":\"uuid"))
				})
			})
		})
	})
})
