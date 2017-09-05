package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/edfungus/conduction/storage"
	"github.com/gorilla/mux"
)

const (
	flowIDPathVariable = "flowID"
)

type Admin struct {
	Router  *mux.Router
	Storage storage.Storage
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewAdmin(storage storage.Storage) *Admin {
	r := mux.NewRouter()
	admin := &Admin{
		Router:  r,
		Storage: storage,
	}

	// r.HandleFunc("/flows", admin.getFlows).Methods("GET") // low priority
	r.HandleFunc("/flows", admin.postFlow).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/flows/{%s}", flowIDPathVariable), admin.getFlowByID).Methods("GET")

	// r.HandleFunc("/paths", getPaths).Methods("GET") // low priority
	// r.HandleFunc("/paths", postPath).Methods("POST")
	// r.HandleFunc("/paths/{uuid}", getPath).Methods("GET")
	// r.HandleFunc("/paths/{uuid}/flows", getFlowsFromPath).Methods("GET")
	// r.HandleFunc("/paths/{uuid}/flows/{uuid}", addFlowToPath).Methods("POST")
	// r.HandleFunc("/paths/{uuid}/flows/{uuid}", deleteFlowFromPath).Methods("DELETE")

	return admin
}

func (a *Admin) postFlow(w http.ResponseWriter, r *http.Request) {
	var flow storage.Flow
	err := getObjectFromRequestBody(r, &flow)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateFlow(flow); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}
	key, err := a.Storage.SaveFlow(flow)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response, err := json.Marshal(key)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, string(response), http.StatusCreated)
}

func (a *Admin) getFlowByID(w http.ResponseWriter, r *http.Request) {
	key, err := getKeyFromRequest(r)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}
	flow, err := a.Storage.GetFlowByKey(key)
	if err != nil {
		respondError(w, err.Error(), http.StatusNotFound)
		return
	}
	response, err := json.Marshal(flow)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, string(response), http.StatusOK)
}

func getKeyFromRequest(r *http.Request) (storage.Key, error) {
	vars := mux.Vars(r)
	id := vars[flowIDPathVariable]
	return storage.NewKeyFromString(id)
}

func getObjectFromRequestBody(r *http.Request, v interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	defer r.Body.Close()
	return json.Unmarshal(buf.Bytes(), v)
}

func respondJSON(w http.ResponseWriter, response string, code int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprint(w, response)
}

func respondError(w http.ResponseWriter, errString string, code int) {
	w.Header().Add("Content-Type", "application/json")
	error := errorResponse{
		Error: errString,
	}
	response, err := json.Marshal(error)
	if err != nil {
		respondError(w, "Error creating error json ... ", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	fmt.Fprintln(w, string(response))
}
